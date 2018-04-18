package render

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
)

func GIS(w world.World, connectionString string, includeLayers []int, terrain, seen, skipEmpty bool) error {
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	var worldID int
	err = db.QueryRow("insert into world (name) values ($1) returning world_id", w.Name).Scan(&worldID)
	if err != nil {
		return err
	}

	emptyRockHash := save.HashTerrainID("empty_rock")
	openAirHash := save.HashTerrainID("empty_rock")
	blankHash := save.HashTerrainID("empty_rock")

	for _, i := range includeLayers {
		if terrain {
			l := w.TerrainLayers[i]

			if l.Empty && skipEmpty {
				continue
			}

			var layerID int
			err = db.QueryRow("insert into layer (world_id, z) values ($1, $2) returning layer_id", worldID, i).Scan(&layerID)
			if err != nil {
				return err
			}

			txn, err := db.Begin()
			if err != nil {
				return err
			}

			stmt, err := txn.Prepare(pq.CopyIn("cell", "layer_id", "id", "name", "the_geom"))
			if err != nil {
				return err
			}

			for ri, r := range l.TerrainRows {
				for ci, k := range r.TerrainCellKeys {
					if k == emptyRockHash || k == openAirHash || k == blankHash {
						continue
					}

					x := float64(ci) * cellWidth
					y := float64(ri) * float64(cellHeight)
					x2 := x + cellWidth
					y2 := y + float64(cellHeight)

					c := w.TerrainCellLookup[k]

					geom := fmt.Sprintf("POLYGON((%[1]f %[2]f, %[3]f %[4]f, %[5]f %[6]f, %[7]f %[8]f, %[1]f %[2]f))", x, y, x2, y, x2, y2, x, y2)
					_, err = stmt.Exec(layerID, c.ID, c.Name, geom)
					if err != nil {
						return err
					}
				}
			}
			_, err = stmt.Exec()
			if err != nil {
				return err
			}

			err = stmt.Close()
			if err != nil {
				return err
			}

			err = txn.Commit()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
