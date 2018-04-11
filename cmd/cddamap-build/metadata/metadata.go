package metadata

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/save"
	log "github.com/sirupsen/logrus"
)

type Overmap struct {
	templates map[string]overmapTerrain
	built     map[string]overmapTerrain
	symbols   map[int]string
	rotations [][]int
}

type overmapTerrain struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Abstract   string   `json:"abstract"`
	Name       string   `json:"name"`
	Sym        int      `json:"sym"`
	Color      string   `json:"color"`
	CopyFrom   string   `json:"copy-from"`
	SeeCost    int      `json:"see_cost"`
	Extras     string   `json:"extras"`
	MonDensity int      `json:"mondensity"`
	Flags      []string `json:"flags"`
	Spawns     spawns   `json:"spawns"`
	MapGen     []mapGen `json:"mapgen"`
}

type spawns struct {
	Group      string `json:"group"`
	Population []int  `json:"population"`
	Chance     int    `json:"chance"`
}

type mapGen struct {
	Method string `json:"method"`
	Name   string `json:"name"`
}

type modInfo struct {
	Ident string `json:"ident"`
}

const overmapTerrainTypeID = "overmap_terrain"

type inLoadOrder []string

func (s inLoadOrder) Len() int {
	return len(s)
}

func (s inLoadOrder) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s inLoadOrder) Less(i, j int) bool {
	c1 := strings.Count(s[i], "/")
	c2 := strings.Count(s[j], "/")

	if c1 == c2 {
		return s[i] < s[j]
	}
	return c1 < c2
}

func indexOf(slice []int, item int) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

var linearSuffixes = []string{
	"_isolated",
	"_end_south",
	"_end_west",
	"_ne",
	"_end_north",
	"_ns",
	"_es",
	"_nes",
	"_end_east",
	"_wn",
	"_ew",
	"_new",
	"_sw",
	"_nsw",
	"_esw",
	"_nesw"}

var linearSuffixSymbols = map[string]int{
	"_isolated":  0,
	"_end_south": 4194424,
	"_end_west":  4194417,
	"_ne":        4194413,
	"_end_north": 4194424,
	"_ns":        4194424,
	"_es":        4194412,
	"_nes":       4194420,
	"_end_east":  4194417,
	"_wn":        4194410,
	"_ew":        4194417,
	"_new":       4194422,
	"_sw":        4194411,
	"_nsw":       4194421,
	"_esw":       4194423,
	"_nesw":      4194414,
}

var rotationSuffixes = []string{
	"_north",
	"_east",
	"_south",
	"_west"}

var symbols map[int]string

var rotations [][]int

type ColorPair struct {
	FG *image.Uniform
	BG *image.Uniform
}

var colors map[string]ColorPair

func init() {
	symbols = map[int]string{
		4194424: "\u2502",
		4194417: "\u2500",
		4194413: "\u2514",
		4194412: "\u250c",
		4194411: "\u2510",
		4194410: "\u2518",
		4194420: "\u251c",
		4194422: "\u2534",
		4194421: "\u2524",
		4194423: "\u252c",
		4194414: "\u253c",
	}

	for i := 0; i < 128; i++ {
		symbols[i] = string(i)
	}

	rotations = make([][]int, 0)
	rotations = append(rotations, []int{60, 94, 62, 118})
	rotations = append(rotations, []int{4194410, 4194413, 4194412, 4194411})
	rotations = append(rotations, []int{4194417, 4194424, 4194417, 4194424})
	rotations = append(rotations, []int{4194420, 4194423, 4194421, 4194422})

	white := image.NewUniform(color.RGBA{150, 150, 150, 255})
	black := image.NewUniform(color.RGBA{0, 0, 0, 255})
	red := image.NewUniform(color.RGBA{255, 0, 0, 255})
	green := image.NewUniform(color.RGBA{0, 110, 0, 255})
	brown := image.NewUniform(color.RGBA{92, 51, 23, 255})
	blue := image.NewUniform(color.RGBA{0, 0, 200, 255})
	magenta := image.NewUniform(color.RGBA{139, 58, 98, 255})
	cyan := image.NewUniform(color.RGBA{0, 150, 180, 255})
	gray := image.NewUniform(color.RGBA{150, 150, 150, 255})
	darkGray := image.NewUniform(color.RGBA{99, 99, 99, 255})
	lightRed := image.NewUniform(color.RGBA{255, 150, 150, 255})
	lightGreen := image.NewUniform(color.RGBA{0, 255, 0, 255})
	yellow := image.NewUniform(color.RGBA{255, 255, 0, 255})
	lightBlue := image.NewUniform(color.RGBA{100, 100, 255, 255})
	lightMagenta := image.NewUniform(color.RGBA{254, 0, 254, 255})
	lightCyan := image.NewUniform(color.RGBA{0, 240, 255, 255})

	colors = make(map[string]ColorPair)

	colors["black_yellow"] = ColorPair{FG: black, BG: yellow}
	colors["blue"] = ColorPair{FG: blue, BG: black}
	colors["brown"] = ColorPair{FG: brown, BG: black}
	colors["c_yellow_green"] = ColorPair{FG: yellow, BG: green}
	colors["cyan"] = ColorPair{FG: cyan, BG: black}
	colors["dark_gray"] = ColorPair{FG: darkGray, BG: black}
	colors["dark_gray_magenta"] = ColorPair{FG: darkGray, BG: magenta}
	colors["green"] = ColorPair{FG: green, BG: black}
	colors["h_dark_gray"] = ColorPair{FG: darkGray, BG: black}
	colors["h_yellow"] = ColorPair{FG: yellow, BG: black}
	colors["i_blue"] = ColorPair{FG: black, BG: blue}
	colors["i_brown"] = ColorPair{FG: black, BG: brown}
	colors["i_cyan"] = ColorPair{FG: black, BG: cyan}
	colors["i_green"] = ColorPair{FG: black, BG: green}
	colors["i_light_blue"] = ColorPair{FG: black, BG: lightBlue}
	colors["i_light_cyan"] = ColorPair{FG: black, BG: lightCyan}
	colors["i_light_gray"] = ColorPair{FG: black, BG: gray}
	colors["i_light_green"] = ColorPair{FG: black, BG: lightGreen}
	colors["i_light_red"] = ColorPair{FG: black, BG: lightRed}
	colors["i_magenta"] = ColorPair{FG: black, BG: magenta}
	colors["i_pink"] = ColorPair{FG: black, BG: lightMagenta}
	colors["i_red"] = ColorPair{FG: black, BG: red}
	colors["i_yellow"] = ColorPair{FG: black, BG: yellow}
	colors["light_blue"] = ColorPair{FG: lightBlue, BG: black}
	colors["light_cyan"] = ColorPair{FG: lightCyan, BG: black}
	colors["light_gray"] = ColorPair{FG: gray, BG: black}
	colors["light_green"] = ColorPair{FG: lightGreen, BG: black}
	colors["light_green_yellow"] = ColorPair{FG: lightGreen, BG: yellow}
	colors["light_red"] = ColorPair{FG: lightRed, BG: black}
	colors["magenta"] = ColorPair{FG: magenta, BG: black}
	colors["pink"] = ColorPair{FG: lightMagenta, BG: black}
	colors["pink_magenta"] = ColorPair{FG: lightMagenta, BG: magenta}
	colors["red"] = ColorPair{FG: red, BG: black}
	colors["white"] = ColorPair{FG: white, BG: black}
	colors["white_magenta"] = ColorPair{FG: white, BG: magenta}
	colors["white_white"] = ColorPair{FG: white, BG: white}
	colors["yellow"] = ColorPair{FG: yellow, BG: black}
	colors["yellow_cyan"] = ColorPair{FG: yellow, BG: cyan}
	colors["yellow_magenta"] = ColorPair{FG: yellow, BG: magenta}
	colors["unset"] = ColorPair{FG: white, BG: black}
}

type Metadata struct {
	Overmap Overmap
}

func Build(save *save.Save, gameRoot string) (*Metadata, error) {
	m := &Metadata{
		Overmap: Overmap{
			templates: make(map[string]overmapTerrain),
			built:     make(map[string]overmapTerrain),
			symbols:   symbols,
			rotations: rotations,
		},
	}

	jsonRoot := path.Join(gameRoot, "data", "json")
	modsRoot := path.Join(gameRoot, "data", "mods")
	files, err := sourceFiles(jsonRoot, modsRoot, save.Mods)
	if err != nil {
		return nil, err
	}

	err = m.Overmap.loadTemplatesFromFiles(files)
	if err != nil {
		return nil, err
	}

	err = m.Overmap.buildTemplates()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (o *Overmap) Exists(id string) bool {
	_, ok := o.built[id]
	return ok
}

func (o *Overmap) Symbol(id string) string {
	if t, tok := o.built[id]; tok {
		if s, sok := o.symbols[t.Sym]; sok {
			return s
		}
	}
	return "?"
}

func (o *Overmap) Color(id string) (*image.Uniform, *image.Uniform) {
	if c, tok := o.built[id]; tok {
		if cp, ok := colors[c.Color]; ok {
			return cp.FG, cp.BG
		}
	}
	unset := colors["unset"]
	return unset.FG, unset.BG
}

func (o *Overmap) Name(id string) string {
	if t, tok := o.built[id]; tok {
		return t.Name
	}
	return "?"
}

func sourceFiles(jsonRoot, modsRoot string, saveMods []string) ([]string, error) {
	files := []string{}

	err := filepath.Walk(jsonRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	activeMods := map[string]string{}
	for _, m := range saveMods {
		activeMods[m] = m
	}

	mods, err := ioutil.ReadDir(modsRoot)
	if err != nil {
		return nil, err
	}

	for _, f := range mods {
		if !f.IsDir() {
			continue
		}

		modInfoPath := path.Join(modsRoot, f.Name(), "modinfo.json")
		b, err := ioutil.ReadFile(modInfoPath)
		if err != nil {
			return nil, err
		}
		var modinfo []modInfo
		err = json.Unmarshal(b, &modinfo)
		if err != nil {
			return nil, err
		}

		if _, ok := activeMods[modinfo[0].Ident]; !ok {
			continue
		}

		err = filepath.Walk(path.Join(modsRoot, f.Name()), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".json") {
				files = append(files, path)
			}
			return nil
		})
	}

	sort.Sort(inLoadOrder(files))

	return files, nil
}

func (o *Overmap) loadTemplates(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if !bytes.Contains(b, []byte(overmapTerrainTypeID)) {
		return nil
	}

	var temp []map[string]interface{}
	err = json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	filteredOvermapTerrains := make([]map[string]interface{}, 0)
	for _, t := range temp {
		if t["type"].(string) == overmapTerrainTypeID {
			filteredOvermapTerrains = append(filteredOvermapTerrains, t)
		}
	}

	filteredText, err := json.Marshal(filteredOvermapTerrains)
	if err != nil {
		return err
	}

	var overmapTerrains []overmapTerrain
	err = json.Unmarshal(filteredText, &overmapTerrains)
	if err != nil {
		return err
	}

	for _, ot := range overmapTerrains {
		if ot.Type != overmapTerrainTypeID {
			continue
		}
		if ot.Abstract != "" {
			o.templates[ot.Abstract] = ot
		} else {
			o.templates[ot.ID] = ot
		}
	}

	return nil
}

func (o *Overmap) loadTemplatesFromFiles(files []string) error {
	for _, f := range files {
		f, err := os.Open(f)
		if err != nil {
			return err
		}
		defer f.Close()

		o.loadTemplates(f)
	}

	return nil
}

func (o *Overmap) buildTemplates() error {
	for _, ot := range o.templates {
		bt := make([]overmapTerrain, 0)
		t := ot
		bt = append(bt, t)
		for t.CopyFrom != "" {
			t = o.templates[t.CopyFrom]
			bt = append(bt, t)
		}

		b := overmapTerrain{}
		for i := len(bt) - 1; i >= 0; i-- {
			if err := mergo.Merge(&b, bt[i], mergo.WithOverride); err != nil {
				return err
			}
		}

		if ot.Abstract == "" {
			b.Abstract = ""
			b.CopyFrom = ""
			o.built[b.ID] = b

			rotate := true

			if b.Flags != nil {
				for _, f := range b.Flags {
					if f == "NO_ROTATE" {
						rotate = false
					} else if f == "LINEAR" {
						for _, suffix := range linearSuffixes {
							bs := overmapTerrain{}
							if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
								return err
							}
							bs.ID = b.ID + suffix
							bs.Sym = linearSuffixSymbols[suffix]
							o.built[bs.ID] = bs
						}
					}
				}
			}

			if rotate {
				for i, suffix := range rotationSuffixes {
					bs := overmapTerrain{}
					if err := mergo.Merge(&bs, b, mergo.WithOverride); err != nil {
						log.Fatal(err)
					}
					bs.ID = b.ID + suffix

					for _, r := range o.rotations {
						index := indexOf(r, b.Sym)
						if index != -1 {
							bs.Sym = r[(i+index+4)%4]
							break
						}
					}
					o.built[bs.ID] = bs
				}
			}
		}
	}

	return nil
}
