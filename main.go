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

func generateContents (c *cli.Context) error {
  output := c.String("output")
  wd := c.String("wd")
  file := c.String("file")
  version := c.String("version")
  arch := c.String("arch")

  pkgDir := filepath.Join(wd)

  if o, err := filepath.Abs(output); err!=nil {
    return cli.NewExitError(err.Error(), 1)
  } else {
    output = o
  }

  debJson := debian.Package{}

  // load the deb.json file
  if err := debJson.Load(file); err!=nil {
    return cli.NewExitError(err.Error(), 1)
  }
  logger.Println("deb.json loaded")

  // normalize data
  debJson.Normalize(pkgDir, version, arch)
  logger.Println("pkg data normalized")

  logger.Printf("Generating files in %s", pkgDir)
  if err := debJson.GenerateFiles(filepath.Dir(file), pkgDir); err !=nil {
    return cli.NewExitError(err.Error(), 1)
  }

  logger.Printf("Building package in %s to %s", wd, output)
  if err := buildPackage(pkgDir, output); err !=nil {
    return cli.NewExitError(err.Error(), 1)
  }

  logger.Printf("linting package in %s to %s", wd, output)
  if err := lintPackage(pkgDir, output); err !=nil {
    return cli.NewExitError(err.Error(), 1)
  }

  // logger.Printf("Copying results from %s to %s", wd, output)
  // if err := debJson.CopyResults(wd, output); err !=nil {
  //   return cli.NewExitError(err.Error(), 1)
  // }


  return nil
}

func testPkg (c *cli.Context) error {
  file := c.String("file")

  debJson := debian.Package{}

  if err := debJson.Load(file); err!=nil {
    return cli.NewExitError(err.Error(), 1)
  }

  fmt.Println("File is correct")

  return nil
}

func buildPackage (wd string, output string) error {
  oCmd := exec.Command("fakeroot", "dpkg-deb", "--build", "debian", output)
  oCmd.Dir = wd
  oCmd.Stdout = os.Stdout
  oCmd.Stderr = os.Stderr
  return oCmd.Run()
}

func lintPackage (wd string, output string) error {
  oCmd := exec.Command("lintian", output)
  oCmd.Dir = wd
  oCmd.Stdout = os.Stdout
  oCmd.Stderr = os.Stderr
  return oCmd.Run()
}
