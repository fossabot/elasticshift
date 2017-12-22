package vcs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"bytes"

	"encoding/base64"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"gitlab.com/conspico/elasticshift/api/types"
	core "gitlab.com/conspico/elasticshift/core/store"
	"gitlab.com/conspico/elasticshift/identity/oauth2/providers"
	"gitlab.com/conspico/elasticshift/identity/team"
	"gitlab.com/conspico/elasticshift/sysconf"
)

var (

	// VCS
	errNoProviderFound         = "No provider found for %s"
	errGetUpdatedFokenFailed   = "Failed to get updated token %s"
	errGettingRepositories     = "Failed to get repositories for %s"
	errVCSAccountAlreadyLinked = "VCS account already linked"
)

// expiryDelta determines how earlier a token should be considered
const expiryDelta = 10 * time.Second

// VCS account owner type
const (
	OwnerTypeUser = 1
	OwnerTypeOrg  = 2
)

// True or False
const (
	True  = 1
	False = 0
)

// Constants for performing encode decode
const (
	EQUAL        = "="
	DOUBLEEQUALS = "=="
	DOT0         = ".0"
	DOT1         = ".1"
	DOT2         = ".2"
)

// Common constants
const (
	SLASH     = "/"
	SEMICOLON = ";"
)

type vcsService struct {
	store        Store
	teamStore    team.Store
	sysconfStore sysconf.Store
	// repoDS       RepoDatastore
	vcsProviders providers.Providers
	logger       logrus.Logger
}

// VCSService ..
type VCSService interface {
	Authorize(w http.ResponseWriter, r *http.Request)
	Authorized(w http.ResponseWriter, r *http.Request)
	// GetVCS(teamID string) (types.VCS, error)
	// SyncVCS(teamID, userName, provider string) (bool, error)
}

// NewVCSService ..
func NewService(logger logrus.Logger, s core.Store, teamStore team.Store, sysconfStore sysconf.Store) VCSService {

	this := &vcsService{
		store:        NewStore(s),
		teamStore:    teamStore,
		sysconfStore: sysconfStore,
		logger:       logger,
		vcsProviders: providers.New(),
	}

	// initialize the providers
	this.initProviders()
	return this
}

// Initialize the registered providers
func (s vcsService) initProviders() {

	vcsConf, err := s.sysconfStore.GetVCSSysConf()
	if err != nil {
		panic(err)
	}

	for _, conf := range vcsConf {

		var prov providers.Provider
		switch conf.Name {
		case providers.GithubProviderName:
			prov = providers.GithubProvider(s.logger, conf.Key, conf.Secret, conf.CallbackURL, conf.HookURL)
		case providers.GitlabProviderName:
			prov = providers.GitlabProvider(s.logger, conf.Key, conf.Secret, conf.CallbackURL, conf.HookURL)
		case providers.BitbucketProviderName:
			prov = providers.BitbucketProvider(s.logger, conf.Key, conf.Secret, conf.CallbackURL, conf.HookURL)
		}

		if prov != nil {
			s.vcsProviders.Set(conf.Name, prov)
		}
	}
}

func (s vcsService) Authorize(w http.ResponseWriter, r *http.Request) {

	team := mux.Vars(r)["team"]
	exist, err := s.teamStore.CheckExists(team)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch the team: %s, error: %v", team, err), http.StatusBadRequest)
		return
	}

	if !exist {
		http.Error(w, fmt.Sprintf("Team '%s' doesn't exist, please provide the valid name", team), http.StatusBadRequest)
		return
	}

	provider := mux.Vars(r)["provider"]
	p, err := s.vcsProviders.Get(provider)

	if err != nil {
		http.Error(w, fmt.Sprintf("Getting provider %s failed: %v", provider, err), http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	buf.WriteString(team)
	buf.WriteString(SEMICOLON)
	buf.WriteString(SLASH)
	buf.WriteString(SLASH)
	buf.WriteString(r.Host)

	url := p.Authorize(s.encode(buf.String()))

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Authorized ..
// Invoked when authorization finished by oauth app
func (s vcsService) Authorized(w http.ResponseWriter, r *http.Request) {

	provider := mux.Vars(r)["provider"]
	p, err := s.vcsProviders.Get(provider)
	if err != nil {
		http.Error(w, fmt.Sprintf("Getting provider %s failed: %v", provider, err), http.StatusBadRequest)
	}

	id := r.FormValue("id")
	code := r.FormValue("code")
	u, err := p.Authorized(id, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Finalize the authorization failed : %v", err), http.StatusBadRequest)
	}

	unescID := s.decode(id)
	escID := strings.Split(unescID, SEMICOLON)

	// persist user
	team := escID[0]

	acc, err := s.teamStore.GetVCSByID(team, u.ID)
	if strings.EqualFold(acc.ID, u.ID) {

		// updvcs.UpdatedDt = time.Now()
		acc.AccessToken = u.AccessToken
		acc.AccessCode = u.AccessCode
		acc.RefreshToken = u.RefreshToken
		acc.OwnerType = u.OwnerType
		acc.TokenExpiry = u.TokenExpiry

		s.teamStore.UpdateVCS(team, acc)

		http.Error(w, errVCSAccountAlreadyLinked, http.StatusConflict)
	}

	// u.ID = utils.NewUUID()
	// u.CreatedDt = time.Now()
	// u.UpdatedDt = time.Now()

	err = s.teamStore.SaveVCS(team, &u)
	if err != nil {
		s.logger.Errorln("SAVE VCS: ", err)
	}

	if err == nil {

		// TODO sync the repo and setup hook reqeust for the repo
		// go p.CreateHook(u.AccessCode, u.Name, u.OwnerType)
	}

	url := escID[1] + "/sysconf/vcs"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// func (s vcsService) GetVCS(teamID string) (GetVCSResponse, error) {

// 	result, err := s.vcsDS.GetVCS(teamID)
// 	return GetVCSResponse{Result: result}, err
// }

// func (s vcsService) SyncVCS(teamID, userName, providerID string) (bool, error) {

// 	acc, err := s.vcsDS.GetByID(providerID)
// 	if err != nil {
// 		return false, fmt.Errorf("Get by VCS ID failed during sync : %v", err)
// 	}

// 	err = s.sync(acc, userName)
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }

// func (s vcsService) sync(acc VCS, userName string) error {

// 	// Get the token
// 	t, err := s.getToken(acc)
// 	if err != nil {
// 		return fmt.Errorf("Get token failed : ", err)
// 	}

// 	// fetch the existing repository
// 	p, err := s.getProvider(acc.Type)
// 	if err != nil {
// 		return fmt.Errorf(errNoProviderFound, err)
// 	}

// 	// repository received from provider
// 	repos, err := p.GetRepos(t, acc.Name, acc.OwnerType)
// 	if err != nil {
// 		return fmt.Errorf("Failed to get repos from provider %s : %v", p.Name(), err)
// 	}

// 	// Fetch the repositories from esh repo store
// 	lrpo, err := s.repoDS.GetReposByVCSID(acc.TeamID, acc.ID)
// 	if err != nil {
// 		return fmt.Errorf("Getting repos by vcs id failed : %v", err)
// 	}

// 	rpo := make(map[string]Repo)
// 	for _, l := range lrpo {
// 		rpo[l.RepoID] = l
// 	}

// 	// combine the result set
// 	for _, rp := range repos {

// 		r, exist := rpo[rp.RepoID]
// 		if exist {

// 			updrepo := Repo{}
// 			updated := false
// 			if r.Name != rp.Name {
// 				updrepo.Name = rp.Name
// 				updated = true
// 			}

// 			if r.Private != rp.Private {
// 				updrepo.Private = rp.Private
// 				updated = true
// 			}

// 			if r.Link != rp.Link {
// 				updrepo.Link = rp.Link
// 				updated = true
// 			}

// 			if r.Description != rp.Description {
// 				updrepo.Description = rp.Description
// 				updated = true
// 			}

// 			if r.Fork != rp.Fork {
// 				updrepo.Fork = rp.Fork
// 				updated = true
// 			}

// 			if r.DefaultBranch != rp.DefaultBranch {
// 				updrepo.DefaultBranch = rp.DefaultBranch
// 				updated = true
// 			}

// 			if r.Language != rp.Language {
// 				updrepo.Language = rp.Language
// 				updated = true
// 			}

// 			if updated {
// 				// perform update
// 				updrepo.UpdatedBy = userName
// 				s.repoDS.Update(r, updrepo)
// 			}
// 		} else {

// 			// perform insert
// 			rp.ID, _ = util.NewUUID()
// 			rp.CreatedDt = time.Now()
// 			rp.UpdatedDt = time.Now()
// 			rp.CreatedBy = userName
// 			rp.TeamID = acc.TeamID
// 			rp.VcsID = acc.ID
// 			s.repoDS.Save(&rp)
// 		}

// 		// removes from the map
// 		if exist {
// 			delete(rpo, r.RepoID)
// 		}
// 	}

// 	var ids []string
// 	// Now iterate thru deleted repositories.
// 	for _, rp := range rpo {
// 		ids = append(ids, rp.ID)
// 	}

// 	err = s.repoDS.DeleteIds(ids)
// 	if err != nil {
// 		return fmt.Errorf("Failed to delete the vcs that does not exist remotly : %v", err)
// 	}

// 	return nil
// }

// Gets the valid token
// Checks whether the token is expired.
// Expired token will get refreshed.
func (s vcsService) getToken(team string, a types.VCS) (string, error) {

	// Never expire type token
	if a.RefreshToken == "" {
		return a.AccessToken, nil
	}

	// Token that requires frequent refresh
	// check if the token is expired
	if !a.TokenExpiry.Add(-expiryDelta).Before(time.Now()) {
		return a.AccessToken, nil
	}

	p, err := s.vcsProviders.Get(a.Kind)
	if err != nil {
		return "", fmt.Errorf(errNoProviderFound, err)
	}

	// Refresh the token
	tok, err := p.RefreshToken(a.RefreshToken)

	a.AccessToken = tok.AccessToken
	a.TokenExpiry = tok.Expiry
	a.RefreshToken = tok.RefreshToken

	// persist the updated token information
	err = s.teamStore.UpdateVCS(team, a)

	if err != nil {
		return "", fmt.Errorf("Failed to update VCS after token refreshed.", err)
	}
	return tok.AccessToken, nil
}

// Gets the provider by type
func (s vcsService) getProvider(vcsType int) (providers.Provider, error) {

	var name string
	switch vcsType {
	case providers.GithubType:
		name = providers.GithubProviderName
	case providers.BitBucketType:
		name = providers.BitbucketProviderName
	case providers.GitlabType:
		name = providers.GitlabProviderName
	}

	return s.vcsProviders.Get(name)
}

func (s vcsService) encode(id string) string {

	eid := base64.URLEncoding.EncodeToString([]byte(id))
	if strings.Contains(eid, DOUBLEEQUALS) {
		eid = strings.TrimRight(eid, DOUBLEEQUALS) + DOT2
	} else if strings.Contains(eid, EQUAL) {
		eid = strings.TrimRight(eid, EQUAL) + DOT1
	} else {
		eid = eid + DOT0
	}
	return eid
}

func (s vcsService) decode(id string) string {

	if strings.Contains(id, DOT2) {
		id = strings.TrimRight(id, DOT2) + DOUBLEEQUALS
	} else if strings.Contains(id, DOT1) {
		id = strings.TrimRight(id, DOT1) + EQUAL
	} else {
		id = strings.TrimRight(id, DOT0)
	}
	did, _ := base64.URLEncoding.DecodeString(id)
	return string(did[:])
}
