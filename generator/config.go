package generator

import (
	"regexp"
	"time"

	"path/filepath"

	"github.com/tdewolff/minify"
)

type Config struct {
	RootPath            string
	Recurse             bool
	IncludePattern      string
	includeMatcher      *regexp.Regexp
	ExcludePattern      string
	excludeMatcher      *regexp.Regexp
	OutputPath          string
	PackageName         string
	BuildConstraints    string
	DevOutputPath       string
	DevBuildConstraints string

	MinifyTypes       map[string]string         // Pattern -> content type
	minifyMatchers    map[*regexp.Regexp]string // Matcher -> content type
	CompressPattern   string
	compressMatcher   *regexp.Regexp
	NoCompressPattern string
	noCompressMatcher *regexp.Regexp
	OverrideModDate   bool

	minifier *minify.M
	now      int64
}

func NewConfig() Config {
	c := Config{
		RootPath:       ".",
		Recurse:        true,
		ExcludePattern: `\.go$`,
		OutputPath:     "embedded.go",
		PackageName:    "embedded",
		MinifyTypes: map[string]string{
			".js$":    "application/javascript",
			".css$":   "text/css",
			".html?$": "text/html",
		},
		CompressPattern:   ".(css|js|html)$",
		NoCompressPattern: `.(jpe?g|png|gif|woff2?|eot|ttf|ico)$`,
	}
	c.now = time.Now().Unix()
	c.minifier = minifier()
	return c
}

func (c *Config) normalize() {
	c.RootPath = filepath.FromSlash(c.RootPath)
	c.OutputPath = filepath.FromSlash(c.OutputPath)
	if c.DevOutputPath != "" {
		c.DevOutputPath = filepath.FromSlash(c.DevOutputPath)
	}
}

func (c *Config) buildMatchers() error {
	var err error
	if c.includeMatcher, err = regexp.Compile(c.IncludePattern); err != nil {
		return err
	}
	if c.ExcludePattern == "" {
		c.excludeMatcher = regexp.MustCompile("^$")
	} else if c.excludeMatcher, err = regexp.Compile(c.ExcludePattern); err != nil {
		return err
	}
	if c.compressMatcher, err = regexp.Compile(c.CompressPattern); err != nil {
		return err
	}
	if c.NoCompressPattern == "" {
		c.noCompressMatcher = regexp.MustCompile("^$")
	} else if c.noCompressMatcher, err = regexp.Compile(c.NoCompressPattern); err != nil {
		return err
	}
	c.minifyMatchers = make(map[*regexp.Regexp]string)
	if c.MinifyTypes == nil {
		return nil
	}
	for pat, ctype := range c.MinifyTypes {
		if mat, err := regexp.Compile(pat); err != nil {
			return err
		} else {
			c.minifyMatchers[mat] = ctype
		}
	}
	return nil
}

func (c *Config) getContentType(name string) string {
	for mat, ctype := range c.minifyMatchers {
		if mat.MatchString(name) {
			return ctype
		}
	}
	return ""
}
