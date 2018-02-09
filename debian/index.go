package debian

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mattn/go-zglob"
	"github.com/mh-cbon/go-bin-deb/stringexec"
	"github.com/mh-cbon/verbose"
)

var logger = verbose.Auto()

type vcsSrc struct {
	Type string `json:"type"` // Type identifier of the vcs source
	URL  string `json:"url"`  // Url-like to the vcs source
}

func (c vcsSrc) String() string {
	return c.Type + ": " + c.URL
}

type filesInstruction struct {
	From  string `json:"from"`  // Source path to the files
	Base  string `json:"base"`  // Base path to copy files from
	To    string `json:"to"`    // Target path to copy the files to
	Fperm string `json:"fperm"` // Permissions to apply such 0755
	Dperm string `json:"dperm"` // Permissions to apply such 0755
}
type copyright struct {
	Files     string `json:"files"`     // A pattern to describe a files selection
	Copyright string `json:"copyright"` // the text of the copyright
	License   string `json:"license"`   // License to apply to the selected files
	File      string `json:"file"`      // Path to the file containing the license content
}
type menu struct {
	Name            string `json:"name"`           // Name of the shortcut
	Description     string `json:"description"`    //
	GenericName     string `json:"generic-name"`   //
	Exec            string `json:"exec"`           // Exec command
	Icon            string `json:"icon"`           // Path to the installed icon
	Type            string `json:"type"`           // Type of shortcut
	StartupNotify   bool   `json:"startup-notify"` // yes/no
	Terminal        bool   `json:"terminal"`       // yes/no
	DBusActivatable bool   `json:"dbus-activable"` // yes/no
	NoDisplay       bool   `json:"no-display"`     // yes/no
	Keywords        string `json:"keywords"`       // ; separated list
	OnlyShowIn      string `json:"only-show-in"`   // ; separated list
	Categories      string `json:"categories"`     // ; separated list
	MimeType        string `json:"mime-type"`      // ; separated list
}

// Package contaisn informtation about a debian package to build
type Package struct {
	Name                string             `json:"name"`                 // Name of the package
	Maintainer          string             `json:"maintainer"`           // Information of the package maintainer
	Changedby           string             `json:"changed-by"`           // Information of the last package maintainer
	Section             string             `json:"section"`              // Classification of the application area
	Priority            string             `json:"priority"`             // Priority of the package (required,important,standard,optional,extra)
	Arch                string             `json:"arch"`                 // Arch targeted by the package
	Homepage            string             `json:"homepage"`             // Url to the homepage of the program
	SourcesURL          string             `json:"sources-url"`          // Url to the source of the program
	Version             string             `json:"version"`              // Version of the package
	Vcs                 []vcsSrc           `json:"vcs"`                  // Vcs information of the package
	Files               []filesInstruction `json:"files"`                // Files information to copy into the package
	CopyrightSpecURL    string             `json:"copyrights-spec-url"`  // Url to the copyright file specification
	Copyrights          []copyright        `json:"copyrights"`           // Copyrights of the package
	Essential           bool               `json:"essential"`            // Indicate if the package is an essential one
	Depends             []string           `json:"depends"`              // Dependency list
	Recommends          []string           `json:"recommends"`           // Recommendation list
	Suggests            []string           `json:"suggests"`             // Suggestion list
	Enhances            []string           `json:"enhances"`             // Enhancement list
	PreDepends          []string           `json:"pre-depends"`          // Pre-dependency list
	Breaks              []string           `json:"breaks"`               // Breaks list
	Conflicts           []string           `json:"conflits"`             // Conflicts list
	Envs                map[string]string  `json:"envs"`                 // Environment variables to define
	Provides            string             `json:"provides"`             // Provides
	Replaces            string             `json:"replaces"`             // Replaces
	BuiltUsing          string             `json:"built-using"`          // Built-using list
	Description         string             `json:"description"`          // A one-line short description
	DescriptionExtended string             `json:"description-extended"` // A multi-line long description
	PackageType         string             `json:"package-type"`         // Type of the package
	CronFiles           map[string]string  `json:"cron-files"`           // Cron files to use for the package
	CronCmds            map[string]string  `json:"cron-cmds"`            // Cron string to use to generate cron files for the package
	SystemdFile         string             `json:"systemd-file"`         // Systemd unit file
	InitFile            string             `json:"init-file"`            // Init file describing a service for the package
	DefaultFile         string             `json:"default-file"`         // Default init file describing a service for the package
	PreinstFile         string             `json:"preinst-file"`         // Pre-inst script path
	PostinstFile        string             `json:"postinst-file"`        // Post-inst script path
	PrermFile           string             `json:"prerm-file"`           // Pre-rm script path
	PostrmFile          string             `json:"postrm-file"`          // Post-rm script path
	Conffiles           []string           `json:"conf-files"`           // A list of the configuration files
	Mans                []string           `json:"mans"`                 // A list of man page in the package
	ChangelogFile       string             `json:"changelog-file"`       // Post-rm to the changelog file to copy to the package
	ChangelogCmd        string             `json:"changelog-cmd"`        // A cmd to run which generates the content of the changelog file
	Menus               []menu             `json:"menus"`                // Desktop shortcuts
}

// Load given deb.json file
func (d *Package) Load(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		m := fmt.Sprintf("json file '%s' does not exist: %s", file, err.Error())
		return errors.New(m)
	}
	byt, err := ioutil.ReadFile(file)
	if err != nil {
		m := fmt.Sprintf("error occured while reading file '%s': %s", file, err.Error())
		return errors.New(m)
	}
	if err := json.Unmarshal(byt, d); err != nil {
		m := fmt.Sprintf("Invalid json file '%s': %s", file, err.Error())
		return errors.New(m)
	}
	return nil
}

// Normalize current metadata
func (d *Package) Normalize(debianDir string, version string, arch string) {

	tokens := make(map[string]string)
	tokens["!version!"] = version
	tokens["!arch!"] = arch
	tokens["!name!"] = d.Name

	d.Version = replaceTokens(d.Version, tokens)
	d.Arch = replaceTokens(d.Arch, tokens)
	d.Homepage = replaceTokens(d.Homepage, tokens)
	d.SourcesURL = replaceTokens(d.SourcesURL, tokens)
	d.CopyrightSpecURL = replaceTokens(d.CopyrightSpecURL, tokens)
	d.Description = replaceTokens(d.Description, tokens)
	d.DescriptionExtended = replaceTokens(d.DescriptionExtended, tokens)
	d.InitFile = replaceTokens(d.InitFile, tokens)
	d.SystemdFile = replaceTokens(d.SystemdFile, tokens)
	d.DefaultFile = replaceTokens(d.DefaultFile, tokens)
	d.PreinstFile = replaceTokens(d.PreinstFile, tokens)
	d.PostinstFile = replaceTokens(d.PostinstFile, tokens)
	d.PrermFile = replaceTokens(d.PrermFile, tokens)
	d.PostrmFile = replaceTokens(d.PostrmFile, tokens)
	d.ChangelogFile = replaceTokens(d.ChangelogFile, tokens)
	d.ChangelogCmd = replaceTokens(d.ChangelogCmd, tokens)

	for i, v := range d.Vcs {
		d.Vcs[i].URL = replaceTokens(v.URL, tokens)
	}
	for i, v := range d.Files {
		d.Files[i].From = replaceTokens(v.From, tokens)
		d.Files[i].Base = replaceTokens(v.Base, tokens)
		d.Files[i].To = replaceTokens(v.To, tokens)
	}
	for i, v := range d.Copyrights {
		d.Copyrights[i].Files = replaceTokens(v.Files, tokens)
		d.Copyrights[i].Copyright = replaceTokens(v.Copyright, tokens)
		d.Copyrights[i].License = replaceTokens(v.License, tokens)
		d.Copyrights[i].File = replaceTokens(v.File, tokens)
	}
	for i, v := range d.CronFiles {
		d.CronFiles[i] = replaceTokens(v, tokens)
	}
	for i, v := range d.CronCmds {
		d.CronCmds[i] = replaceTokens(v, tokens)
	}
	for i, v := range d.Conffiles {
		d.Conffiles[i] = replaceTokens(v, tokens)
	}
	for i, v := range d.Mans {
		d.Mans[i] = replaceTokens(v, tokens)
	}
	for i, v := range d.Menus {
		d.Menus[i].Name = replaceTokens(v.Name, tokens)
		d.Menus[i].Description = replaceTokens(v.Description, tokens)
		d.Menus[i].GenericName = replaceTokens(v.GenericName, tokens)
		d.Menus[i].Exec = replaceTokens(v.Exec, tokens)
		d.Menus[i].Icon = replaceTokens(v.Icon, tokens)
	}

	if d.CopyrightSpecURL == "" {
		d.CopyrightSpecURL = "http://anonscm.debian.org/viewvc/dep/web/deps/dep5/copyright-format.xml?view=markup"
	}
	if d.Version == "" {
		d.Version = version
	}
	if d.Arch == "" {
		d.Arch = arch
	}
	if d.PackageType == "" {
		d.PackageType = "deb"
	}
	if d.Changedby == "" {
		d.Changedby = d.Maintainer
	}
	if d.Section == "" {
		d.Section = "unknown"
	}
	if d.Priority == "" {
		d.Priority = "extra"
	}

	if d.InitFile != "" && contains(d.Conffiles, d.InitFile) == false {
		d.Conffiles = append(d.Conffiles, filepath.Join("etc", "init.d", d.Name+".sh"))
	}
	if d.DefaultFile != "" && contains(d.Conffiles, d.DefaultFile) == false {
		d.Conffiles = append(d.Conffiles, filepath.Join("etc", "default", d.Name+".sh"))
	}

	if len(d.Envs) > 0 {
		d.Conffiles = append(d.Conffiles, filepath.Join("etc", "profile.d", d.Name+".sh"))
	}
}

func replaceTokens(in string, tokens map[string]string) string {
	for token, v := range tokens {
		in = strings.Replace(in, token, v, -1)
	}
	return in
}

// GenerateFiles from sourceDir to pkgDir
func (d *Package) GenerateFiles(sourceDir string, pkgDir string) error {

	dataDir := filepath.Join(pkgDir, "debian")
	debianDir := filepath.Join(pkgDir, "debian", "DEBIAN")
	// create the base structure
	if err := os.MkdirAll(filepath.Join(debianDir), 0755); err != nil {
		return err
	}
	logger.Printf("base structure created: %s", debianDir)

	if err := os.MkdirAll(filepath.Join(dataDir), 0755); err != nil {
		return err
	}
	logger.Printf("data dir created: %s", dataDir)

	// copy the files
	if err := d.ImportFiles(dataDir); err != nil {
		m := fmt.Sprintf("Could not copy the files: %s", err.Error())
		return errors.New(m)
	}
	logger.Println("files structure created")

	// generate shortcuts
	if err := d.WriteShortcuts(dataDir); err != nil {
		m := fmt.Sprintf("Could not generate shortcuts: %s", err.Error())
		return errors.New(m)
	}
	logger.Println("shortcuts created")

	// generate env variable file
	if err := d.WriteEnvProfile(dataDir); err != nil {
		m := fmt.Sprintf("Could not generate etc/profile.d/%s.sh: %s", d.Name, err.Error())
		return errors.New(m)
	}
	logger.Println("env file created")

	// generate the init file
	if err := d.WriteInitFile(dataDir); err != nil {
		m := fmt.Sprintf("Could not generate init file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("init file created\n")

	// generate the unit file
	if err := d.WriteUnitFile(dataDir); err != nil {
		m := fmt.Sprintf("Could not generate unit file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("unit file created\n")

	// generate the default init file
	if err := d.WriteDefaultInitFile(dataDir); err != nil {
		m := fmt.Sprintf("Could not generate default init file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("init default file created\n")

	// generate the conffiles file
	if err := d.WriteConffiles(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate conffiles file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("conffiles file created\n")

	// compute file size
	var size uint64
	s, err := d.ComputeSize(pkgDir)
	if err != nil {
		m := fmt.Sprintf("Could not compute install size: %s", err.Error())
		return errors.New(m)
	}
	size = uint64(s)
	logger.Printf("size=%d\n", size)

	// generate the control file
	if err := d.WriteControlFile(debianDir, uint64(size)); err != nil {
		m := fmt.Sprintf("Could not generate control file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("control file created\n")

	// generate the changelog file
	//  /usr/share/doc/pkg/
	pkgDoc := filepath.Join(dataDir, "usr", "share", "doc", d.Name)
	if err := d.WriteChangelogFile(pkgDoc); err != nil {
		m := fmt.Sprintf("Could not generate changelog file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("changelog file created\n")

	// generate the copyright file
	if err := d.WriteCopyrightFile(pkgDoc); err != nil {
		m := fmt.Sprintf("Could not generate copyright file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("copyright file created\n")

	// generate the cron files
	if err := d.WriteCronFiles(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate cron files: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("cron files created\n")

	// generate the preinst file
	if err := d.WritePreInstFile(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate preinst file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("preinst file created\n")

	// generate the postinst file
	if err := d.WritePostInstFile(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate postinst file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("postinst file created\n")

	// generate the prerm file
	if err := d.WritePreRmFile(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate prerm file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("prerm file created\n")

	// generate the postrm file
	if err := d.WritePostRmFile(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate postrm file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("postrm file created\n")

	// generate the manpage index file
	if err := d.WriteManPageIndexFile(debianDir); err != nil {
		m := fmt.Sprintf("Could not generate man pages index file: %s", err.Error())
		return errors.New(m)
	}
	logger.Printf("man pages index file created\n")

	return nil
}

// GenerateInstall generates install file.
func (d *Package) GenerateInstall(sourceDir string, debianDir string, dataDir string) error {
	var err error
	content := ""
	if sourceDir, err = filepath.Abs(sourceDir); err != nil {
		return err
	}
	for _, fileInst := range d.Files {
		from := fileInst.From
		to := fileInst.To
		base := fileInst.Base

		if filepath.IsAbs(from) == false {
			from = filepath.Join(sourceDir, from)
		}
		if filepath.IsAbs(to) { // to must not be absolute..
			to = to[1:]
		}
		if filepath.IsAbs(base) == false {
			base = filepath.Join(sourceDir, base)
		}

		logger.Printf("fileInst.From=%q\n", from)
		logger.Printf("fileInst.To=%q\n", to)
		logger.Printf("fileInst.Base=%q\n", base)

		items, err := zglob.Glob(from)
		if err != nil {
			m := fmt.Sprintf("Could not glob files source '%s': %s", from, err.Error())
			return errors.New(m)
		}
		logger.Printf("items=%q\n", items)
		for _, item := range items {
			n := item
			if len(item) >= len(base) && item[0:len(base)] == base {
				n = item[len(base):]
			}
			n = filepath.Join(to, n)
			content += fmt.Sprintf("%s %s\n", item, filepath.Dir(n))
		}
	}
	for _, m := range d.Menus {
		file := filepath.Join(dataDir, m.Name+".desktop")
		icon := m.Icon
		if filepath.IsAbs(icon) == false {
			icon = filepath.Join(sourceDir, icon)
		}
		content += fmt.Sprintf("%s %s\n", file, "/usr/share/applications/"+d.Name+".desktop")
		content += fmt.Sprintf("%s %s\n", icon, "/usr/share/pixmaps/"+filepath.Base(icon))
	}
	logger.Printf("content=\n%s\n", content)

	f := filepath.Join(debianDir, "install")
	return ioutil.WriteFile(f, []byte(content), 0644)
}

// WriteConffiles updates the debian directory
func (d *Package) WriteConffiles(debianDir string) error {
	content := ""
	for _, f := range d.Conffiles {
		if filepath.IsAbs(f) == false {
			f = "/" + f // lintian says: must not be rel.
		}
		content += fmt.Sprintf("%s\n", f)
	}
	if content == "" {
		return nil
	}
	f := filepath.Join(debianDir, "conffiles")
	return ioutil.WriteFile(f, []byte(content), 0644)
}

// WriteEnvProfile generates an etc/profile.d/plg.name.sh
func (d *Package) WriteEnvProfile(debianDir string) error {
	content := ""
	for k, v := range d.Envs {
		content += fmt.Sprintf("%s=%q\n", k, v)
	}
	if content == "" {
		return nil
	}
	f := filepath.Join(debianDir, "etc", "profile.d", d.Name+".sh")
	if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(f, []byte(content), 0644)
}

// ImportFiles add files to the package.
func (d *Package) ImportFiles(sourceDir string) error {
	for _, fileInst := range d.Files {
		var fperm int32
		var dperm int32 = 0755
		if fileInst.Fperm != "" {
			p, err := strconv.ParseInt(fileInst.Fperm, 8, 32)
			if err != nil {
				return err
			}
			fperm = int32(p)
		}
		if fileInst.Dperm != "" {
			p, err := strconv.ParseInt(fileInst.Dperm, 8, 32)
			if err != nil {
				return err
			}
			dperm = int32(p)
		}
		items, err := zglob.Glob(fileInst.From)
		logger.Printf("fileInst.From=%q\n", fileInst.From)
		logger.Printf("fileInst.To=%q\n", fileInst.To)
		logger.Printf("fileInst.Base=%q\n", fileInst.Base)
		logger.Printf("fileInst.Fperm=%q\n", fileInst.Fperm)
		logger.Printf("fileInst.fperm=%#o\n", fperm)
		logger.Printf("fileInst.Dperm=%q\n", fileInst.Dperm)
		logger.Printf("fileInst.dperm=%#o\n", dperm)
		if err != nil {
			m := fmt.Sprintf("Could not glob files source '%s': %s", fileInst.From, err.Error())
			return errors.New(m)
		}
		logger.Printf("items=%q\n", items)
		targetItems := make([]string, 0)
		for _, item := range items {
			n := item[len(fileInst.Base):]
			n = filepath.Join(sourceDir, fileInst.To, n)
			targetItems = append(targetItems, n)
		}
		logger.Printf("targetItems=%q\n", targetItems)
		for i, item := range items {
			s, err := os.Stat(item)
			if err != nil {
				m := fmt.Sprintf("Could not stat source file '%s': %s", item, err.Error())
				return errors.New(m)
			}
			if s.IsDir() {
				if err := os.MkdirAll(targetItems[i], os.FileMode(dperm)); err != nil {
					m := fmt.Sprintf("Could not create directory file '%s': %s", targetItems[i], err.Error())
					return errors.New(m)
				}
			} else {
				d := filepath.Dir(targetItems[i])
				if err := os.MkdirAll(d, os.FileMode(dperm)); err != nil {
					m := fmt.Sprintf("Could not create directory file '%s': %s", d, err.Error())
					return errors.New(m)
				}
			}
		}
		for i, item := range items {
			s, err := os.Stat(item)
			if err != nil {
				m := fmt.Sprintf("Could not stat source file '%s': %s", item, err.Error())
				return errors.New(m)
			}
			if s.IsDir() == false {
				if err := cp(targetItems[i], item); err != nil {
					m := fmt.Sprintf("Could not copy file from '%s' to '%s': %s", item, targetItems[i], err.Error())
					return errors.New(m)
				}
				if fileInst.Fperm != "" {
					if err := os.Chmod(targetItems[i], os.FileMode(fperm)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ComputeSize returns size of a directory
func (d *Package) ComputeSize(sourceDir string) (int64, error) {
	var size int64

	// exclDir := filepath.Join(sourceDir)
	walkFn := func(path string, info os.FileInfo, err error) error {
		// if len(path)>=len(exclDir) && path[0:len(exclDir)]==exclDir {
		//   return nil
		// }
		if err == nil {
			s, _ := os.Stat(path)
			if err == nil {
				if s.IsDir() == false {
					size += s.Size()
				}
			}
		}
		return nil
	}
	err := filepath.Walk(sourceDir, walkFn)
	if err != nil {
		return size, err
	}
	return size / 1024, nil
}

// WriteControlFile writes the control file.
func (d *Package) WriteControlFile(debianDir string, size uint64) error {

	desc := d.Description
	if d.DescriptionExtended != "" {
		desc += "\n"
		for _, line := range strings.Split(d.DescriptionExtended, "\n") {
			desc += " " + line + "\n"
		}
		desc = strings.TrimSpace(desc)
	}

	arch := d.Arch
	if arch == "386" { // go style
		arch = "i386" // deb style
	}

	P := ""
	P += strAppend("Package", d.Name)
	P += strAppend("Version", d.Version)
	P += strAppend("Source", d.Name)
	P += strAppend("Section", d.Section)
	P += strAppend("Priority", d.Priority)
	P += strAppend("Maintainer", d.Maintainer)
	P += strAppend("Homepage", d.Homepage)
	P += strAppend("Description", desc)
	P += strAppend("Architecture", arch)
	P += vcsSliceAppend(d.Vcs)
	P += boolAppend("Essential", d.Essential)
	P += strSliceAppend("Depends", d.Depends, ",")
	P += strSliceAppend("Recommends", d.Recommends, ",")
	P += strSliceAppend("Suggests", d.Suggests, ",")
	P += strSliceAppend("Enhances", d.Enhances, ",")
	P += strSliceAppend("Pre-Depends", d.PreDepends, ",")
	P += strSliceAppend("Breaks", d.Breaks, ",")
	P += strSliceAppend("Conflicts", d.Conflicts, ",")
	P += strAppend("Provides", d.Provides)
	P += strAppend("Replaces", d.Replaces)
	P += strAppend("Built-Using", d.BuiltUsing)
	P += strAppend("Installed-Size", strconv.FormatUint(size,10))
	P += strAppend("Package-Type", d.PackageType)

	controlContent := []byte(P)
	control := filepath.Join(debianDir, "control")

	return ioutil.WriteFile(control, controlContent, 0644)
}

// WriteCopyrightFile writes the copyright file.
func (d *Package) WriteCopyrightFile(debianDir string) error {
	if err := os.MkdirAll(debianDir, 0755); err != nil {
		return err
	}
	content := ""
	content += strAppend("Format-Specification", d.CopyrightSpecURL)
	content += strAppend("Name", d.Name)
	content += strAppend("Maintainer", d.Maintainer)
	sourcesURL := d.SourcesURL
	if sourcesURL == "" {
		sourcesURL = d.Homepage
	}
	content += strAppend("Source", sourcesURL)
	content += "\n"

	// write the copyrights
	for _, c := range d.Copyrights {
		t := ""
		if c.Files != "" {
			t += strAppend("Files", c.Files)
		}
		if c.Copyright != "" {
			t += strAppend("Copyright", c.Copyright)
		}
		if c.License != "" {
			t += strAppend("License", c.License)
		}
		if c.File != "" {
			t += strAppend("File", c.File)
		}
		if t != "" {
			content += t + "\n"
		}
	}
	file := filepath.Join(debianDir, "copyright")
	return ioutil.WriteFile(file, []byte(content), 0644)
}

// WriteCronFiles writes the cron file.
func (d *Package) WriteCronFiles(debianDir string) error {

	for k, val := range d.CronFiles {
		if val != "" {
			file := filepath.Join(debianDir, d.Name+"-f.cron."+k)
			if err := cp(file, val); err != nil {
				m := fmt.Sprintf("Failed to copy '%s' to '%s': %s", val, file, err.Error())
				return errors.New(m)
			}
		} else {
			fmt.Printf("cron job '%s' is empty!\n", k)
		}
	}
	for k, content := range d.CronCmds {
		if content != "" {
			file := filepath.Join(debianDir, d.Name+"-c.cron."+k)
			if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteChangelogFile writes the changelog file.
func (d *Package) WriteChangelogFile(debianDir string) error {
	if err := os.MkdirAll(debianDir, 0755); err != nil {
		return err
	}

	file := filepath.Join(debianDir, "changelog")
	if d.ChangelogFile != "" {
		if err := cp(file, d.ChangelogFile); err != nil {
			return err
		}
	} else if d.ChangelogCmd != "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		cmd, err := stringexec.Command(wd, d.ChangelogCmd)
		if err != nil {
			return err
		}
		cmd.Stdout = nil
		content, err := cmd.Output()
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	} else {
		ioutil.WriteFile(file, []byte(""), 0644) // force empty changelog
	}

	oCmd := exec.Command("gzip", "--best", file)
	oCmd.Stdout = os.Stdout
	oCmd.Stderr = os.Stderr
	return oCmd.Run()
}

// WriteManPageIndexFile writes the map page index file.
func (d *Package) WriteManPageIndexFile(debianDir string) error {

	file := filepath.Join(debianDir, d.Name+".manpages")
	if len(d.Mans) > 0 {
		content := strings.Join(d.Mans, "\n")
		if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// WriteShortcuts writes the application shortcuts.
func (d *Package) WriteShortcuts(dataDir string) error {

	for _, m := range d.Menus {
		s := ""

		if m.Name != "" {
			s += fmt.Sprintf("%s=%s\n", "Name", m.Name)
		}

		if m.Description != "" {
			s += fmt.Sprintf("%s=%s\n", "Description", m.Description)
		}

		if m.GenericName != "" {
			s += fmt.Sprintf("%s=%s\n", "GenericName", m.GenericName)
		}

		if m.Exec != "" {
			s += fmt.Sprintf("%s=%s\n", "Exec", m.Exec)
		}

		if m.Icon != "" {
			s += fmt.Sprintf("%s=%s\n", "Icon", "/usr/share/pixmaps/"+filepath.Base(m.Icon))
		}

		if m.Type != "" {
			s += fmt.Sprintf("%s=%s\n", "Type", m.Type)
		}

		if m.Categories != "" {
			s += fmt.Sprintf("%s=%s\n", "Categories", m.Categories)
		}

		if m.MimeType != "" {
			s += fmt.Sprintf("%s=%s\n", "MimeType", m.MimeType)
		}

		if m.OnlyShowIn != "" {
			s += fmt.Sprintf("%s=%s\n", "OnlyShowIn", m.OnlyShowIn)
		}

		if m.Keywords != "" {
			s += fmt.Sprintf("%s=%s\n", "Keywords", m.Keywords)
		}

		if s != "" {

			if m.StartupNotify {
				s += fmt.Sprintf("%s=%s\n", "StartupNotify", "true")
			} else {
				s += fmt.Sprintf("%s=%s\n", "StartupNotify", "false")
			}

			if m.DBusActivatable {
				s += fmt.Sprintf("%s=%s\n", "DBusActivatable", "true")
			} else {
				s += fmt.Sprintf("%s=%s\n", "DBusActivatable", "false")
			}

			if m.NoDisplay {
				s += fmt.Sprintf("%s=%s\n", "NoDisplay", "true")
			} else {
				s += fmt.Sprintf("%s=%s\n", "NoDisplay", "false")
			}

			if m.Terminal {
				s += fmt.Sprintf("%s=%s\n", "Terminal", "true")
			} else {
				s += fmt.Sprintf("%s=%s\n", "Terminal", "false")
			}

			s = "[Desktop Entry]\n" + s

			file := filepath.Join(dataDir, "usr", "share", "applications", m.Name+".desktop")
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				return err
			}
			if err := ioutil.WriteFile(file, []byte(s), 0644); err != nil {
				return err
			}

			icoFile := filepath.Join(dataDir, "usr", "share", "pixmaps", filepath.Base(m.Icon))
			if err := os.MkdirAll(filepath.Dir(icoFile), 0755); err != nil {
				return err
			}
			if err := cp(icoFile, m.Icon); err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteUnitFile writes the unit.d file.
func (d *Package) WriteUnitFile(dataDir string) error {
	if d.SystemdFile != "" {
		f := filepath.Join(dataDir, "lib", "systemd", "system", filepath.Base(d.SystemdFile))
		if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
			return err
		}
		if err := writeAFile(f, d.SystemdFile); err != nil {
			return err
		}
		return os.Chmod(f, 0644)
	}
	return nil
}

// WriteInitFile writes the etc/init.d file.
func (d *Package) WriteInitFile(dataDir string) error {
	if d.InitFile != "" {
		f := filepath.Join(dataDir, "etc", "init.d", d.Name+".sh")
		if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
			return err
		}
		if err := writeAFile(f, d.InitFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// WriteDefaultInitFile writes the etc/default file.
func (d *Package) WriteDefaultInitFile(dataDir string) error {
	if d.DefaultFile != "" {
		f := filepath.Join(dataDir, "etc", "default", d.Name+".sh")
		if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
			return err
		}
		if err := writeAFile(f, d.DefaultFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// WritePreInstFile writes the preinst file.
func (d *Package) WritePreInstFile(debianDir string) error {
	if d.PreinstFile != "" {
		f := filepath.Join(debianDir, "preinst")
		if err := writeAFile(f, d.PreinstFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// WritePostInstFile writes the postinst file.
func (d *Package) WritePostInstFile(debianDir string) error {
	if d.PostinstFile != "" {
		f := filepath.Join(debianDir, "postinst")
		if err := writeAFile(f, d.PostinstFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// WritePreRmFile writes the prerm file.
func (d *Package) WritePreRmFile(debianDir string) error {
	if d.PrermFile != "" {
		f := filepath.Join(debianDir, "prerm")
		if err := writeAFile(f, d.PrermFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// WritePostRmFile writes the postrm file.
func (d *Package) WritePostRmFile(debianDir string) error {
	if d.PostrmFile != "" {
		f := filepath.Join(debianDir, "postrm")
		if err := writeAFile(f, d.PostrmFile); err != nil {
			return err
		}
		return os.Chmod(f, 0755)
	}
	return nil
}

// CopyResults copy the packages to the path..
func (d *Package) CopyResults(from string, to string) error {
	items, err := zglob.Glob(from + "/" + d.Name + "*")
	if err != nil {
		return err
	}
	for _, item := range items {
		b := filepath.Base(item)
		if err := cp(filepath.Join(to, b), item); err != nil {
			return err
		}
	}
	return nil
}

func writeAFile(dst string, src string) error {
	if len(src) > 0 {
		err := cp(dst, src)
		if err != nil {
			m := fmt.Sprintf("Could not write file '%s': %s", src, err.Error())
			return errors.New(m)
		}
	}
	return nil
}

func strAppend(name string, value string) string {
	ret := ""
	if len(value) > 0 {
		ret = fmt.Sprintf("%s: %s\n", name, value)
	}
	return ret
}
func boolAppend(name string, value bool) string {
	ret := ""
	if value {
		ret = fmt.Sprintf("%s: %s\n", name, "yes")
	}
	return ret
}
func vcsSliceAppend(s []vcsSrc) string {
	ret := ""
	for _, k := range s {
		ret += structAppend(k)
	}
	return ret
}
func structAppend(s fmt.Stringer) string {
	ret := ""
	value := s.String()
	if len(value) > 0 {
		ret = fmt.Sprintf("%s\n", value)
	}
	return ret
}
func strSliceAppend(name string, value []string, j string) string {
	ret := ""
	if len(value) > 0 {
		ret = fmt.Sprintf("%s: %s\n", name, strings.Join(value, j))
	}
	return ret
}
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func contains(s []string, v string) bool {
	for _, k := range s {
		if v == k {
			return true
		}
	}
	return false
}
