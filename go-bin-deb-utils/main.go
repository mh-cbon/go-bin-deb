// go-bin-deb-utils is a cli tool to generate debian package and repos.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	flag.Parse()
	action := flag.Arg(0)

	// basic arg parsing
	var reposlug string
	var email string
	var ghToken string
	var version string
	var archs string
	var out string

	flag.StringVar(&reposlug, "repo", "", "The repo slug such USER/REPO.")
	flag.StringVar(&ghToken, "ghToken", "", "The ghToken to write on your repository.")
	flag.StringVar(&email, "email", "", "Your gh email.")
	flag.StringVar(&version, "version", "", "The package version.")
	flag.StringVar(&archs, "archs", "386,amd64", "The archs to build.")
	flag.StringVar(&out, "out", "", "The out build directory.")
	push := flag.Bool("push", false, "Push the new assets")
	flag.CommandLine.Parse(os.Args[2:])

	// os.Env fallback
	email = readEnv(email, "EMAIL", "MYEMAIL")
	reposlug = readEnv(reposlug, "REPO")
	ghToken = readEnv(ghToken, "GH_TOKEN")

	// ci fallback
	// todo: make use of pre defined ci env
	if isTravis() {
		version = readEnv(version, "TRAVIS_TAG")
		out = readEnv(out, "TRAVIS_BUILD_DIR")
	}
	if isVagrant() {
		version = readEnv(version, "VERSION")
		out = readEnv(out, "BUILD_DIR")
	}

	// integrity check
	requireArg(reposlug, "repo", "REPO")
	requireArg(ghToken, "ghToken", "GH_TOKEN")
	requireArg(email, "email", "EMAIL", "MYEMAIL")
	if isTravis() {
		requireArg(version, "version", "TRAVIS_TAG")
		requireArg(out, "out", "TRAVIS_BUILD_DIR")
	} else if isVagrant() {
		requireArg(version, "version", "VERSION")
		requireArg(out, "out", "BUILD_DIR")
	} else {
		panic("nop, no such ci system...")
	}

	// execute some common setup, in case.
	alwaysHide[ghToken] = "$GH_TOKEN"

	// removeAll(out)
	mkdirAll(out)

	if version == "LAST" {
		version = latestGhRelease(reposlug)
	}

	// execute the action
	if action == "create-packages" {
		CreatePackage(reposlug, ghToken, email, version, archs, out, *push)

	} else if action == "setup-repository" {
		SetupRepo(reposlug, ghToken, email, version, archs, out, *push)
	}
}

func requireArg(val, n string, env ...string) {
	if val == "" {
		log.Printf("missing argument -%v or env %q\n", n, env)
		os.Exit(1)
	}
}

func readEnv(c string, k ...string) string {
	if c == "" {
		for _, kk := range k {
			c = os.Getenv(kk)
			if c != "" {
				break
			}
		}
	}
	return c
}

func mkdirAll(f string) error {
	fmt.Println("mkdirAll", f)
	return os.MkdirAll(f, os.ModePerm)
}
func removeAll(f string) error {
	fmt.Println("removeAll", f)
	return os.RemoveAll(f)
}
func chdir(f string) error {
	fmt.Println("Chdir", f)
	return os.Chdir(f)
}

func isTravis() bool {
	return strings.ToLower(os.Getenv("CI")) == "true" &&
		strings.ToLower(os.Getenv("TRAVIS")) == "true"
}

func isVagrant() bool {
	_, s := os.Stat("/vagrant/")
	return !os.IsNotExist(s)
}

func latestGhRelease(repo string) string {
	ret := ""
	u := fmt.Sprintf(`https://api.github.com/repos/%v/releases/latest`, repo)
	fmt.Println("latestGhRelease", u)
	r := getURL(u)
	k := map[string]interface{}{}
	json.Unmarshal(r, &k)

	if x, ok := k["tag_name"]; ok {
		ret = x.(string)
	} else {
		panic("latest version not found")
	}
	fmt.Println("latestGhRelease", ret)
	return ret
}
