package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// SetupPpa creates a debian repository
func SetupPpa(reposlug, ghToken, email, version, archs, srepos, outbuild string, push, keep bool) {

	repos := []string{}
	for _, repo := range strings.Split(srepos, ",") {
		if strings.TrimSpace(repo) == "" {
			log.Printf("Ignored repo %q\n", repo)
		} else {
			repos = append(repos, repo)
		}
	}
	if len(repos) < 1 {
		fmt.Println("-repo argument is required (example: user/repo1, user/repo2)")
		os.Exit(1)
	}

	x := strings.Split(reposlug, "/")
	user := x[0]
	name := x[1]

	gopath := os.Getenv("GOPATH")
	repoPath := filepath.Join(gopath, "src", "github.com", reposlug)
	fmt.Println("repoPath", repoPath)

	setupGitRepo(repoPath, reposlug, user, email)
	chdir(repoPath)

	maybesudo(`apt-get install build-essential -y --quiet`)

	if tryexec(`latest -v`) != nil {
		exec(`git clone https://github.com/mh-cbon/latest.git %v/src/github.com/mh-cbon/latest`, gopath)
		exec(`go install github.com/mh-cbon/latest`)
	}
	if tryexec(`gh-api-cli -v`) != nil {
		exec(`latest -repo=%v`, "mh-cbon/gh-api-cli")
	}

	resetGit(repoPath)
	tryexec(`git remote -vv`)
	tryexec(`git branch -aav`)
	getBranchGit(repoPath, reposlug, "gh-pages", "deborigin")
	tryexec(`git remote -vv`)
	tryexec(`git branch -aav`)
	resetGit(repoPath)
	exec(`git status`)

	tryexec(`ls -al`)

	aptlyDir := filepath.Join(repoPath, "aptly_0.9.7_linux_amd64")
	aptlyGz := filepath.Join(repoPath, "aptly_0.9.7_linux_amd64.tar.gz")
	aptlyBin := filepath.Join(aptlyDir, "aptly")
	aptlyConf := filepath.Join(repoPath, "aptly.conf")

	if _, s := os.Stat(aptlyDir); os.IsNotExist(s) {
		u := "https://bintray.com/artifact/download/smira/aptly/" + "aptly_0.9.7_linux_amd64.tar.gz"
		dlURL(u, aptlyGz)
		exec(`tar xzf ` + aptlyGz)
		removeAll(aptlyGz)
	}

	conf := `{
	  "rootDir": "` + outbuild + `",
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

	t := make(chan string, 2)
	d := make(chan bool)
	go func() {
		for cmd := range t {
			go func(c string) {
				exec(c)
				d <- true
			}(cmd)
		}
	}()

	outP := "%%r-%%v_%%a.deb"
	for _, repo := range repos {
		y := strings.Split(strings.TrimSpace(repo), "/")
		t <- fmt.Sprintf(`gh-api-cli dl-assets -t %q -o %v -r %v -g '*deb' -out '%v/%v'`, ghToken, y[0], y[1], outbuild, outP)
	}
	close(t)
	for range repos {
		<-d
	}
	close(d)

	mkdirAll(outbuild)
	chdir(outbuild)

	exec(`%v repo create -config=%v -distribution=all -component=main %v`, aptlyBin, aptlyConf, reposlug)
	exec(`%v repo add -config=%v %v %v`, aptlyBin, aptlyConf, reposlug, outbuild)
	exec(`%v publish -component=contrib -config=%v repo %v`, aptlyBin, aptlyConf, reposlug)
	exec(`%v repo show -config=%v -with-packages %v`, aptlyBin, aptlyConf, reposlug)

	listFile := fmt.Sprintf(`%v/%v.list`, outbuild, name)
	listContent := fmt.Sprintf("deb [trusted=yes] https://%v.github.io/%v/%v/public/ all contrib\n", user, name, filepath.Base(outbuild))
	writeFile(listFile, listContent)
	exec(`rm -f %v/*.deb`, outbuild)

	chdir(repoPath)
	removeAll(aptlyGz)
	removeAll(aptlyGz + ".*") // handle aptly_0.9.7_linux_amd64.tar.gz.1
	removeAll(aptlyConf)
	removeAll(aptlyDir)

	tryexec(`git status`)

	fmt.Println("push", push)
	if push {
		commitPushGit(repoPath, ghToken, reposlug, "gh-pages", "debian repository")
		if keep == false {
			removeAll(outbuild)
		}
	}
}
