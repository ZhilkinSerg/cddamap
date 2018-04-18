package main

import (
	"encoding/json"
	//"github.com/pkg/errors"
	"github.com/ralreegorganon/cddamap/internal/server"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	//"github.com/asticode/go-astilog"
	_ "github.com/lib/pq"
	
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

func blam(ci cellIndex) (interface{}, error) {
	var db server.DB
	
	if err := db.Open("postgres://cddamap:cddamap@localhost:9432/cddamap?sslmode=disable"); err != nil {
		return nil, err
	}

	j, err := db.GetCellJson(ci.L, ci.X, ci.Y)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err = json.Unmarshal(j, &m); err != nil {
		return nil, err
	}

	return m, nil
}