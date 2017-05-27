package loader

import (
	"net/http"
)

type Loader interface {
	http.Handler
	GetContents(path string) ([]byte, error)
}
