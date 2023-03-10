// Package go-bin-deb creates binary package for debian system
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mh-cbon/go-bin-deb/debian"
	"github.com/mh-cbon/verbose"
	"github.com/urfave/cli"
)

// VERSION is the last build number.
var VERSION = "0.0.0"
var logger = verbose.Auto()

func main() {
	app := cli.NewApp()
	app.Name = "go-bin-deb"
	app.Version = VERSION
	app.Usage = "Generate a binary debian package"
	app.UsageText = "go-bin-deb <cmd> <options>"
	app.Commands = []cli.Command{
		{
			Name:   "generate",
			Usage:  "Generate the contents of the package",
			Action: generateContents,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "wd, w",
					Value: "pkg-build",
					Usage: "Working directory to prepare the package",
				},
				cli.StringFlag{
					Name:  "output, o",
					Value: "",
					Usage: "Output directory for the debian package files",
				},
				cli.StringFlag{
					Name:  "file, f",
					Value: "deb.json",
					Usage: "Path to the deb.json file",
				},
				cli.StringFlag{
					Name:  "version",
					Value: "",
					Usage: "Version of the package",
				},
				cli.StringFlag{
					Name:  "arch, a",
					Value: "",
					Usage: "Arch of the package",
				},
				cli.StringFlag{
					Name:  "compression, z",
					Value: "",
					Usage: "Compression to use (via dpkg-deb -Z)",
				},
			},
		},
		{
			Name:   "test",
			Usage:  "Test the package json file",
			Action: testPkg,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Value: "deb.json",
					Usage: "Path to the deb.json file",
				},
			},
		},
	}

	app.Run(os.Args)
}

func generateContents(c *cli.Context) error {
	output := c.String("output")
	wd := c.String("wd")
	file := c.String("file")
	version := c.String("version")
	arch := c.String("arch")
	compression := c.String("compression")

	pkgDir := filepath.Join(wd)

	if o, err := filepath.Abs(output); err != nil {
		return cli.NewExitError(err.Error(), 1)
	} else {
		output = o
	}

	debJSON := debian.Package{}

	// load the deb.json file
	if err := debJSON.Load(file); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	logger.Println("deb.json loaded")

	// normalize data
	debJSON.Normalize(pkgDir, version, arch)
	logger.Println("pkg data normalized")

	logger.Printf("Generating files in %s", pkgDir)
	if err := debJSON.GenerateFiles(filepath.Dir(file), pkgDir); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	extraArgs := []string{}
	if compression != "" {
		extraArgs = append(extraArgs, "-Z"+compression)
	}
	logger.Printf("Building package in %s to %s with extra args %v", wd, output, extraArgs)
	if err := buildPackage(pkgDir, output, extraArgs); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	logger.Printf("linting package in %s to %s", wd, output)
	lintPackage(pkgDir, output) // it does not need to fail.

	return nil
}

func testPkg(c *cli.Context) error {
	file := c.String("file")

	debJSON := debian.Package{}

	if err := debJSON.Load(file); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Println("File is correct")

	return nil
}

func buildPackage(wd string, output string, extraArgs []string) error {
	args := []string{"dpkg-deb", "--build", "debian", output}
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}
	oCmd := exec.Command("fakeroot", args...)
	oCmd.Dir = wd
	oCmd.Stdout = os.Stdout
	oCmd.Stderr = os.Stderr
	return oCmd.Run()
}

func lintPackage(wd string, output string) error {
	oCmd := exec.Command("lintian", output)
	oCmd.Dir = wd
	oCmd.Stdout = os.Stdout
	oCmd.Stderr = os.Stderr
	return oCmd.Run()
}
