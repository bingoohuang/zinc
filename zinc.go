package zinc

import (
	"embed"
	"github.com/prabhatsharma/zinc/pkg/auth"
	"github.com/prabhatsharma/zinc/pkg/core"
	"io/fs"
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

func Init() {
	core.Init()
	auth.Init()
}
