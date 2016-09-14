package team

import (
	"time"

	"gitlab.com/conspico/esh/core/util"
)

// Service ..
type Service interface {
	Create(name string) (bool, error)

	IsAvailable(name string) (bool, error)
}

type service struct {
	teamRepository Repository
}

func (t service) Create(name string) (bool, error) {

	id, err := util.NewUUID()
	if err != nil {

	}

	team := &Team{
		PUUID:     id,
		Name:      name,
		Domain:    name,
		CreatedBy: "sysadmin",
		CreatedDt: time.Now(),
		UpdatedBy: "sysadmin",
		UpdatedDt: time.Now(),
	}

	t.teamRepository.Save(team)

	return true, nil
}

func (t service) IsAvailable(name string) (bool, error) {
	return false, nil
}

// NewService ..
func NewService(t Repository) Service {
	return &service{
		teamRepository: t,
	}
}