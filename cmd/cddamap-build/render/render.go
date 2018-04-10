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

	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/world"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

func Text(w *world.World, outputRoot string, includeLayers []int) error {
	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	for _, i := range includeLayers {
		l := w.Layers[i]
		filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v", i))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		var b strings.Builder
		for _, r := range l.Rows {
			for _, c := range r.Cells {
				b.WriteString(c.Symbol)
			}
			b.WriteString("\n")
		}
		f.WriteString(b.String())
	}
	return nil
}

func Image(w *world.World, outputRoot string, includeLayers []int) error {
	dpi := 72.0
	size := 24.0
	spacing := 1.0
	cellWidth := 21.3594
	cellHeight := 24
	cellOverprintWidth := 22

	err := os.MkdirAll(outputRoot, os.ModePerm)
	if err != nil {
		return err
	}

	fontBytes, err := Asset("Topaz-8.ttf")
	if err != nil {
		return err
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	for _, i := range includeLayers {
		l := w.Layers[i]

		width := int(cellWidth * float64(len(l.Rows[0].Cells)))
		height := cellHeight * len(l.Rows)
		rgba := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
		c := freetype.NewContext()
		c.SetDPI(dpi)
		c.SetFont(f)
		c.SetFontSize(size)
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetHinting(font.HintingNone)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.Rows {
			for _, cell := range r.Cells {
				draw.Draw(rgba, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), cell.ColorBG, image.ZP, draw.Src)
				c.SetSrc(cell.ColorFG)
				c.DrawString(cell.Symbol, pt)
				pt.X += c.PointToFixed(cellWidth)
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v.png", i))
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
