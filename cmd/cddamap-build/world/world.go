package world

import (
	"fmt"
	"image"
	"image/color"

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
	Name          string
	TerrainLayers []TerrainLayer
	SeenLayers    map[string][]SeenLayer
}

type TerrainLayer struct {
	TerrainRows []TerrainRow
}

type TerrainRow struct {
	TerrainCells []TerrainCell
}

type TerrainCell struct {
	Symbol  string
	ColorFG *image.Uniform
	ColorBG *image.Uniform
	Name    string
	ID      string
}

type SeenLayer struct {
	SeenRows []SeenRow
}

type SeenRow struct {
	SeenCells []SeenCell
}

type SeenCell struct {
	Symbol  string
	Seen    bool
	ColorFG *image.Uniform
	ColorBG *image.Uniform
}

func Build(m *metadata.Metadata, s *save.Save) (*World, error) {
	terrainLayers := buildTerrainLayers(m, s)
	characterSeenLayers := buildCharacterSeenLayers(m, s)

	world := &World{
		Name:          s.Name,
		TerrainLayers: terrainLayers,
		SeenLayers:    characterSeenLayers,
	}

	return world, nil
}

type worldChunkDimensions struct {
	XSize int
	YSize int
	XMin  int
	XMax  int
	YMin  int
	YMax  int
}

func calculateWorldChunkDimensions(m *metadata.Metadata, s *save.Save) worldChunkDimensions {
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

	wcd := worldChunkDimensions{
		XSize: cXSize,
		YSize: cYSize,
		XMin:  cXMin,
		XMax:  cXMax,
		YMin:  cYMin,
		YMax:  cYMax,
	}
	return wcd
}

func buildCharacterSeenLayers(m *metadata.Metadata, s *save.Save) map[string][]SeenLayer {
	wcd := calculateWorldChunkDimensions(m, s)
	chunkCapacity := wcd.XSize * wcd.YSize
	dfg := image.NewUniform(color.RGBA{44, 44, 44, 255})
	dbg := image.NewUniform(color.RGBA{0, 0, 0, 255})

	seen := make(map[string][]SeenLayer)

	for name, chunks := range s.Seen {
		doneChunks := make(map[int]bool)
		cells := make([]SeenCell, 680400*chunkCapacity)
		for _, c := range chunks.Chunks {
			ci := c.X + (0 - wcd.XMin) + wcd.XSize*(c.Y+0-wcd.YMin)
			doneChunks[ci] = true
			for li, l := range c.Visible {
				lzp := 0
				for _, e := range l {
					for i := 0; i < int(e.Count); i++ {
						tmi := ci*680400 + li*32400 + lzp

						if e.Seen {
							cells[tmi] = SeenCell{
								Symbol:  " ",
								Seen:    true,
								ColorFG: image.Transparent,
								ColorBG: image.Transparent,
							}
						} else {
							cells[tmi] = SeenCell{
								Symbol:  " ",
								Seen:    false,
								ColorFG: dfg,
								ColorBG: dbg,
							}
						}

						lzp++
					}
				}
			}
		}

		for i := 0; i < chunkCapacity; i++ {
			if _, ok := doneChunks[i]; !ok {
				for e := 0; e < 680400; e++ {
					cells[i*680400+e] = SeenCell{
						Symbol:  " ",
						Seen:    false,
						ColorFG: dfg,
						ColorBG: dbg,
					}
				}
			}
		}

		layers := make([]SeenLayer, 21)
		for l := 0; l < 21; l++ {
			layers[l].SeenRows = make([]SeenRow, 180*wcd.YSize)
			for r := 0; r < 180*wcd.YSize; r++ {
				layers[l].SeenRows[r].SeenCells = make([]SeenCell, 180*wcd.XSize)
			}
		}

		for li := 0; li < 21; li++ {
			for xi := 0; xi < wcd.XSize; xi++ {
				for yi := 0; yi < wcd.YSize; yi++ {
					for ri := 0; ri < 180; ri++ {
						for ci := 0; ci < 180; ci++ {
							layers[li].SeenRows[yi*180+ri].SeenCells[xi*180+ci] = cells[(xi+yi*wcd.XSize)*680400+li*32400+ri*180+ci]
						}
					}
				}
			}
		}
		seen[name] = layers
	}

	return seen
}

func buildTerrainLayers(m *metadata.Metadata, s *save.Save) []TerrainLayer {
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

	wcd := calculateWorldChunkDimensions(m, s)
	chunkCapacity := wcd.XSize * wcd.YSize

	doneChunks := make(map[int]bool)
	cells := make([]TerrainCell, 680400*chunkCapacity)
	for _, c := range s.Overmap.Chunks {
		ci := c.X + (0 - wcd.XMin) + wcd.XSize*(c.Y+0-wcd.YMin)
		doneChunks[ci] = true
		for li, l := range c.Layers {
			lzp := 0
			for _, e := range l {
				s := m.Overmap.Symbol(e.OvermapTerrainID)
				cfg, cbg := m.Overmap.Color(e.OvermapTerrainID)
				n := m.Overmap.Name(e.OvermapTerrainID)

				for i := 0; i < int(e.Count); i++ {
					tmi := ci*680400 + li*32400 + lzp
					cells[tmi] = TerrainCell{
						ID:      e.OvermapTerrainID,
						Symbol:  s,
						ColorFG: cfg,
						ColorBG: cbg,
						Name:    n,
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
				cells[i*680400+e] = TerrainCell{
					Symbol:  " ",
					ColorFG: dfg,
					ColorBG: dbg,
				}
			}
		}
	}

	layers := make([]TerrainLayer, 21)
	for l := 0; l < 21; l++ {
		layers[l].TerrainRows = make([]TerrainRow, 180*wcd.YSize)
		for r := 0; r < 180*wcd.YSize; r++ {
			layers[l].TerrainRows[r].TerrainCells = make([]TerrainCell, 180*wcd.XSize)
		}
	}

	for li := 0; li < 21; li++ {
		for xi := 0; xi < wcd.XSize; xi++ {
			for yi := 0; yi < wcd.YSize; yi++ {
				for ri := 0; ri < 180; ri++ {
					for ci := 0; ci < 180; ci++ {
						layers[li].TerrainRows[yi*180+ri].TerrainCells[xi*180+ci] = cells[(xi+yi*wcd.XSize)*680400+li*32400+ri*180+ci]
					}
				}
			}
		}
	}

	return layers
}
