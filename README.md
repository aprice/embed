# embed
Static content embedding for Golang, ©2017 Adrian Price. Usage indicates acceptance
of the license found in the LICENSE file.

## Purpose & Functionality
Embed is a tool for embedding static content files into a Go binary, to create
a single, self-contained binary. Embed uses build constraints to offer a development
mode that loads content from disk for rapid iteration during local development;
once content is stable, you can run `go generate` to update the embedded content
files, and then rebuild with different tags to serve the embedded content instead
of the files on disk.

In addition to simply embedding content within your binary, embed will minify
HTML, CSS, and JavaScript; gzip files for serving to clients that accept compressed
content; and calculate checksums for handling Etag-based conditional requests.

## Installation
Embedding tool: `go get github.com/aprice/embed/cmd/embed`

Usage library: `go get github.com/aprice/loader`

## Usage
Embed can easily be run with `go generate`.

```go
//go:generate embed -c "embed.json"
```

To use embed, create a config file specifying what's to be generated:
```json
{
	"RootPath": ".",
	"Recurse": true,
	"IncludePattern": "",
	"ExcludePattern": "(^\\.|\\.go$)",
	"OutputPath": "embedded.go",
	"BuildConstraints": "",
	"PackageName": "embedded",
	"DevOutputPath": "",
	"DevBuildConstraints": "",
	"MinifyTypes": {
		"\\.html?$": "text/html",
		"\\.css$": "text/css",
		"\\.js$": "application/javascript"
	},
	"CompressPattern": "\\.(css|js|html)$",
	"NoCompressPattern": "\\.(jpe?g|png|gif|woff2?|eot|ttf|ico)$",
	"OverrideModDate": false
}
```

The values above are the defaults.

`RootPath` is the directory where source files to be embedded will be scanned.
If `Recurse` is true, subdirectories will be scanned as well. Each entry will
be compared against `IncludePattern` and `ExcludePattern`; if a file does not
match `IncludePattern` or does match `ExcludePattern`, it will not be included.

`OutputPath` is the path where the embedded content go file will be written. If
`BuildConstraints` is not empty, it will be added to the output file; for example,
`"BuildConstraints": "!dev"` will result in a file that will not be built by
a build command including `-tags="dev"`. `PackageName` is the package name that
will be used for the output file.

`DevOutputPath` and `DevBuildConstraints` work the same as their non-`Dev`
counterparts, but apply to a separate "dev mode" file; if `DevOutputPath` is not
empty, a dev mode file will be written which reads all content from disk instead
of using embedded content. This allows for rapid iteration during local development.
The dev mode file will use `PackageName` for its package.

`MinifyTypes` is a mapping of file name regular expressions to content types that
should be minified. Minifiers are enabled for `text/html`, `text/css`,
`text/javascript` (or `application/javascript`), and `image/svg+xml`. Any other
content type (or an empty content type) will not be minified. The minifier used
is [github.com/tdewolff/minify](https://github.com/tdewolff/minify).

`CompressPattern` is a regular expression matching file names that should be
gzip compressed for clients that accept compressed data. `NoCompressPattern`
is for excluding files which otherwise match `CompressPattern`.

If `OverrideModDate` is true, the modification date for embedded files will be
set to the timestamp when generation is run; otherwise, it will be the modification
date of the source files.

## Referencing Embedded Content
To reference embedded content, call the `GetEmbeddedContent()` function in the
package where your generated file was created. This returns a `Loader`, which
can be used directly as an `http.Handler` to serve the embedded content the
same as `http.FileServer(http.Dir(RootPath))`. It also exposes a `GetContents()`
method, for loading embedded content as a byte slice for programmatic use,
such as embedding template files.