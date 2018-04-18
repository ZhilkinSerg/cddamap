package server

type World struct {
	ID   string `json:"id" db:"world_id"`
	Name string `json:"name" db:"name"`
}

type Layer struct {
	ID      string `json:"id" db:"layer_id"`
	WorldID string `json:"worldId" db:"world_id"`
	Z       int    `json:"z" db:"z"`
}

type Cell struct {
	ID      string `json:"id" db:"cell_id"`
	LayerID string `json:"layerID" db:"layer_id"`
	Name    string `json:"name" db:"name"`
}
