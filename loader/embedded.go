package loader

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type EmbeddedLoader struct {
	h       http.Handler
	content map[string]*Content
	dirs    map[string]embeddedDir
}

// NewOnDisk creates a new Loader that loads embedded content.
func New() *EmbeddedLoader {
	l := &EmbeddedLoader{
		content: make(map[string]*Content),
		dirs:    map[string]embeddedDir{"/": embeddedDir{name: "/"}},
	}
	l.h = http.FileServer(l)
	return l
}

func (l *EmbeddedLoader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := path.Clean("/" + r.URL.Path)
	c := l.content[name]
	if r.Method == http.MethodGet && len(c.CompressedBytes) > 0 && r.Header.Get("Range") == "" && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Etag", `"`+c.Hash+`-gzip"`)
		done, _ := checkPreconditions(w, r, c.Modified)
		if done {
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		ctype := mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			n := 512
			ln := len(c.RawBytes)
			if ln < 512 {
				n = ln
			}
			ctype = http.DetectContentType(c.RawBytes[:n])
		}
		w.Header().Set("Content-Type", ctype)
		w.Write(c.CompressedBytes)
		return
	}

	w.Header().Set("Etag", `"`+c.Hash+`"`)
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	serveFile(w, r, l, path.Clean(upath), true)
}

// Add an embedded file to the Loader.
func (l *EmbeddedLoader) Add(c *Content) {
	if len(c.Compressed) > 0 {
		c.CompressedBytes, _ = base64.RawStdEncoding.DecodeString(c.Compressed)
	} else {
		c.CompressedBytes = make([]byte, 0)
	}
	if strings.TrimSpace(c.Raw) != "" {
		c.RawBytes, _ = base64.RawStdEncoding.DecodeString(c.Raw)
	} else if len(c.CompressedBytes) > 0 {
		gzr, _ := gzip.NewReader(bytes.NewReader(c.CompressedBytes))
		c.RawBytes, _ = ioutil.ReadAll(gzr)
	}
	c.Raw = ""
	c.Compressed = ""
	l.content[c.Path] = c
	for d := path.Dir(c.Path); d != "/"; d = path.Dir(d) {
		if _, ok := l.dirs[d]; !ok {
			l.dirs[d] = embeddedDir{name: d}
		}
	}
}

func (l *EmbeddedLoader) GetContents(path string) ([]byte, error) {
	if c, ok := l.content[path]; ok {
		return c.RawBytes, nil
	}
	return nil, os.ErrNotExist
}
