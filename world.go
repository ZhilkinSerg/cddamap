package cddamap

type World struct {
	ID   string `json:"id" db:"world_id"`
	Name string `json:"name" db:"name"`
}
