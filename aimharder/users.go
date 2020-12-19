package aimharder

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

//User represents a user in Aimharder platform
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Sex         string `json:"sex"`
	UserFNac    string `json:"userFNac"`
	Desc        string `json:"desc"`
	DescNoParse string `json:"descNoParse"`
	Pic         string `json:"pic"`
	Box         string `json:"box"`
	BoxShort    string `json:"boxShort"`
	BoxID       string `json:"boxID"`
	Followers   string `json:"followers"`
	Following   string `json:"following"`
	Routines    string `json:"routines"`
	UserMail    string `json:"userMail"`
	UserLangUse string `json:"userLangUse"`
}

// UsersService handles communication with the users methods
type UsersService service

// Me returns logged user info
func (s *UsersService) Me(ctx context.Context) (*User, error) {
	userID := strings.Split(s.client.authToken, "|")[0]
	if userID == "" {
		return nil, MissingUserIDAuthToken
	}

	return s.Get(ctx, userID)
}

// Get returns user given its id
func (s *UsersService) Get(ctx context.Context, userID string) (*User, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, fmt.Sprintf("user/%s", userID), nil)
	if err != nil {
		return nil, err
	}

	var u User
	bodyResp := apiBodyDecoder{
		success: &u,
	}

	if _, err := s.client.do(ctx, req, &bodyResp); err != nil {
		return nil, err
	}

	if err := bodyResp.Error(); err != nil {
		return nil, err
	}

	if u.ID == "" {
		return nil, UserNotFound
	}

	return &u, nil
}
