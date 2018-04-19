package render

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ralreegorganon/cddamap/internal/gen/save"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
)

func Sqlite(w world.World, outputRoot string, includeLayers []int, terrain, seen, skipEmpty bool) error {
	filename := filepath.Join(outputRoot, "map.db")
	db, _ := sql.Open("sqlite3", filename)

	createStmt, _ := db.Prepare("create table if not exists cell (id integer primary key, layer integer, x1 double, y1 double, x2 double, y2 double, name text)")
	createStmt.Exec()

	emptyRockHash := save.HashTerrainID("empty_rock")
	openAirHash := save.HashTerrainID("empty_rock")
	blankHash := save.HashTerrainID("empty_rock")

	for _, i := range includeLayers {
		if terrain {
			l := w.TerrainLayers[i]

			if l.Empty && skipEmpty {
				continue
			}

			txn, err := db.Begin()
			if err != nil {
				return err
			}

			stmt, err := txn.Prepare("insert into cell (layer, x1, y1, x2, y2, name) values (?,?,?,?,?,?)")
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

					_, err = stmt.Exec(i, x, y, x2, y2, c.Name)
					if err != nil {
						return err
					}
				}
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

	indexStmt, _ := db.Prepare("create index idx_cell_all on cell(layer, x1, x2, y1, y2, name)")
	indexStmt.Exec()

	return nil
}
