package esh

import (
	"context"
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var (
	// Team
	errDomainNameIsEmpty         = errors.New("Team name is empty")
	errDomainNameMinLength       = errors.New("Team name should should be minimum of 6 chars")
	errDomainNameMaxLength       = errors.New("Team name should not exceed 63 chars")
	errDomainNameContainsSymbols = errors.New("Team name should be alpha-numeric, no special chars or whitespace is allowed")
	errTeamAlreadyExists         = errors.New("Team name already exists")

	// User
	errUserAlreadyExists         = errors.New("User already exists")
	errUserCreationFailed        = errors.New("User creation failed ")
	errNoTeamIDNotExist          = errors.New("Team ID does not exist")
	errNOTeamIDProvided          = errors.New("No team id provided")
	errVerificationCodeIsEmpty   = errors.New("Verification code is empty")
	errVerificationCodeExpired   = errors.New("Verification code expired")
	errVerificationCodeFailed    = errors.New("Verification code seems to be failed")
	errUsernameOrPasswordIsEmpty = errors.New("User or password is empty ")
	errInvalidEmailOrPassword    = errors.New("Invalid email or password")

	// VCS
	errNoProviderFound         = "No provider found for %s"
	errGetUpdatedFokenFailed   = "Failed to get updated token %s"
	errGettingRepositories     = "Failed to get repositories for %s"
	errVCSAccountAlreadyLinked = errors.New("VCS account already linked")
)

// Common constants
const (
	SLASH     = "/"
	SEMICOLON = ";"
)

// Config ..
type Config struct {
	Github struct {
		Key      string
		Secret   string
		Callback string
	}

	Gitlab struct {
		Key      string
		Secret   string
		Callback string
	}

	Bitbucket struct {
		Key      string
		Secret   string
		Callback string
	}

	DB struct {
		Dialect        string
		Datasource     string
		IdleConnection int
		MaxConnection  int
		Log            bool
		Retry          int
		Reconnect      int
	}
	Key struct {
		VerifyCode string
		Signer     string
		Verifier   string
	}
	CSRF struct {
		Key    string
		Secure bool
	}
}

// AppContext ..
type AppContext struct {
	Context  context.Context
	Router   *mux.Router
	Logger   *logrus.Logger
	Signer   interface{}
	Verifier interface{}
	Config   Config

	SecureChain alice.Chain
	PublicChain alice.Chain

	TeamService TeamService
	UserService UserService
	VCSService  VCSService
	RepoService RepoService

	Datasource    Datastore
	TeamDatastore TeamDatastore
	UserDatastore UserDatastore
	VCSDatastore  VCSDatastore
	RepoDatastore RepoDatastore
}

// Team ..
type Team struct {
	ID        string
	Domain    string
	Name      string
	CreatedDt time.Time
	CreatedBy string
	UpdatedDt time.Time
	UpdatedBy string
}

// User ..
type User struct {
	ID         string
	TeamID     string
	Fullname   string
	Username   string
	Email      string
	Password   string
	Locked     int8
	Active     int8
	BadAttempt int8
	CreatedDt  time.Time
	CreatedBy  string
	UpdatedDt  time.Time
	UpdatedBy  string
}

// VCS contains the information common amongst most OAuth and OAuth2 providers.
// All of the "raw" datafrom the provider can be found in the `RawData` field.
type VCS struct {
	ID           string `json:"ID"`
	TeamID       string `json:"-"`
	VcsID        string `json:"_"`
	Name         string
	Type         int
	OwnerType    int
	AvatarURL    string    `json:"avatar"`
	AccessCode   string    `json:"-"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenType    string    `json:"-"`
	TokenExpiry  time.Time `json:"-"`
	CreatedDt    time.Time `json:"-"`
	CreatedBy    string    `json:"-"`
	UpdatedDt    time.Time `json:"lastUpdated"`
	UpdatedBy    string    `json:"-"`
}

// Repo ..
// Represents as vcs repositories or projects
type Repo struct {
	ID            string `json:"ID"`
	TeamID        string `json:"-"`
	VcsID         string `json:"-"`
	RepoID        string
	Name          string
	Private       int
	Link          string
	Description   string
	Fork          int
	DefaultBranch string
	Language      string
	CreatedDt     time.Time `json:"-"`
	CreatedBy     string    `json:"-"`
	UpdatedDt     time.Time `json:"lastUpdated"`
	UpdatedBy     string    `json:"-"`
}
