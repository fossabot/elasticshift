/*
Copyright 2015 The Elasticshift Authors.
*/
package providers

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"

	"gitlab.com/conspico/elasticshift/api/types"
	"gitlab.com/conspico/elasticshift/core/dispatch"
	"golang.org/x/oauth2"
	gh "golang.org/x/oauth2/github"
)

// Github related properties
const (
	GithubProviderName   = "github"
	GithubBaseURL        = "https://api.github.com"
	GithubProfileURL     = GithubBaseURL + "/user"
	GithubGetUserRepoURL = GithubBaseURL + "/users/:user/repos"
	GithubGetOrgRepoURL  = GithubBaseURL + "/orgs/:org/repos"
	GithubCreateHookURL  = GithubBaseURL + "/repos/:owner/:repo/hooks"
)

// hook events that github should invoke eshift.
var hooks = []string{
	"commit_comment",
	"create",
	"delete",
	"fork",
	"member",
	"public",
	"pull_request",
	"push",
	"status",
	"team_add",
}

// Github ...
type Github struct {
	CallbackURL string
	HookURL     string
	Config      *oauth2.Config
	logger      logrus.Logger
}

// GithubUser ..
type githubUser struct {
	RawData     map[string]interface{}
	Type        int
	AccessToken string
	AvatarURL   string
}

// GithubProvider ...
// Creates a new Github provider
func GithubProvider(logger logrus.Logger, clientID, secret, callbackURL, hookURL string) *Github {

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		Scopes:       []string{"user,repo"},
		Endpoint:     gh.Endpoint,
	}

	return &Github{
		callbackURL,
		hookURL,
		conf,
		logger,
	}
}

// Name of the provider
func (g *Github) Name() string {
	return GithubProviderName
}

// Authorize ...
// Provide access to esh app on accessing the github user and repos.
// the elasticshift application to have access to github repo
func (g *Github) Authorize(baseURL string) string {
	// g.Config.RedirectURL = g.CallbackURL + "?id=" + baseURL
	g.logger.Warnln("Redirect URI :", g.CallbackURL+"?id="+baseURL)

	opts := oauth2.SetAuthURLParam("redirect_uri", g.CallbackURL+"?id="+baseURL)

	url := g.Config.AuthCodeURL("state", oauth2.AccessTypeOffline, opts)
	return url
}

// Authorized ...
// Finishes the authorize
func (g *Github) Authorized(id, code string) (types.VCS, error) {

	g.logger.Infoln("Callback code: ", code)
	g.logger.Warnln("github config: ", g)

	// var tok Token
	// tok := &Token{}
	tok, err := g.Config.Exchange(oauth2.NoContext, code)

	fmt.Println(tok)

	u := types.VCS{}
	if err != nil {
		return u, fmt.Errorf("Exchange token after authorization failed: ", err)
	}

	u.AccessCode = code
	u.RefreshToken = tok.RefreshToken
	u.AccessToken = tok.AccessToken
	if !tok.Expiry.IsZero() { // zero never expires
		u.TokenExpiry = tok.Expiry
	}

	u.Kind = GithubType

	us := struct {
		VcsID   int    `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Login   string `json:"login"`
		Picture string `json:"avatar_url"`
		Type    string
	}{}

	r := dispatch.NewGetRequestMaker(GithubProfileURL)
	r.SetLogger(g.logger)

	r.Header("Accept", "application/json")
	r.QueryParam("access_token", tok.AccessToken)
	err = r.Scan(&us).Dispatch()
	if err != nil {
		return u, err
	}

	g.logger.Warnln("Callback response: ", us)

	u.AvatarURL = us.Picture
	u.Name = us.Login
	if "User" == us.Type {
		u.OwnerType = OwnerTypeUser
	} else {
		u.OwnerType = OwnerTypeOrg
	}
	u.ID = strconv.Itoa(us.VcsID)
	return u, err
}

// RefreshToken ..
func (g *Github) RefreshToken(token string) (*oauth2.Token, error) {

	r := dispatch.NewGetRequestMaker(g.Config.Endpoint.TokenURL)
	r.SetLogger(g.logger)

	r.Header("Accept", "application/json")
	r.Header("Content-Type", "application/x-www-form-urlencoded")

	r.QueryParam("client_id", g.Config.ClientID)
	r.QueryParam("client_secret", g.Config.ClientSecret)
	r.QueryParam("grant_type", "refresh_token")
	r.QueryParam("refresh_token", token)

	var tok oauth2.Token
	err := r.Scan(&tok).Dispatch()

	if err != nil {
		return nil, err
	}
	return &tok, nil
}

// GetRepos ..
// returns the list of repositories
func (g *Github) GetRepos(token, accountName string, ownerType string) ([]types.Repository, error) {

	var url string
	if OwnerTypeUser == ownerType {
		url = GithubGetUserRepoURL
	} else if OwnerTypeOrg == ownerType {
		url = GithubGetUserRepoURL
	}

	r := dispatch.NewGetRequestMaker(url)
	r.SetLogger(g.logger)

	r.Header("Accept", dispatch.JSON)
	r.Header("Content-Type", dispatch.URLENCODED)

	r.PathParams(accountName)

	r.QueryParam("access_token", token)

	result := []struct {
		RepoID        int `json:"id"`
		Name          string
		Private       bool
		Link          string `json:"html_url"`
		Description   string
		Fork          bool
		DefaultBranch string `json:"default_branch"`
		Language      string
	}{}

	err := r.Scan(&result).Dispatch()
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	var repos []types.Repository
	for _, repo := range result {

		rp := &types.Repository{
			RepoID:        strconv.Itoa(repo.RepoID),
			Name:          repo.Name,
			Link:          repo.Link,
			Description:   repo.Description,
			DefaultBranch: repo.DefaultBranch,
			Language:      repo.Language,
		}

		if repo.Private {
			rp.Private = true
		}

		if repo.Fork {
			rp.Fork = true
		}
		repos = append(repos, *rp)
	}
	return repos, err
}

// CreateHook ..
// Create a new hook
func (g *Github) CreateHook(token, owner, repo string) error {

	r := dispatch.NewPostRequestMaker(GithubCreateHookURL)
	r.SetLogger(g.logger)

	r.SetContentType(dispatch.JSON)
	r.PathParams(owner, repo)

	r.QueryParam("access_token", token)

	body := struct {
		Name   string   `json:"name"`
		Active bool     `json:"active"`
		Events []string `json:"events"`
		Config struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
		} `json:"config"`
	}{}

	body.Name = "web"
	body.Active = true
	body.Events = hooks
	body.Config.URL = g.HookURL
	body.Config.ContentType = "JSON"

	r.Body(body)

	err := r.Dispatch()
	return err
}
