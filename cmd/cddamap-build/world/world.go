package world

import (
	"fmt"
	"image"

	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/metadata"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/save"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func indexOf(slice []int, item int) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type World struct {
	Layers []Layer
}

type Layer struct {
	Rows []Row
}

type Row struct {
	Cells []Cell
}

type Cell struct {
	Symbol  string
	ColorFG *image.Uniform
	ColorBG *image.Uniform
}

func Build(m *metadata.Metadata, s *save.Save) (*World, error) {
	missingTerrain := make(map[string]int)
	for _, c := range s.Overmap.Chunks {
		for _, l := range c.Layers {
			for _, e := range l {
				if exists := m.Overmap.Exists(e.OvermapTerrainID); !exists {
					if _, ok := missingTerrain[e.OvermapTerrainID]; !ok {
						missingTerrain[e.OvermapTerrainID] = 0
					}
					missingTerrain[e.OvermapTerrainID]++
				}
			}
		}
	}

	for k, v := range missingTerrain {
		fmt.Printf("missing terrain: %v x %v\n", k, v)
	}

	cXMax := 0
	cXMin := 0
	cYMax := 0
	cYMin := 0

	for _, c := range s.Overmap.Chunks {
		if c.X > cXMax {
			cXMax = c.X
		}
		if c.Y > cYMax {
			cYMax = c.Y
		}
		if c.X < cXMin {
			cXMin = c.X
		}
		if c.Y < cYMin {
			cYMin = c.Y
		}
	}

	cXSize := abs(cXMax) + abs(cXMin) + 1
	cYSize := abs(cYMax) + abs(cYMin) + 1

	chunkCapacity := cXSize * cYSize

	doneChunks := make(map[int]bool)
	cells := make([]Cell, 680400*chunkCapacity)
	for _, c := range s.Overmap.Chunks {
		ci := c.X + (0 - cXMin) + cXSize*(c.Y+0-cYMin)
		doneChunks[ci] = true
		for li, l := range c.Layers {
			lzp := 0
			for _, e := range l {
				s := m.Overmap.Symbol(e.OvermapTerrainID)
				cfg, cbg := m.Overmap.Color(e.OvermapTerrainID)
				for i := 0; i < int(e.Count); i++ {
					tmi := ci*680400 + li*32400 + lzp
					cells[tmi] = Cell{
						Symbol:  s,
						ColorFG: cfg,
						ColorBG: cbg,
					}
					lzp++
				}
			}
		}
	}

	dfg, dbg := m.Overmap.Color("default")
	for i := 0; i < chunkCapacity; i++ {
		if _, ok := doneChunks[i]; !ok {
			for e := 0; e < 680400; e++ {
				cells[i*680400+e] = Cell{
					Symbol:  " ",
					ColorFG: dfg,
					ColorBG: dbg,
				}
			}
		}
	}

	layers := make([]Layer, 21)
	for l := 0; l < 21; l++ {
		layers[l].Rows = make([]Row, 180*cYSize)
		for r := 0; r < 180*cYSize; r++ {
			layers[l].Rows[r].Cells = make([]Cell, 180*cXSize)
		}
	}

	for li := 0; li < 21; li++ {
		for xi := 0; xi < cXSize; xi++ {
			for yi := 0; yi < cYSize; yi++ {
				for ri := 0; ri < 180; ri++ {
					for ci := 0; ci < 180; ci++ {
						layers[li].Rows[yi*180+ri].Cells[xi*180+ci] = cells[(xi+yi*cXSize)*680400+li*32400+ri*180+ci]
					}
				}
			}
		}
	}

	return &World{Layers: layers}, nil
}
