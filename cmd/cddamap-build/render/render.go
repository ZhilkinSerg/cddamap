package render

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/save"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/world"
	"golang.org/x/image/font"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var dpi = 72.0
var size = 24.0
var spacing = 1.0
var cellWidth = 21.3594
var cellHeight = 24
var cellOverprintWidth = 22
var mapFont *truetype.Font
var colorCache map[color.RGBA]*image.Uniform

func init() {
	fontBytes, err := Asset("Topaz-8.ttf")
	if err != nil {
		panic(err)
	}

	mapFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}

	colorCache = make(map[color.RGBA]*image.Uniform)
}

func terrainToText(w world.World, outputRoot string, layerID int, skipEmpty bool) error {
	l := w.TerrainLayers[layerID]

	if l.Empty && skipEmpty {
		return nil
	}

	var b strings.Builder
	for _, r := range l.TerrainRows {
		for _, k := range r.TerrainCellKeys {
			c := w.TerrainCellLookup[k]
			b.WriteString(c.Symbol)
		}
		b.WriteString("\n")
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v", layerID))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(b.String())
	return nil
}

func seenToText(w world.World, outputRoot string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		var b strings.Builder
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				b.WriteString(cell.Symbol)

			}
			b.WriteString("\n")
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_%v", name, layerID))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString(b.String())
	}
	return nil
}

func Text(w world.World, outputRoot string, includeLayers []int, terrain, seen, skipEmpty bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	for _, layerID := range includeLayers {
		if terrain {
			err := terrainToText(w, outputRoot, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}
		if seen {
			err = seenToText(w, outputRoot, layerID, skipEmpty)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func terrainToImage(e *png.Encoder, rgba *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty bool) error {
	l := w.TerrainLayers[layerID]

	if l.Empty && skipEmpty {
		return nil
	}

	draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)

	pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
	for _, r := range l.TerrainRows {
		for _, k := range r.TerrainCellKeys {
			cell := w.TerrainCellLookup[k]
			bg, ok := colorCache[cell.ColorBG]
			if !ok {
				bg = image.NewUniform(cell.ColorBG)
				colorCache[cell.ColorBG] = bg
			}

			fg, ok := colorCache[cell.ColorFG]
			if !ok {
				fg = image.NewUniform(cell.ColorFG)
				colorCache[cell.ColorFG] = fg
			}

			draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
			c.SetSrc(fg)
			c.DrawString(cell.Symbol, pt)
			pt.X += c.PointToFixed(cellWidth)
		}
		pt.X = c.PointToFixed(0)
		pt.Y += c.PointToFixed(size * spacing)
	}

	filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v.png", layerID))
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	b := bufio.NewWriter(outFile)
	err = e.Encode(b, rgba)
	if err != nil {
		return err
	}

	err = b.Flush()
	if err != nil {
		return err
	}

	return nil
}

func seenToImage(e *png.Encoder, rgba *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				bg, ok := colorCache[cell.ColorBG]
				if !ok {
					bg = image.NewUniform(cell.ColorBG)
					colorCache[cell.ColorBG] = bg
				}

				fg, ok := colorCache[cell.ColorFG]
				if !ok {
					fg = image.NewUniform(cell.ColorFG)
					colorCache[cell.ColorFG] = fg
				}

				draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				c.SetSrc(fg)
				c.DrawString(cell.Symbol, pt)
				pt.X += c.PointToFixed(cellWidth)
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_%v.png", name, layerID))
		outFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer outFile.Close()

		b := bufio.NewWriter(outFile)
		err = e.Encode(b, rgba)
		if err != nil {
			return err
		}

		err = b.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

func seenToImageSolid(e *png.Encoder, rgba *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty bool) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				bg, ok := colorCache[cell.ColorBG]
				if !ok {
					bg = image.NewUniform(cell.ColorBG)
					colorCache[cell.ColorBG] = bg
				}

				draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				pt.X += c.PointToFixed(cellWidth)
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_solid_%v.png", name, layerID))
		outFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer outFile.Close()

		b := bufio.NewWriter(outFile)
		err = e.Encode(b, rgba)
		if err != nil {
			return err
		}

		err = b.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

type pool struct {
	b *png.EncoderBuffer
}

func (p *pool) Get() *png.EncoderBuffer {
	return p.b
}

func (p *pool) Put(b *png.EncoderBuffer) {
	p.b = b
}

func Image(w world.World, outputRoot string, includeLayers []int, terrain, seen, seenSolid, skipEmpty bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	e := &png.Encoder{
		BufferPool: &pool{},
	}

	if len(includeLayers) == 0 {
		return nil
	}

	l := w.TerrainLayers[includeLayers[0]]

	width := int(cellWidth * float64(len(l.TerrainRows[0].TerrainCellKeys)))
	height := cellHeight * len(l.TerrainRows)
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(mapFont)
	c.SetFontSize(size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetHinting(font.HintingNone)

	for _, layerID := range includeLayers {
		if terrain {
			terrainToImage(e, rgba, c, w, outputRoot, layerID, skipEmpty)
		}

		if seen {
			seenToImage(e, rgba, c, w, outputRoot, layerID, skipEmpty)
		}

		if seenSolid {
			seenToImageSolid(e, rgba, c, w, outputRoot, layerID, skipEmpty)
		}
	}

	return nil
}

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
