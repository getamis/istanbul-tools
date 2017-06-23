package utils

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// NewApp creates an app with sane defaults.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	app.Email = ""
	return app
}
