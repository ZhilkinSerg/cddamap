package render

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/ralreegorganon/cddamap/internal/gen/world"
	"golang.org/x/image/font"
)

var dpi = 72.0
var size = 24.0
var spacing = 1.0
var cellWidth = 21.3594
var cellHeight = 24
var cellOverprintWidth = 22
var mapFont *truetype.Font
var colorCache map[color.RGBA]*image.Uniform
var tileSize = 256

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

func Image(w world.World, outputRoot string, includeLayers []int, terrain, seen, seenSolid, skipEmpty, chop, resume bool) error {
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

	tileXCount := int(math.Ceil(float64(width) / float64(tileSize)))
	tileYCount := int(math.Ceil(float64(height) / float64(tileSize)))
	if chop {
		xPaddingRequired := tileXCount*tileSize - width
		yPaddingRequired := tileYCount*tileSize - height

		if xPaddingRequired > 0 {
			width += xPaddingRequired
		}

		if yPaddingRequired > 0 {
			height += yPaddingRequired
		}
	}

	fullImage := image.NewRGBA(image.Rect(0, 0, width, height))

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(mapFont)
	c.SetFontSize(size)
	c.SetClip(fullImage.Bounds())
	c.SetDst(fullImage)
	c.SetHinting(font.HintingNone)

	for _, layerID := range includeLayers {
		if terrain {
			err := terrainToImage(e, fullImage, c, w, outputRoot, layerID, skipEmpty, chop, resume, tileXCount, tileYCount)
			if err != nil {
				return err
			}
		}

		if seen {
			err := seenToImage(e, fullImage, c, w, outputRoot, layerID, skipEmpty, chop, resume, tileXCount, tileYCount)
			if err != nil {
				return err
			}
		}

		if seenSolid {
			err := seenToImageSolid(e, fullImage, c, w, outputRoot, layerID, skipEmpty, chop, resume, tileXCount, tileYCount)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func terrainToImage(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty, chop, resume bool, xCount, yCount int) error {
	l := w.TerrainLayers[layerID]

	if l.Empty && skipEmpty {
		return nil
	}

	draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

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

			draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
			c.SetSrc(fg)
			c.DrawString(cell.Symbol, pt)
			pt.X += c.PointToFixed(cellWidth)
		}
		pt.X = c.PointToFixed(0)
		pt.Y += c.PointToFixed(size * spacing)
	}

	if chop {
		layerTilesFolder := filepath.Join(outputRoot, fmt.Sprintf("o_%v_tiles", layerID))
		err := chopIntoTiles(e, layerTilesFolder, fullImage, xCount, yCount, resume)
		if err != nil {
			return err
		}
	} else {
		filename := filepath.Join(outputRoot, fmt.Sprintf("o_%v.png", layerID))
		outFile, err := os.Create(filename)
		if err != nil {
			return err
		}

		b := bufio.NewWriter(outFile)
		err = e.Encode(b, fullImage)
		if err != nil {
			outFile.Close()
			return err
		}

		err = b.Flush()
		if err != nil {
			outFile.Close()
			return err
		}

		outFile.Close()
	}

	return nil
}

func nativeZoom(xCount, yCount int) int {
	return int(math.Max(math.Ceil(math.Log2(float64(xCount))), math.Ceil(math.Log2(float64(yCount)))))
}

func chopIntoTiles(e *png.Encoder, layerFolder string, fullImage *image.RGBA, xCount, yCount int, resume bool) error {
	err := os.MkdirAll(layerFolder, os.ModePerm)
	if err != nil {
		return err
	}

	zCount := nativeZoom(xCount, yCount)
	bounds := fullImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for z := 0; z <= zCount; z++ {
		zFolder := filepath.Join(layerFolder, strconv.Itoa(z))
		cover := int(math.Pow(2, float64(zCount-z))) * tileSize
		txc := int(math.Ceil(float64(width) / float64(cover)))
		tyc := int(math.Ceil(float64(height) / float64(cover)))
		tile := image.NewRGBA(image.Rect(0, 0, cover, cover))
		tileBounds := tile.Bounds()

		for x := 0; x < txc; x++ {
			xFolder := filepath.Join(zFolder, strconv.Itoa(x))
			err := os.MkdirAll(xFolder, os.ModePerm)
			if err != nil {
				return err
			}

			for y := 0; y < tyc; y++ {
				filename := filepath.Join(xFolder, fmt.Sprintf("%v.png", y))

				if _, err := os.Stat(filename); resume && !os.IsNotExist(err) {
					continue
				}

				draw.Draw(tile, tileBounds, image.Transparent, image.ZP, draw.Src)
				clipRect := image.Rect(x*cover, y*cover, x*cover+cover, y*cover+cover)
				draw.Draw(tile, tileBounds, fullImage, clipRect.Min, draw.Src)

				outFile, err := os.Create(filename)
				if err != nil {
					return err
				}

				b := bufio.NewWriter(outFile)
				if tileSize == cover {
					err = e.Encode(b, tile)
					if err != nil {
						outFile.Close()
						return err
					}
				} else {
					resizedTile := imaging.Resize(tile, tileSize, tileSize, imaging.Lanczos)
					err = e.Encode(b, resizedTile)
					if err != nil {
						outFile.Close()
						return err
					}
				}
				err = b.Flush()
				if err != nil {
					return err
				}
				outFile.Close()
			}
		}
	}

	return nil
}

func seenToImage(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty, chop, resume bool, xCount, yCount int) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

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

				draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				c.SetSrc(fg)
				c.DrawString(cell.Symbol, pt)
				pt.X += c.PointToFixed(cellWidth)
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		if chop {
			layerTilesFolder := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_%v_tiles", name, layerID))
			err := chopIntoTiles(e, layerTilesFolder, fullImage, xCount, yCount, resume)
			if err != nil {
				return err
			}
		} else {
			filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_%v.png", name, layerID))
			outFile, err := os.Create(filename)
			if err != nil {
				return err
			}

			b := bufio.NewWriter(outFile)
			err = e.Encode(b, fullImage)
			if err != nil {
				outFile.Close()
				return err
			}

			err = b.Flush()
			if err != nil {
				outFile.Close()
				return err
			}

			outFile.Close()
		}
	}

	return nil
}

func seenToImageSolid(e *png.Encoder, fullImage *image.RGBA, c *freetype.Context, w world.World, outputRoot string, layerID int, skipEmpty, chop, resume bool, xCount, yCount int) error {
	for name, layers := range w.SeenLayers {
		l := layers[layerID]

		if l.Empty && skipEmpty {
			continue
		}

		draw.Draw(fullImage, fullImage.Bounds(), image.Black, image.ZP, draw.Src)

		pt := freetype.Pt(0, 0+int(c.PointToFixed(size)>>6))
		for _, r := range l.SeenRows {
			for _, k := range r.SeenCellKeys {
				cell := w.SeenCellLookup[k]
				bg, ok := colorCache[cell.ColorBG]
				if !ok {
					bg = image.NewUniform(cell.ColorBG)
					colorCache[cell.ColorBG] = bg
				}

				draw.Draw(fullImage, image.Rect(int(pt.X>>6), int(pt.Y>>6), int(pt.X>>6)+cellOverprintWidth, int(pt.Y>>6)-cellHeight), bg, image.ZP, draw.Src)
				pt.X += c.PointToFixed(cellWidth)
			}
			pt.X = c.PointToFixed(0)
			pt.Y += c.PointToFixed(size * spacing)
		}

		if chop {
			layerTilesFolder := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_solid_%v_tiles", name, layerID))
			err := chopIntoTiles(e, layerTilesFolder, fullImage, xCount, yCount, resume)
			if err != nil {
				return err
			}
		} else {
			filename := filepath.Join(outputRoot, fmt.Sprintf("%v_visible_solid_%v.png", name, layerID))
			outFile, err := os.Create(filename)
			if err != nil {
				return err
			}

			b := bufio.NewWriter(outFile)
			err = e.Encode(b, fullImage)
			if err != nil {
				outFile.Close()
				return err
			}

			err = b.Flush()
			if err != nil {
				outFile.Close()
				return err
			}

			outFile.Close()
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
