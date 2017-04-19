package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// CreatePackage creates a debian package
func CreatePackage(reposlug, ghToken, email, version, archs, outbuild string, push bool) {

	x := strings.Split(reposlug, "/")
	user := x[0]
	name := x[1]

	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", reposlug)
	fmt.Println("repoPath", repoPath)

	setupGitRepo(repoPath, reposlug, user, email)
	chdir(repoPath)

	maybesudo(`apt-get install build-essential lintian curl -y`)

	if tryexec(`latest -v`) != nil {
		exec(`git clone https://github.com/mh-cbon/latest.git %v/src/github.com/mh-cbon/latest`, gopath)
		exec(`go install github.com/mh-cbon/latest`)
	}

	if tryexec(`changelog -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/changelog")
	}

	if tryexec(`go-bin-deb -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/go-bin-deb")
	}

	dir, err := ioutil.TempDir("", "go-bin-deb")
	if err != nil {
		panic(err)
	}

	exec(`ls -al %v`, repoPath)

	for _, arch := range strings.Split(archs, ",") {
		arch = strings.TrimSpace(arch)
		arch = strings.ToLower(arch)
		if arch == "i386" {
			arch = "386"
		} else if arch == "x64" {
			arch = "amd64"
		}

		workDir := filepath.Join(dir, arch)
		outFile := fmt.Sprintf("%v-%v.deb", name, arch)
		out := filepath.Join(outbuild, outFile)

		mkdirAll(workDir)
		exec(`go-bin-deb generate -a %v --version %v -w %v -o %v`, arch, version, workDir, out)
	}

	exec(`ls -al .`)
	exec(`ls -al %v`, outbuild)

	if push == true {
		pushAssetsGh(version, ghToken, outbuild, "*.deb")
	}

}
