package generator

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/svg"
)

func Generate(conf Config) error {
	conf.normalize()
	err := conf.buildMatchers()
	if err != nil {
		return fmt.Errorf("failed compiling patterns: %s", err)
	}

	out, err := os.OpenFile(conf.OutputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file %s: %s", conf.OutputPath, err)
	}
	err = generate(conf, out)
	if err != nil {
		return fmt.Errorf("failed generating file: %s", err)
	}
	out.Close()

	if conf.DevOutputPath != "" {
		out, err = os.OpenFile(conf.DevOutputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open dev file %s: %s", conf.DevOutputPath, err)
		}
		generateDev(conf, out)
		out.Close()
	}
	return nil
}

func generate(conf Config, out io.Writer) error {
	header(out, conf)
	files, err := getSourceFiles(conf, conf.RootPath)
	if err != nil {
		return fmt.Errorf("failed getting source files: %s", err)
	}
	for _, fpath := range files {
		f, err := os.Open(fpath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", fpath, err)
		}
		defer f.Close()
		t := conf.now
		if !conf.OverrideModDate {
			stat, err := f.Stat()
			if err != nil {
				return fmt.Errorf("failed to stat %s: %s", fpath, err)
			}
			t = stat.ModTime().Unix()
		}
		err = generateFile(f, out, strings.TrimPrefix(fpath, conf.RootPath), conf, t)
		if err != nil {
			return fmt.Errorf("failed to embed %s: %s", fpath, err)
		}
	}
	footer(out)
	return nil
}

func generateDev(conf Config, out io.Writer) {
	if conf.DevBuildConstraints != "" {
		fmt.Fprintf(out, "//+build %s\n\n", conf.DevBuildConstraints)
	}

	fmt.Fprintf(out, `package %s

import (
	"sync"
	"github.com/aprice/embed/loader"
)

var _embeddedContentLoader loader.Loader
var _initOnce sync.Once

// GetEmbeddedContent returns the Loader for embedded content files.
func GetEmbeddedContent() loader.Loader {
	_initOnce.Do(func() {
		_embeddedContentLoader = loader.NewOnDisk("%s")
	})
	return _embeddedContentLoader
}
`, conf.PackageName, conf.RootPath)
}

func getSourceFiles(conf Config, root string) ([]string, error) {
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed listing directory %s: %s", root, err)
	}
	files := make([]string, 0)
	for _, fi := range entries {
		fpath := filepath.Join(root, fi.Name())
		if conf.excludeMatcher.Match([]byte(fpath)) || !conf.includeMatcher.Match([]byte(fpath)) {
			continue
		}
		if fi.IsDir() {
			if !conf.Recurse {
				continue
			}
			children, err := getSourceFiles(conf, fpath)
			if err != nil {
				return nil, fmt.Errorf("failed getting children of %s: %s", fpath, err)
			}
			files = append(files, children...)
		} else {
			files = append(files, fpath)
		}
	}
	return files, nil
}

func minifier() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", &html.Minifier{
		KeepConditionalComments: true,
		KeepDocumentTags:        true,
		KeepEndTags:             true,
	})
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	return m
}

func header(out io.Writer, conf Config) {
	if conf.BuildConstraints != "" {
		fmt.Fprintf(out, "//+build %s\n\n", conf.BuildConstraints)
	}

	fmt.Fprintf(out, `package %s

import (
	"sync"
	"time"
	"github.com/aprice/embed/loader"
)

var _embeddedContentLoader loader.Loader
var _initOnce sync.Once

// GetEmbeddedContent returns the Loader for embedded content files.
func GetEmbeddedContent() loader.Loader {
	_initOnce.Do(_initEmbeddedContent)
	return _embeddedContentLoader
}

func _initEmbeddedContent() {
	l := loader.New()
`, conf.PackageName)
}

func generateFile(in io.Reader, out io.Writer, name string, conf Config, modTime int64) error {
	ctype := conf.getContentType(name)
	var contents []byte
	var err error
	if ctype != "" {
		contents, err = ioutil.ReadAll(conf.minifier.Reader(ctype, in))
	} else {
		contents, err = ioutil.ReadAll(in)
	}
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", name, err)
	}
	fmt.Fprintf(out, `
	l.Add(&loader.Content{
		Path:   %q,
		Hash:    %q,
		Modified: time.Unix(%v, 0),
`,
		filepath.ToSlash(name), hash(contents), modTime)

	if conf.compressMatcher.MatchString(name) && !conf.noCompressMatcher.MatchString(name) {
		fmt.Fprintf(out, "\t\tCompressed: `\n%s`,\n", encodeB64(compress(contents)))
	} else {
		fmt.Fprintf(out, "\t\tRaw: `\n%s`,\n", encodeB64(contents))
	}
	fmt.Fprint(out, "\t})\n")
	return nil
}

func footer(out io.Writer) {
	fmt.Fprint(out, `
	_embeddedContentLoader = l
}
`)
}

func hash(in []byte) string {
	hash := md5.Sum(in)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func encodeB64(in []byte) string {
	var buf bytes.Buffer
	b64 := base64.NewEncoder(base64.RawStdEncoding, &buf)
	b64.Write(in)
	b64.Close()
	var out string
	chunk := make([]byte, 80)
	for n, _ := buf.Read(chunk); n > 0; n, _ = buf.Read(chunk) {
		out += string(chunk[0:n]) + "\n"
	}
	return out
}

func compress(in []byte) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(in)
	gz.Close()
	return buf.Bytes()
}
