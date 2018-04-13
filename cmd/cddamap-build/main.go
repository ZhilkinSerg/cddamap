package main

import (
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/jessevdk/go-flags"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/metadata"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/render"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/save"
	"github.com/ralreegorganon/cddamap/cmd/cddamap-build/world"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	GameRoot           string `short:"g" long:"game" required:"true" description:"Cataclysm: DDA game root directory"`
	Save               string `short:"s" long:"save" required:"true" description:"Game save directory to process"`
	OutputDir          string `short:"o" long:"output" required:"true" description:"Output folder"`
	Text               bool   `short:"t" long:"text" description:"Render to text files"`
	Images             bool   `short:"i" long:"images" description:"Render to images"`
	Layers             []int  `short:"l" long:"layer" description:"Layer to render, 0-20. Repeat flag for multiple layers or omit for all."`
	DBConnectionString string `short:"c" long:"connectionString" description:"PostGIS database connection string"`
	Terrain            bool   `short:"r" long:"terrain" description:"Render terrain"`
	Seen               bool   `short:"e" long:"seen" description:"Render seen"`
	SeenSolid          bool   `short:"d" long:"seensolid" description:"Render seen as a solid overlay"`
	SkipEmpty          bool   `short:"k" long:"skipempty" description:"Skip rendering empty layers"`
	Chop               bool   `short:"p" long:"chop" description:"Chop images into tiles instead of rendering as a single image"`
	Resume             bool   `short:"z" long:"resume" description:"Resume tile building, instead of overwriting"`
}

func init() {
	f := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(f)
}

func main() {
	// defer profile.Start().Stop()
	// defer profile.Start(profile.MemProfile).Stop()
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if len(opts.Layers) == 0 {
		for i := 0; i < 21; i++ {
			opts.Layers = append(opts.Layers, i)
		}
	}

	s, err := save.Build(opts.Save)
	if err != nil {
		log.Fatal(err)
	}

	o, err := metadata.Build(s, opts.GameRoot)
	if err != nil {
		log.Fatal(err)
	}

	w, err := world.Build(o, s)
	if err != nil {
		log.Fatal(err)
	}

	if opts.Text {
		err = render.Text(w, opts.OutputDir, opts.Layers, opts.Terrain, opts.Seen, opts.SkipEmpty)
		if err != nil {
			log.Fatal(err)
		}
	}

	if opts.Images {
		err = render.Image(w, opts.OutputDir, opts.Layers, opts.Terrain, opts.Seen, opts.SeenSolid, opts.SkipEmpty, opts.Chop, opts.Resume)
		if err != nil {
			log.Fatal(err)
		}
	}

	if opts.DBConnectionString != "" {
		err = render.GIS(w, opts.DBConnectionString, opts.Layers, opts.Terrain, opts.Seen, opts.SkipEmpty)
		if err != nil {
			log.Fatal(err)
		}
	}
}
