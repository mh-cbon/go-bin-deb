package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mh-cbon/go-bin-deb/stringexec"
)

// CreatePackage creates a debian package
func CreatePackage(reposlug, ghToken, email, version, archs, outbuild string, push bool) {

	x := strings.Split(reposlug, "/")
	user := x[0]
	name := x[1]

	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", reposlug)
	fmt.Println("repoPath", repoPath)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		os.MkdirAll(repoPath, os.ModePerm)
		os.Chdir(repoPath)
		exec(`git clone https://github.com/%v.git .`, reposlug)
		exec(`git config user.name %v`, user)
		exec(`git config user.email %v`, email)
	}

	os.Chdir(repoPath)
	fmt.Println("Chdir", repoPath)

	exec(`sudo apt-get install build-essential lintian curl -y`)

	if tryexec(`latest -v`) != nil {
		exec(`go get -u github.com/mh-cbon/latest`)
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

		os.MkdirAll(workDir, os.ModePerm)
		exec(`go-bin-deb generate -a %v --version %v -w %v -o %v`, arch, version, workDir, out)
	}

	exec(`ls -al .`)
	exec(`ls -al %v`, outbuild)

	if push == true {
		if tryexec(`gh-api-cli -v`) != nil {
			exec(`latest -repo=%v`, "mh-cbon/gh-api-cli")
		}
		exec(`gh-api-cli rm-assets --owner %v --repository %v --ver %v -t %v --glob %v`, user, name, version, ghToken, "*.deb")
		exec(`gh-api-cli upload-release-asset --owner %v --repository %v --ver %v -t %v --glob %q`, user, name, version, ghToken, outbuild+"/*.deb")
	}
}

var alwaysHide = map[string]string{}

func clean(s string) string {
	for search, replace := range alwaysHide {
		s = strings.Replace(s, search, replace, -1)
	}
	return s
}

func tryexec(w string, params ...interface{}) error {
	w = fmt.Sprintf(w, params...)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println("exec", clean(w))
	cmd, err := stringexec.Command(cwd, w)
	if err != nil {
		return err
	}
	// out, err := cmd.CombinedOutput()
	// sout := string(out)
	// fmt.Println(clean(sout))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func exec(w string, params ...interface{}) {
	if err := tryexec(w, params...); err != nil {
		panic(err)
	}
}
