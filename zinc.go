package zinc

import (
	"embed"
	"io/fs"

	"github.com/prabhatsharma/zinc/pkg/auth"
	"github.com/prabhatsharma/zinc/pkg/core"
)

//go:embed web/dist
var embedFrontend embed.FS

// FrontendAssets returns the frontend assets.
var FrontendAssets = func() fs.FS {
	f, err := fs.Sub(embedFrontend, "web/dist")
	if err != nil {
		panic(err)
	}

	return f
}()

// Init initializes zinc.
func Init() {
	core.Init()
	auth.Init()
}
