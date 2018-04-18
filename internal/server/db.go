package server

import (
	"fmt"

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

func (db *DB) GetCellJson(layerID int, x, y float64) ([]byte, error) {
	sql := fmt.Sprintf(`
		select
			row_to_json(fc) geojson
		from
			(
				select
					'FeatureCollection' as type,
					array_to_json(array_agg(f)) as features
				from
				(
					select
						'Feature' as type,
						st_asgeojson(the_geom)::json as geometry,
						json_build_object(
							'id', id, 
							'name', name
						) as properties
					from
						cell
					where 
						layer_id = $1
						and ST_CoveredBy(ST_GeomFromText('POINT(%[1]f %[2]f)'), the_geom)
				) as f
			) as fc
		`, x, y)

	var json []byte
	err := db.QueryRow(sql, layerID).Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}
