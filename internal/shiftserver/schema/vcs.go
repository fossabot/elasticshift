/*
Copyright 2017 The Elasticshift Authors.
*/
package schema

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/elasticshift/elasticshift/internal/pkg/logger"
	"github.com/elasticshift/elasticshift/internal/pkg/utils"
	"github.com/elasticshift/elasticshift/internal/shiftserver/identity/oauth2/providers"
	"github.com/elasticshift/elasticshift/internal/shiftserver/store"
	"github.com/elasticshift/elasticshift/internal/shiftserver/vcs"
)

func newVcsSchema(
	ctx context.Context,
	loggr logger.Loggr,
	providers providers.Providers,
	s store.Shift,
) (queries graphql.Fields, mutations graphql.Fields) {

	r, _ := vcs.NewResolver(ctx, loggr, s, providers)

	fields := graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "Represents the version control system ID",
		},

		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "Name of the version control system, it shall be organization or user",
		},

		"kind": &graphql.Field{
			Type:        graphql.String,
			Description: "Represents the repository type such as github, gitlab, bitbucket etc",
		},

		"owner_type": &graphql.Field{
			Type:        graphql.String,
			Description: "Represent the repository type sych as user or organization",
		},

		"avatar": &graphql.Field{
			Type:        graphql.String,
			Description: "An url that point the account profile picture",
		},

		"access_token": &graphql.Field{
			Type:        graphql.String,
			Description: "An access token that can be used to access this repository",
		},

		"refresh_token": &graphql.Field{
			Type:        graphql.String,
			Description: "The refresh token used to refresh the access token",
		},

		"token_expiry": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "Time when the token will be expired",
		},
	}

	vcsType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:        "VCS",
			Fields:      fields,
			Description: "An object of vcs type",
		},
	)

	args := graphql.FieldConfigArgument{
		"team": &graphql.ArgumentConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Represent the team name or ID",
		},
	}

	queries = graphql.Fields{
		"vcs": utils.MakeListType("VCSList", vcsType, r.FetchVCS, args),
	}

	mutations = graphql.Fields{}

	return queries, mutations
}
