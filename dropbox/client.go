package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var baseURL = url.URL{
	Scheme: "https",
	Host:   "api.dropboxapi.com",
	Path:   "/2/",
}

type Client struct {
	BaseURL     *url.URL
	accessToken string
	httpClient  *http.Client
}

func NewClient(accessToken string) *Client {
	return &Client{
		BaseURL:     &baseURL,
		accessToken: accessToken,
		httpClient:  &http.Client{},
	}
}

func (c *Client) sendRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter

	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

type ListFileMembersRequestBody struct {
	File             string `json:"file"`
	IncludeInherited bool   `json:"include_inherited"`
	Limit            int32  `json:"limit"`
}

type User struct {
	AccessType struct {
		Tag string `json:".tag"`
	} `json:"access_type"`
	User struct {
		AccountID    string `json:"account_id"`
		Email        string `json:"email"`
		DisplayName  string `json:"display_name"`
		SameTeam     bool   `json:"same_team"`
		TeamMemberID string `json:"team_member_id"`
	} `json:"user"`
	Permissions  []interface{} `json:"permissions"`
	IsInherited  bool          `json:"is_inherited"`
	TimeLastSeen time.Time     `json:"time_last_seen"`
	PlatformType struct {
		Tag string `json:".tag"`
	} `json:"platform_type"`
}

type Group struct {
	AccessType struct {
		Tag string `json:".tag"`
	} `json:"access_type"`
	Group struct {
		GroupName           string `json:"group_name"`
		GroupID             string `json:"group_id"`
		GroupManagementType struct {
			Tag string `json:".tag"`
		} `json:"group_management_type"`
		GroupType struct {
			Tag string `json:".tag"`
		} `json:"group_type"`
		IsMember    bool `json:"is_member"`
		IsOwner     bool `json:"is_owner"`
		SameTeam    bool `json:"same_team"`
		MemberCount int  `json:"member_count"`
	} `json:"group"`
	Permissions []interface{} `json:"permissions"`
	IsInherited bool          `json:"is_inherited"`
}

type Invitee struct {
	AccessType struct {
		Tag string `json:".tag"`
	} `json:"access_type"`
	Invitee struct {
		Tag   string `json:".tag"`
		Email string `json:"email"`
	} `json:"invitee"`
	Permissions []interface{} `json:"permissions"`
	IsInherited bool          `json:"is_inherited"`
}

type ListFileMembersResponseBody struct {
	Users    []User    `json:"users"`
	Groups   []Group   `json:"groups"`
	Invitees []Invitee `json:"invitees"`
	Cursor   *string   `json:"cursor,omitempty"`
}

type ListMembersContinueRequestBody struct {
	Cursor string `json:"cursor"`
}

type FileMembers struct {
	Users    []User
	Groups   []Group
	Invitees []Invitee
}

// ListFileMembers returns file members
func (c *Client) ListFileMembers(fileID string, includeInherited bool, limit int32) (*FileMembers, error) {
	requestBody := ListFileMembersRequestBody{
		File:             fileID,
		IncludeInherited: includeInherited,
		Limit:            limit,
	}
	req, err := c.sendRequest("POST", "sharing/list_file_members", requestBody)
	if err != nil {
		return nil, err
	}

	var resp ListFileMembersResponseBody
	var fileMembers FileMembers

	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	fileMembers.Users = resp.Users
	fileMembers.Groups = resp.Groups
	fileMembers.Invitees = resp.Invitees

	var cursorPointer = resp.Cursor

	for cursorPointer != nil {
		req, err = c.sendRequest("POST", "sharing/list_file_members/continue", ListMembersContinueRequestBody{
			Cursor: *cursorPointer,
		})
		if err != nil {
			return nil, err
		}

		var resp ListFileMembersResponseBody
		_, err = c.do(req, &resp)
		if err != nil {
			return nil, err
		}

		fileMembers.Users = append(fileMembers.Users, resp.Users...)
		fileMembers.Groups = append(fileMembers.Groups, resp.Groups...)
		fileMembers.Invitees = append(fileMembers.Invitees, resp.Invitees...)

		cursorPointer = resp.Cursor
	}

	return &fileMembers, err
}
