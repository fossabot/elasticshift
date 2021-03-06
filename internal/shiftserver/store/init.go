/*
Copyright 2018 The Elasticshift Authors.
*/
package store

import mgo "gopkg.in/mgo.v2"

type Shift struct {
	Team           Team
	Vcs            Vcs
	Sysconf        Sysconf
	Build          Build
	Repository     Repository
	Plugin         Plugin
	Container      Container
	Integration    Integration
	Infrastructure Infrastructure
	Defaults       Defaults
	Secret         Secret
	Shiftfile      Shiftfile
}

type Database struct {
	Session *mgo.Session
	Name    string
}

// New ..
// Create a new base datasource
func NewDatabase(dbname string, session *mgo.Session) Database {
	return Database{Name: dbname, Session: session}
}

func (db Database) InitShiftStore() Shift {

	return Shift{
		Team:        newTeamStore(db),
		Vcs:         newVcsStore(db),
		Sysconf:     newSysconfStore(db),
		Build:       newBuildStore(db),
		Repository:  newRepositoryStore(db),
		Plugin:      newPluginStore(db),
		Container:   newContainerStore(db),
		Integration: newIntegrationStore(db),
		Defaults:    newDefaultsStore(db),
		Secret:      newSecretStore(db),
		Shiftfile:   newShiftfileStore(db),
	}
}
