package debian

import (
	"fmt"
	"os"
	"io"
	"strings"
	"errors"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/mh-cbon/go-bin-deb/stringexec"
	"github.com/mh-cbon/verbose"
  "github.com/dustin/go-humanize"
  "github.com/mattn/go-zglob"
)

var logger = verbose.Auto()

type VcsSrc struct {
	Type string `json:"type"` // Type identifier of the vcs source
	Url  string `json:"url"`  // Url-like to the vcs source
}

func (c VcsSrc) String() string {
	return c.Type + ": " + c.Url
}

type FilesInstruction struct {
	From string `json:"from"` // Source path to the files
	Base string `json:"base"` // Base path to copy files from
	To   string `json:"to"`   // Target path to copy the files to
}
type Copyright struct {
	Files     string `json:"files"`     // A pattern to describe a files selection
	Copyright string `json:"copyright"` // the text of the copyright
	License   string `json:"license"`   // License to apply to the selected files
	File      string `json:"file"`      // Path to the file containing the license content
}
type Menu struct {
	Name          string `json:"name"`           // Name of the shortcut
	Description   string `json:"description"`    //
	GenericName   string `json:"generic-name"`   //
	Exec          string `json:"exec"`           // Exec command
	Icon          string `json:"icon"`           // Path to the installed icon
	Type          string `json:"type"`           // Type of shortcut
	StartupNotify bool   `json:"startup-notify"` // yes/no
	Categories    string `json:"categories"`     // ; separated list
	MimeType      string `json:"mime-type"`      // ; separated list
}

type Package struct {
	Name                string             `json:"name"`                    // Name of the package
	Maintainer          string             `json:"maintainer"`              // Information of the package maintainer
	Uploaders           []string           `json:"uploaders"`               // Name of the package maintainer
	Changedby           string             `json:"changed-by"`              // Information of the last package maintainer
	Section             string             `json:"section"`                 // Classification of the application area
	Priority            string             `json:"priority"`                // Priority of the package (required,important,standard,optional,extra)
	Arch                string             `json:"arch"`                    // Arch targeted by the package
	StandardsVersion    string             `json:"standards-version"`       // Version of the standards the package complies with
	Homepage            string             `json:"homepage"`                // Url to the homepage of the program
	SourcesUrl          string             `json:"sources-url"`             // Url to the source of the program
	Version             string             `json:"version"`                 // Version of the package
	Vcs                 []VcsSrc           `json:"vcs"`                     // Version of the package
	Files               []FilesInstruction `json:"files"`                   // Files information to copy into the package
	CopyrightSpecUrl    string             `json:"copyrights-spec-url"`     // Url to the copyright file specification
	Copyrights          []Copyright        `json:"copyrights"`              // Version of the package
	Essential           bool               `json:"essential"`               // Indicate if the package is an essential one
	Depends             []string           `json:"depends"`                 // Dependency list
	Recommends          []string           `json:"recommends"`              // Recommendation list
	Suggests            []string           `json:"suggests"`                // Suggestion list
	Enhances            []string           `json:"enhances"`                // Enhancement list
	PreDepends          []string           `json:"pre-depends"`             // Pre-dependency list
	Breaks              []string           `json:"breaks"`                  // Breaks list
	Conflicts           []string           `json:"conflits"`                // Conflicts list
	Provides            string             `json:"provides"`                // Provides
	Replaces            string             `json:"replaces"`                // Replaces
	BuildDepends        []string           `json:"build-depends"`           // Build-dependency list
	BuildDependsIndep   []string           `json:"build-dependends-indeps"` // Build-dependency-indep list
	BuildConflicts      []string           `json:"build-conflicts"`         // Build-conflicts list
	BuildConflictsIndep []string           `json:"build-conflicts-indeps"`  // Build-conflicts-indep list
	BuiltUsing          string             `json:"built-using"`             // Built-using list
	Description         string             `json:"description"`             // A one-line short description
	DescriptionExtended string             `json:"description-extended"`    // A multi-line long description
	PackageType         string             `json:"package-type"`            // Type of the package
	Compat              string             `json:"compat"`                  // Compatibility version of the package
	CronFiles           map[string]string  `json:"cron-files"`              // Cron files to use for the package
	CronCmds            map[string]string  `json:"cron-cmds"`               // Cron string to use to generate cron files for the package
	InitFile            string             `json:"init-file"`               // Init file describing a service for the package
	DefaultFile         string             `json:"default-file"`            // Default init file describing a service for the package
	PreinstFile         string             `json:"preinst-file"`            // Pre-inst script path
	PostinstFile        string             `json:"postinst-file"`           // Post-inst script path
	PrermFile           string             `json:"prerm-file"`              // Pre-rm script path
	PostrmFile          string             `json:"postrm-file"`             // Post-rm script path
	RulesFile           string             `json:"rules-file"`              // rules script path
	Mans                []string           `json:"mans"`                    // A list of man page in the package
	ChangelogFile       string             `json:"changelog-file"`          // Post-rm to the changelog file to copy to the package
	ChangelogCmd        string             `json:"changelog-cmd"`           // A cmd to run which generates the content of the changelog file
	Menus               []Menu             `json:"menus"`                   // Desktop shortcuts
}

func (d *Package) Load (file string) error {
  if _, err := os.Stat(file); os.IsNotExist(err) {
    m := fmt.Sprintf("json file '%s' does not exist: %s", file, err.Error())
  	return errors.New(m)
  }
  byt, err := ioutil.ReadFile(file)
  if err!=nil {
    m := fmt.Sprintf("error occured while reading file '%s': %s", file, err.Error())
    return errors.New(m)
  }
  if err := json.Unmarshal(byt, d); err != nil {
    m := fmt.Sprintf("Invalid json file '%s': %s", file, err.Error())
    return errors.New(m)
  }
  return nil
}

func (d *Package) Normalize (sourceDir string) {
  if d.Compat=="" {
    d.Compat = "9"
  }
  if d.CopyrightSpecUrl == "" {
    d.CopyrightSpecUrl = "http://anonscm.debian.org/viewvc/dep/web/deps/dep5/copyright-format.xml?view=markup"
  }
  if d.StandardsVersion == "" {
    d.StandardsVersion = "3.9.6"
  }
  if d.PackageType == "" {
    d.PackageType = "deb"
  }
  if d.Changedby == "" {
    d.Changedby = d.Maintainer
  }
}

func (d *Package) GenerateFiles (sourceDir string) error {

  // create the base structure
  if err := os.MkdirAll(filepath.Join(sourceDir, "DEBIAN"), 0755); err!=nil {
    return err
  }
  logger.Println("base structure created")

  // copy the files
  if err := d.ImportFiles(sourceDir); err != nil {
    m := fmt.Sprintf("Could not copy the files: %s", err.Error())
    return errors.New(m)
  }
  logger.Println("files structure created")

  // generate shortcuts
  if err := d.WriteShortcuts(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate shortcuts: %s", err.Error())
    return errors.New(m)
  }
  logger.Println("shortcuts created")

  // compute file size
  var size uint64
  if s, err := d.ComputeSize(sourceDir); err != nil {
    m := fmt.Sprintf("Could not compute install size: %s", err.Error())
    return errors.New(m)
  } else {
    size = uint64(s)
  }
  logger.Printf("size=%q\n", size)

  // generate the control file
  if err := d.WriteControlFile(sourceDir, uint64(size)); err != nil {
    m := fmt.Sprintf("Could not generate control file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("control file created\n")

  // generate the copyright file
  if err := d.WriteCopyrightFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate copyright file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("copyright file created\n")

  // generate the compat file
  if err := d.WriteCompatFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate compat file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("compat file created\n")

  // generate the cron files
  if err := d.WriteCronFiles(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate cron files: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("cron files created\n")

  // generate the init file
  if err := d.WriteInitFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate init file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("init file created\n")

  // generate the default init file
  if err := d.WriteDefaultInitFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate default init file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("init default file created\n")

  // generate the preinst file
  if err := d.WritePreInstFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate preinst file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("preinst file created\n")

  // generate the postinst file
  if err := d.WritePostInstFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate postinst file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("postinst file created\n")

  // generate the prerm file
  if err := d.WritePreRmFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate prerm file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("prerm file created\n")

  // generate the postrm file
  if err := d.WritePostRmFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate postrm file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("postrm file created\n")

  // generate the rules file
  if err := d.WriteRulesFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate rules file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("rules file created\n")

  // generate the changelog file
  if err := d.WriteChangelogFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate changelog file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("changelog file created\n")

  // generate the manpage index file
  if err := d.WriteManPageIndexFile(sourceDir); err != nil {
    m := fmt.Sprintf("Could not generate man pages index file: %s", err.Error())
    return errors.New(m)
  }
  logger.Printf("man pages index file created\n")

  return nil
}

func (d *Package) ImportFiles (sourceDir string) error {
  for _, fileInst := range d.Files {
    items, err := zglob.Glob(fileInst.From)
    logger.Printf("fileInst.From=%q\n", fileInst.From)
    logger.Printf("fileInst.To=%q\n", fileInst.To)
    logger.Printf("fileInst.Base=%q\n", fileInst.Base)
    if err!=nil {
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
      if err!=nil {
        m := fmt.Sprintf("Could not stat source file '%s': %s", item, err.Error())
        return errors.New(m)
      }
      if s.IsDir() {
        err := os.MkdirAll(targetItems[i], 0755)
        if err!=nil {
          m := fmt.Sprintf("Could not create directory file '%s': %s", targetItems[i], err.Error())
          return errors.New(m)
        }
      } else {
        d := filepath.Dir(targetItems[i])
        err := os.MkdirAll(d, 0755)
        if err!=nil {
          m := fmt.Sprintf("Could not create directory file '%s': %s", d, err.Error())
          return errors.New(m)
        }
      }
    }
    for i, item := range items {
      s, err := os.Stat(item)
      if err!=nil {
        m := fmt.Sprintf("Could not stat source file '%s': %s", item, err.Error())
        return errors.New(m)
      }
      if s.IsDir()==false {
        err := cp(targetItems[i], item)
        if err!=nil {
          m := fmt.Sprintf("Could not copy file from '%s' to '%s': %s", item, targetItems[i], err.Error())
          return errors.New(m)
        }
      }
    }
  }
  return nil
}

func (d *Package) ComputeSize (sourceDir string) (int64, error) {
  var size int64 = 0

  exclDir := filepath.Join(sourceDir, "DEBIAN")
  walkFn := func (path string, info os.FileInfo, err error) error {
    if len(path)>=len(exclDir) && path[0:len(exclDir)]==exclDir {
      return nil
    }
    if err == nil {
      s, _ := os.Stat(path)
      if err==nil {
        if s.IsDir()==false {
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
  return size, nil
}

func (d *Package) WriteControlFile (sourceDir string, size uint64) error {

  desc := d.Description
  if d.DescriptionExtended!="" {
    desc += "\n"
    for _, line := range strings.Split(d.DescriptionExtended, "\n") {
      desc += " " + line +"\n"
    }
    desc = strings.TrimSpace(desc)
  }

  P := ""
  P += strAppend("Package", d.Name)
  P += strAppend("Version", d.Version)
  P += strAppend("Source", d.Name)
  P += strAppend("Section", d.Section)
  P += strAppend("Priority", d.Priority)
  P += strAppend("Maintainer", d.Maintainer)
  P += strSliceAppend("Build-Depends", d.BuildDepends, ",")
  P += strSliceAppend("Build-Depends-Indep", d.BuildDepends, ",")
  P += strSliceAppend("Build-Conflicts", d.BuildConflicts, ",")
  P += strSliceAppend("Build-Conflicts-Indep", d.BuildConflictsIndep, ",")
  P += strAppend("Standards-Version", d.StandardsVersion)
  P += strAppend("Homepage", d.Homepage)
  P += strAppend("Description", desc)
  P += strAppend("Architecture", d.Arch)
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
  P += strAppend("Installed-Size", humanize.Bytes(size))
  P += strAppend("Package-Type", d.PackageType)

  controlContent := []byte(P)
  control := filepath.Join(sourceDir, "DEBIAN", "control")

  return ioutil.WriteFile(control, controlContent, 0644)
}

func (d *Package) WriteCopyrightFile (sourceDir string) error {
  content := ""
  content += strAppend("Format-Specification", d.CopyrightSpecUrl)
  content += strAppend("Name", d.Name)
  content += strAppend("Maintainer", d.Maintainer)
  sourcesUrl := d.SourcesUrl
  if sourcesUrl == "" {
    sourcesUrl = d.Homepage
  }
  content += strAppend("Source", sourcesUrl)
  content += "\n"

  // write the copyrights
  for _, c := range d.Copyrights {
    t := ""
    if c.Files!= "" {
      t += strAppend("Files", c.Files)
    }
    if c.Copyright!= "" {
      t += strAppend("Copyright", c.Copyright)
    }
    if c.License!= "" {
      t += strAppend("License", c.License)
    }
    if t!= "" {
      content += t + "\n"
    }
  }
  file := filepath.Join(sourceDir, "DEBIAN", "copyright")
  return ioutil.WriteFile(file, []byte(content), 0644)
}

func (d *Package) WriteCompatFile (sourceDir string) error {
  content := d.Compat+"\n"
  file := filepath.Join(sourceDir, "DEBIAN", "compat")
  return ioutil.WriteFile(file, []byte(content), 0644)
}

func (d *Package) WriteCronFiles (sourceDir string) error {

  for k, val := range d.CronFiles {
    if val!="" {
      file := filepath.Join(sourceDir, "DEBIAN", d.Name + "-f.cron."+k)
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
      file := filepath.Join(sourceDir, "DEBIAN", d.Name + "-c.cron."+k)
      if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
        return err
      } else {
        fmt.Printf("cron job '%s' is empty!\n", k)
      }
    }
  }

  return nil
}

func (d *Package) WriteChangelogFile (sourceDir string) error {

  file := filepath.Join(sourceDir, "DEBIAN", "changelog")
  if d.ChangelogFile!="" {
    if err := cp(file, d.ChangelogFile); err != nil {
      return err
    }
  } else if d.ChangelogCmd!="" {
    cmd, err := stringexec.Command(sourceDir, d.ChangelogCmd)
    if err != nil {
      return err
    }
    content, err := cmd.Output()
    if err != nil {
      return err
    }
    if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
      return err
    }
  }

  return nil
}

func (d *Package) WriteManPageIndexFile (sourceDir string) error {

  file := filepath.Join(sourceDir, "DEBIAN", d.Name + ".manpages")
  if len(d.Mans)>0 {
    content := strings.Join(d.Mans, "\n")
    if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
      return err
    }
  }

  return nil
}

func (d *Package) WriteShortcuts (sourceDir string) error {

  for _, m := range d.Menus {
    s := ""

    if m.Name!="" {
      s = "Name=" + m.Name
    }

    if m.Description!="" {
      s = "Description=" + m.Description
    }

    if m.GenericName!="" {
      s = "GenericName=" + m.GenericName
    }

    if m.Exec!="" {
      s = "Exec=" + m.Exec
    }

    if m.Icon!="" {
      s = "Icon=/usr/share/pixmaps/" + filepath.Base(m.Icon)
    }

    if m.Categories!="" {
      s = "Categories=" + m.Categories
    }

    if m.MimeType!="" {
      s = "MimeType=" + m.MimeType
    }

    if s!="" {

      if m.StartupNotify {
        s = "StartupNotify=true"
      } else {
        s = "StartupNotify=false"
      }

      s = "[Desktop Entry]\n" + s

      file := filepath.Join(sourceDir, "usr", "share", "application", m.Name + ".desktop")
      if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
        return err
      }
      if err := ioutil.WriteFile(file, []byte(s), 0644); err != nil {
        return err
      }

      icoFile := filepath.Join(sourceDir, "usr", "share", "pixmaps", filepath.Base(m.Icon) )
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

func (d *Package) WriteInitFile (sourceDir string) error {
  if d.InitFile!="" {
    return writeAFile(filepath.Join(sourceDir, d.Name+".init"), d.InitFile)
  }
  return nil
}

func (d *Package) WriteDefaultInitFile (sourceDir string) error {
  if d.DefaultFile!="" {
    return writeAFile(filepath.Join(sourceDir, d.Name+".default"), d.DefaultFile)
  }
  return nil
}

func (d *Package) WritePreInstFile (sourceDir string) error {
  if d.PreinstFile!="" {
    return writeAFile(filepath.Join(sourceDir, "preinst"), d.PreinstFile)
  }
  return nil
}

func (d *Package) WritePostInstFile (sourceDir string) error {
  if d.PostinstFile!="" {
    return writeAFile(filepath.Join(sourceDir, "postinst"), d.PostinstFile)
  }
  return nil
}

func (d *Package) WritePreRmFile (sourceDir string) error {
  if d.PrermFile!="" {
    return writeAFile(filepath.Join(sourceDir, "prerm"), d.PrermFile)
  }
  return nil
}

func (d *Package) WritePostRmFile (sourceDir string) error {
  if d.PostrmFile!="" {
    return writeAFile(filepath.Join(sourceDir, "postrm"), d.PostrmFile)
  }
  return nil
}

func (d *Package) WriteRulesFile (sourceDir string) error {
  if d.RulesFile!="" {
    return writeAFile(filepath.Join(sourceDir, "rules"), d.RulesFile)
  }
  return nil
}


func writeAFile (dst string, src string) error {
  if len(src)>0 {
    err := cp(dst, src)
    if err !=nil {
      m := fmt.Sprintf("Could not write file '%s': %s", src, err.Error())
      return errors.New(m)
    }
  }
  return nil
}


func strAppend(name string, value string) string {
  ret := ""
  if len(value)>0 {
    ret = fmt.Sprintf("%s: %s\n", name, value)
  }
  return ret
}
func boolAppend(name string, value bool) string {
  ret := ""
  v := "no"
  if value {
    v = "yes"
  }
  ret = fmt.Sprintf("%s: %s\n", name, v)
  return ret
}
func vcsSliceAppend(s []VcsSrc) string {
  ret := ""
  for _, k := range s {
    ret += structAppend(k)
  }
  return ret
}
func structAppend(s fmt.Stringer) string {
  ret := ""
  value := s.String()
  if len(value)>0 {
    ret = fmt.Sprintf("%s\n", value)
  }
  return ret
}
func strSliceAppend(name string, value []string, j string) string {
  ret := ""
  if len(value)>0 {
    ret = fmt.Sprintf("%s: %s\n", name, strings.Join(value, j))
  }
  return ret
}
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
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
