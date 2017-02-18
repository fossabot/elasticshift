// Package esh ...
// Author: Ghazni Nattarshah
// Date: Sep 20, 2016
package esh

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/palantir/stacktrace"
)

// user registration
type signupRequest struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Team     string `json:"team"`
}

type signInRequest struct {
	Team     string `json:"team"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signOut struct {
	Request *http.Request
}

func decodeSignUpRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	var req signupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return false, stacktrace.Propagate(err, "Unable to decode signup request")
	}

	if req.Team == "" {
		req.Team = ctx.Value("subdomain").(string)
	}

	// validate email
	// validate fullname
	// validate password
	return req, nil
}

func decodeSignInRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	var req signInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return false, stacktrace.Propagate(err, "Unable to decode signin request")
	}

	if req.Team == "" {
		req.Team = ctx.Value("subdomain").(string)
	}

	// validate username and password
	if req.Email == "" || req.Password == "" {
		return false, errUsernameOrPasswordIsEmpty
	}

	return req, nil
}

// verify code
type verifyCodeRequest struct {
	Code string
}

func decodeVerifyCodeRequest(ctx context.Context, r *http.Request) (interface{}, error) {

	//params := ctx.Value("params").(map[string]string)
	//code := params["code"]
	//if len(code) == 0 {
	//	return false, errVerificationCodeIsEmpty
	//}
	//return verifyCodeRequest{Code: code}, nil
	return verifyCodeRequest{}, nil
}

func decodeSignOutRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return signOut{Request: r}, nil
}
