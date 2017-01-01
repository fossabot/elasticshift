// Package esh ...
// Author: Ghazni Nattarshah
// Date: Dec 30, 2016
package esh

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

// Token ..
type Token struct {
	AccessToken string `json:"access_token"`

	// (bearer, mac, etc)
	TokenType string `json:"token_type"`

	// The refresh token, which can be used to obtain new
	// access tokens using the same authorization grant
	RefreshToken string `json:"refresh_token"`

	// The lifetime in seconds of the access token.
	ExpiresIn int64 `json:"expires_in"`

	Expiry time.Time `json:"expiry,omitempty"`

	CreatedAt int64 `json:"created_at"`
	Scope     string
}

// Provider ..
type Provider interface {
	Name() string

	Authorize(baseURL string) string

	GetRedirectURL(id string) string

	Authorized(code, redirectURL string) (VCS, error)

	RefreshToken(token string) (*oauth2.Token, error)

	GetRepos(token, accountName string, owner string) ([]Repo, error)
}

// Providers type
type Providers struct {
	Providers map[string]Provider
}

// NewProviders ...
func NewProviders(pvider ...Provider) *Providers {

	var prov = make(map[string]Provider)
	for _, p := range pvider {
		prov[p.Name()] = p
	}
	return &Providers{prov}
}

// Get the provider by namee
func (prov Providers) Get(name string) (Provider, error) {

	p := prov.Providers[name]
	if p == nil {
		return nil, fmt.Errorf(errNoProviderFound, name)
	}
	return p, nil
}
