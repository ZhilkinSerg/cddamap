package cddamap

import (
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func (db *DB) Open(connectionString string) error {
	d, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	db.DB = d
	return nil
}

func (db *DB) GetWorlds() ([]*World, error) {
	worlds := []*World{}
	err := db.Select(&worlds, `
		select 
			world_id, 
			name
		from 
			world
	`)
	if err != nil {
		return nil, err
	}
	return worlds, nil
}
