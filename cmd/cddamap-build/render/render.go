package render

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func init() {
	fontBytes, err := Asset("Topaz-8.ttf")
	if err != nil {
		panic(err)
	}

	mapFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
}

func terrainToText(w *world.World, outputRoot string, layerID int) error {
	l := w.TerrainLayers[layerID]
	filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v", layerID))
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	var b strings.Builder
	for _, r := range l.TerrainRows {
		for _, c := range r.TerrainCells {
			b.WriteString(c.Symbol)
		}
		b.WriteString("\n")
	}
	f.WriteString(b.String())
	return nil
}

func seenToText(w *world.World, outputRoot string, layerID int) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]
		filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_%v", name, layerID))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		var b strings.Builder
		for _, r := range l.SeenRows {
			for _, c := range r.SeenCells {
				if c.Seen {
					b.WriteString(" ")
				} else {
					b.WriteString("#")
				}
			}
			b.WriteString("\n")
		}
		f.WriteString(b.String())
	}
	return nil
}

func Text(w *world.World, outputRoot string, includeLayers []int, terrain, seen bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	for _, layerID := range includeLayers {
		if terrain {
			err := terrainToText(w, outputRoot, layerID)
			if err != nil {
				return err
			}
		}
		if seen {
			err = seenToText(w, outputRoot, layerID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func terrainToImage(w *world.World, outputRoot string, layerID int) error {
	l := w.TerrainLayers[layerID]

	width := int(cellWidth * float64(len(l.TerrainRows[0].TerrainCells)))
	height := cellHeight * len(l.TerrainRows)
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(mapFont)
	c.SetFontSize(size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetHinting(font.HintingNone)

	pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
	for _, r := range l.TerrainRows {
		for _, cell := range r.TerrainCells {
			draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), cell.ColorBG, image.ZP, draw.Src)
			c.SetSrc(cell.ColorFG)
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
	err = png.Encode(b, rgba)
	if err != nil {
		return err
	}

	err = b.Flush()
	if err != nil {
		return err
	}

	return nil
}

func seenToImage(w *world.World, outputRoot string, layerID int) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		width := int(cellWidth * float64(len(l.SeenRows[0].SeenCells)))
		height := cellHeight * len(l.SeenRows)
		rgba := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
		c := freetype.NewContext()
		c.SetDPI(dpi)
		c.SetFont(mapFont)
		c.SetFontSize(size)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetHinting(font.HintingNone)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, cell := range r.SeenCells {
				draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), cell.ColorBG, image.ZP, draw.Src)
				c.SetSrc(cell.ColorFG)
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
		err = png.Encode(b, rgba)
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

func seenToImageSolid(w *world.World, outputRoot string, layerID int) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		width := int(cellWidth * float64(len(l.SeenRows[0].SeenCells)))
		height := cellHeight * len(l.SeenRows)
		rgba := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
		c := freetype.NewContext()
		c.SetDPI(dpi)
		c.SetFont(mapFont)
		c.SetFontSize(size)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetHinting(font.HintingNone)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, cell := range r.SeenCells {
				draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), cell.ColorBG, image.ZP, draw.Src)
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
		err = png.Encode(b, rgba)
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

func Image(w *world.World, outputRoot string, includeLayers []int, terrain, seen, seenSolid bool) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	for _, layerID := range includeLayers {
		if terrain {
			terrainToImage(w, outputRoot, layerID)

		}

		if seen {
			seenToImage(w, outputRoot, layerID)

			if seenSolid {
				seenToImageSolid(w, outputRoot, layerID)
			}
		}
	}

	return nil
}

func GIS(w *world.World, connectionString string, includeLayers []int, terrain, seen bool) error {
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	var worldID int
	err = db.QueryRow("insert into world (name) values ($1) returning world_id", w.Name).Scan(&worldID)
	if err != nil {
		return err
	}

	for _, i := range includeLayers {
		var layerID int
		err = db.QueryRow("insert into layer (world_id, z) values ($1, $2) returning layer_id", worldID, i).Scan(&layerID)
		if err != nil {
			return err
		}

		if terrain {
			txn, err := db.Begin()
			if err != nil {
				return err
			}

			stmt, err := txn.Prepare(pq.CopyIn("cell", "layer_id", "id", "name", "the_geom"))
			if err != nil {
				return err
			}

			l := w.TerrainLayers[i]

			for ri, r := range l.TerrainRows {
				for ci, c := range r.TerrainCells {
					x := float64(ci) * 21.3594
					y := float64(ri) * 24.0
					x2 := x + 21.3594
					y2 := y + 24.0

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
