package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SetupRepo creates a debian repository
func SetupRepo(reposlug, ghToken, email, version, archs, outbuild string, push bool) {

	x := strings.Split(reposlug, "/")
	user := x[0]
	name := x[1]

	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", reposlug)
	fmt.Println("repoPath", repoPath)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		mkdirAll(repoPath)
		chdir(repoPath)
		exec(`git clone https://github.com/%v.git .`, reposlug)
		exec(`git config user.name %v`, user)
		exec(`git config user.email %v`, email)
	}

	chdir(repoPath)

	exec(`sudo apt-get install build-essential -y`)

	if tryexec(`latest -v`) != nil {
		exec(`go get -u github.com/mh-cbon/latest`)
	}
	if tryexec(`gh-api-cli -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/gh-api-cli")
	}

	tryexec(`git remote rm origin`)
	tryexec(`git remote add origin https://github.com/%v.git`, reposlug)
	tryexec(`git remote -vv`)
	tryexec(`git branch -aav`)

	tryexec("yes | git fetch")
	if tryexec(`git checkout gh-pages`) != nil {
		exec(`git checkout -b gh-pages`)
	}
	exec(`git reset HEAD --hard`)
	exec(`git clean -ffx`)
	exec(`git clean -ffd`)
	exec(`git clean -ffX`)
	exec(`git status`)

	tryexec(`ls -al`)

	aptDir := filepath.Join(repoPath, "apt")
	pkgDir := filepath.Join(repoPath, "pkg")
	aptlyDir := filepath.Join(repoPath, "aptly_0.9.7_linux_amd64")
	aptlyGz := filepath.Join(repoPath, "aptly_0.9.7_linux_amd64.tar.gz")
	aptlyBin := filepath.Join(aptlyDir, "aptly")
	aptlyConf := filepath.Join(repoPath, "aptly.conf")

	removeAll(aptDir)

	if _, s := os.Stat(aptlyDir); os.IsNotExist(s) {
		u := "https://bintray.com/artifact/download/smira/aptly/" + "aptly_0.9.7_linux_amd64.tar.gz"
		dlURL(u, aptlyGz)
		exec(`tar xzf ` + aptlyGz)
		removeAll(aptlyGz)
	}

	conf := `{
	  "rootDir": "` + repoPath + `/apt",
	  "downloadConcurrency": 4,
	  "downloadSpeedLimit": 0,
	  "architectures": [],
	  "dependencyFollowSuggests": false,
	  "dependencyFollowRecommends": false,
	  "dependencyFollowAllVariants": false,
	  "dependencyFollowSource": false,
	  "gpgDisableSign": true,
	  "gpgDisableVerify": true,
	  "downloadSourcePackages": false,
	  "ppaDistributorID": "",
	  "ppaCodename": ""
	}`
	writeFile(aptlyConf, conf)

	exec(`gh-api-cli dl-assets -t %q -o %v -r %v -g '*deb' -out '%v/%%r-%%v_%%a.deb'`, ghToken, user, name, pkgDir)

	mkdirAll(aptDir)
	chdir(aptDir)

	exec(`%v repo create -config=%v -distribution=all -component=main %v`, aptlyBin, aptlyConf, reposlug)
	exec(`%v repo add -config=%v %v %v`, aptlyBin, aptlyConf, reposlug, pkgDir)
	exec(`%v publish -component=contrib -config=%v repo %v`, aptlyBin, aptlyConf, reposlug)
	exec(`%v repo show -config=%v -with-packages %v`, aptlyBin, aptlyConf, reposlug)

	listFile := fmt.Sprintf(`%v.list`, name)
	listContent := fmt.Sprintf("deb [trusted=yes] https://%v.github.io/%v/apt/public/ all contrib\n", user, name)
	writeFile(listFile, listContent)

	chdir(repoPath)
	removeAll(aptlyGz)
	removeAll(aptlyGz + ".*") // handle aptly_0.9.7_linux_amd64.tar.gz.1
	removeAll(aptlyConf)
	removeAll(aptlyDir)
	removeAll(pkgDir)

	exec(`ls -al .`)
	exec(`ls -al apt`)

	tryexec(`git status`)

	fmt.Println("push", push)
	if push {
		exec(`git add -A`)
		exec(`git commit -m "debian repository"`)
		gU := fmt.Sprintf(`https://%v@github.com/%v.git`, ghToken, reposlug)
		exec(`git push --force --quiet %q gh-pages`, gU)
	}
}

func writeFile(f, content string) {
	fmt.Println("writeFile", f)
	err := ioutil.WriteFile(f, []byte(content), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func getURL(u string) []byte {
	response, err := http.Get(u)
	fmt.Println("getURL", u)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	var ret bytes.Buffer
	_, err = io.Copy(&ret, response.Body)
	if err != nil {
		panic(err)
	}
	return ret.Bytes()
}

func dlURL(u, to string) bool {
	fmt.Println("dlURL ", u)
	fmt.Println("to ", to)
	response, err := http.Get(u)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	f, err := os.Create(to)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, response.Body)
	if err != nil {
		panic(err)
	}
	return true
}
