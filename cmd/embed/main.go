package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/aprice/embed/generator"
	_ "github.com/aprice/embed/loader"
)

func main() {
	cpath := flag.StringP("config", "c", "embed.json", "JSON config file")
	flag.Parse()
	cfile, err := os.Open(*cpath)
	if err != nil {
		fmt.Printf("Reading config failed: %s\n", err)
		os.Exit(2)
	}

	conf := generator.NewConfig()
	err = json.NewDecoder(cfile).Decode(&conf)
	if err != nil {
		fmt.Printf("Parsing config failed: %s", err)
		os.Exit(2)
	}

	t1 := time.Now()
	err = generator.Generate(conf)
	if err != nil {
		fmt.Printf("Generation failed: %s\n", err)
		os.Exit(2)
	} else {
		fmt.Printf("Generation complete in %v\n", time.Since(t1))
	}
}
