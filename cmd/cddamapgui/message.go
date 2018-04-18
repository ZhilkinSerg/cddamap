package main

import (
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/jmoiron/sqlx"
	"github.com/paulmach/go.geojson"
)

// handleMessages handles messages
func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	switch m.Name {
	case "cell":
		var ci cellIndex
		if len(m.Payload) > 0 {
			if err = json.Unmarshal(m.Payload, &ci); err != nil {
				payload = err.Error()
				return
			}
		}

		if payload, err = blam(ci); err != nil {
			payload = err.Error()
			return
		}
	}
	return
}

type cellIndex struct {
	L int
	X float64
	Y float64
}

type cell struct {
	X1   float64 `db:"x1"`
	X2   float64 `db:"x2"`
	Y1   float64 `db:"y1"`
	Y2   float64 `db:"y2"`
	Name string  `db:"name"`
}

func blam(ci cellIndex) (interface{}, error) {
	db, err := sqlx.Open("sqlite3", "/Users/jj/Desktop/TrinityCenter/map.db")
	if err != nil {
		return nil, err
	}

	sql := `
		select
			x1, y1, x2, y2, name
		from
			cell
		where
			layer = $1 
			and x1 <= $2 
			and x2 >= $2
			and y1 <= $3
			and y2 >= $3
		`

	cells := []*cell{}
	err = db.Select(&cells, sql, ci.L, ci.X, ci.Y)
	if err != nil {
		return nil, err
	}

	fc := geojson.NewFeatureCollection()

	for _, c := range cells {
		foo := [][][]float64{
			[][]float64{
				[]float64{c.X1, c.Y1},
				[]float64{c.X2, c.Y1},
				[]float64{c.X2, c.Y2},
				[]float64{c.X1, c.Y2},
				[]float64{c.X1, c.Y1},
			},
		}
		f := geojson.NewPolygonFeature(foo)
		f.Properties["name"] = c.Name
		fc.AddFeature(f)
	}

	rawJSON, err := fc.MarshalJSON()

	var m map[string]interface{}
	if err = json.Unmarshal(rawJSON, &m); err != nil {
		return nil, err
	}

	return m, nil
}
