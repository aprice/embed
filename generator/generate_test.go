package generator

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	conf := Config{
		IncludePattern: ".(css|js|html)$",
		MinifyTypes: map[string]string{
			".js$":    "application/javascript",
			".css$":   "text/css",
			".html?$": "text/html",
			".jpe?g$": "image/jpeg",
		},
		CompressPattern:  ".(css|js|html)$",
		BuildConstraints: "!dev",
		PackageName:      "embedded",
		RootPath:         "../loader/testdata",
	}
	err := conf.buildMatchers()
	if err != nil {
		t.Error(err)
	}
	conf.minifier = minifier()
	conf.normalize()
	buf := new(bytes.Buffer)
	err = generate(conf, buf)
	if err != nil {
		t.Error(err)
	}
	stat, err := os.Stat("../loader/testdata/example.html")
	if err != nil {
		t.Error(err)
	}
	mt := fmt.Sprintf("%v", stat.ModTime().Unix())
	expected := `//+build !dev

package embedded

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

	l.Add(&loader.Content{
		Path:   "/example.html",
		Hash:    "BqZLX-Gc5JCfp_iX562hjA",
		Modified: time.Unix(` + mt + `, 0),
		Compressed: ` + "`" + `
H4sIAAAAAAAA/3yTTY/TMBCG7/yKYbmA1DQtsFClbgQSSFy4ceE4jcfNaG1PZE/SVlX/O0qyZUF8XCJ5
Yr/vk0eOeW6l0XNH0GrwtXl8EtraKKun+vMJQ+cJPklAjqacpyaQIjQtpky669UVm//MIgbaDUzHTpJC
I1Ep6u7uyFbbnaWBGyqmxYIjK6MvcoOeduu72mQ9e6r3Ys+XPTYPhyR9tEUjXlL1wq3cyr3eBkwHjtVq
26G1HA/VauskauEwsD9X0lGEjDEvWvIDKTcIkXpafLktFx8To1+Me4pMid3V8nCZkKp3q1V3ulXcUwDs
VX423Y8v/wLm3HYvyVIqElruc7WmcMXKc3xYYDVwZiV7edz9ZvN2s3FbpZMWlhpJqCyxihLp+iGQZXwZ
8DQrqt6PPK8u/zLifmWfUG92ZvDfqZ6UjXxXU866TTlfgbGkNpaH2rTrP65Cu65NV39rOYOdRsAZKCvu
PeeWLKjAnqDPZMFJAva+zzp+3EBAc1gGjmCl6QNFzUv4Lj0EPI+HQFvOz27J8enEkbWVXqFLLAkakWQ5
TspAEmB+4HiYCjtKgXNmiUtTdiOsQWgTuV2r2lVleTwel4wRl5IO5dyUy8ee+qskAo5OUpjCl8ulKbGe
ksrJSTn7Kae/5kcAAAD///DU6xJLAwAA
` + "`" + `,
	})

	_embeddedContentLoader = l
}`
	actual := strings.TrimSpace(buf.String())
	if actual != expected {
		t.Errorf("Actual did not match expected.\nWant:\n%s\nGot:\n%s\n", expected, actual)
	}
}

func TestGenerateFile(t *testing.T) {
	conf := Config{
		MinifyTypes: map[string]string{
			".html?$": "text/html",
		},
		CompressPattern: ".html$",
	}
	err := conf.buildMatchers()
	if err != nil {
		t.Error(err)
	}
	conf.minifier = minifier()
	conf.normalize()
	file := strings.NewReader(exampleDotOrg)
	buf := new(bytes.Buffer)
	err = generateFile(file, buf, "/example.html", conf, 1483228800)
	if err != nil {
		t.Error(err)
	}
	expected := `l.Add(&loader.Content{
		Path:   "/example.html",
		Hash:    "BqZLX-Gc5JCfp_iX562hjA",
		Modified: time.Unix(1483228800, 0),
		Compressed: ` + "`" + `
H4sIAAAAAAAA/3yTTY/TMBCG7/yKYbmA1DQtsFClbgQSSFy4ceE4jcfNaG1PZE/SVlX/O0qyZUF8XCJ5
Yr/vk0eOeW6l0XNH0GrwtXl8EtraKKun+vMJQ+cJPklAjqacpyaQIjQtpky669UVm//MIgbaDUzHTpJC
I1Ep6u7uyFbbnaWBGyqmxYIjK6MvcoOeduu72mQ9e6r3Ys+XPTYPhyR9tEUjXlL1wq3cyr3eBkwHjtVq
26G1HA/VauskauEwsD9X0lGEjDEvWvIDKTcIkXpafLktFx8To1+Me4pMid3V8nCZkKp3q1V3ulXcUwDs
VX423Y8v/wLm3HYvyVIqElruc7WmcMXKc3xYYDVwZiV7edz9ZvN2s3FbpZMWlhpJqCyxihLp+iGQZXwZ
8DQrqt6PPK8u/zLifmWfUG92ZvDfqZ6UjXxXU866TTlfgbGkNpaH2rTrP65Cu65NV39rOYOdRsAZKCvu
PeeWLKjAnqDPZMFJAva+zzp+3EBAc1gGjmCl6QNFzUv4Lj0EPI+HQFvOz27J8enEkbWVXqFLLAkakWQ5
TspAEmB+4HiYCjtKgXNmiUtTdiOsQWgTuV2r2lVleTwel4wRl5IO5dyUy8ee+qskAo5OUpjCl8ulKbGe
ksrJSTn7Kae/5kcAAAD///DU6xJLAwAA
` + "`" + `,
	})`
	actual := strings.TrimSpace(buf.String())
	if actual != expected {
		t.Errorf("Actual did not match expected.\nWant:\n%s\nGot:\n%s\nNo match\n", expected, actual)
	}
}

func TestGetSourceFiles(t *testing.T) {
	conf := Config{RootPath: "."}
	conf.buildMatchers()
	conf.normalize()
	files, err := getSourceFiles(conf, conf.RootPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", files)
	// TODO: Validate against expected output
}

const exampleDotOrg = `
<!doctype html>
<html>
<head>
    <title>Example Domain</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;

    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 50px;
        background-color: #fff;
        border-radius: 1em;
    }
    a:link, a:visited {
        color: #38488f;
        text-decoration: none;
    }
    @media (max-width: 700px) {
        body {
            background-color: #fff;
        }
        div {
            width: auto;
            margin: 0 auto;
            border-radius: 0;
            padding: 1em;
        }
    }
    </style>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is established to be used for illustrative examples in documents. You may use this
    domain in examples without prior coordination or asking for permission.</p>
    <p><a href="http://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>
`
