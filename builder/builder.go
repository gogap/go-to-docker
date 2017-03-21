package builder

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/gogap/logrus_mate"
)

var (
	logger = logrus_mate.Logger()
)

type BuildOption func(*BuildOptions)

type Builder struct {
	Options BuildOptions

	initOnce sync.Once
}

type BuildOptions struct {
	Verbose            bool
	BuilderImage       string
	AppImage           string
	WorkDir            string
	AppName            string
	DockerfileTmpl     string
	AppImageTags       []string
	Exposes            []string
	BuildOutputDir     string
	RegistryHost       string
	RegistryOrg        string
	RegistryUsername   string
	RegistryPassword   string
	AppArgs            map[string]string
	TriggerURIs        []string
	Resources          []string
	BuilderImageUser   string
	AppImageUser       string
	RevisionBranch     string
	RevisionID         string
	BranchTagsConfig   BranchTagsConfig
	DockerInDockerUser string
	GoPath             string
}

func Verbose(v bool) BuildOption {
	return func(o *BuildOptions) {
		o.Verbose = v
		if v {
			logger.Level = logrus.DebugLevel
		} else {
			logger.Level = logrus.WarnLevel
		}
	}
}

func BuilderImage(img string, tags ...string) BuildOption {
	return func(o *BuildOptions) {
		o.BuilderImage = img
	}
}

func DockerfileTemplate(tmpl string) BuildOption {
	return func(o *BuildOptions) {
		o.DockerfileTmpl = tmpl
	}
}

func AppImage(img string) BuildOption {
	return func(o *BuildOptions) {
		o.AppImage = img
	}
}

func WorkDir(dir string) BuildOption {
	return func(o *BuildOptions) {
		o.WorkDir = dir
	}
}

func AppName(name string) BuildOption {
	return func(o *BuildOptions) {
		o.AppName = name
	}
}

func AppImageTags(tags ...string) BuildOption {
	return func(o *BuildOptions) {
		o.AppImageTags = tags
	}
}

func BuildOutputDir(dir string) BuildOption {
	return func(o *BuildOptions) {
		o.BuildOutputDir = dir
	}
}

func (p *Builder) initOptions() (err error) {

	p.initOnce.Do(func() {

		if p.Options.AppName == "" {
			p.Options.AppName = "app"
		}

		if p.Options.WorkDir == "" {
			if p.Options.WorkDir, err = os.Getwd(); err != nil {
				return
			}
		}

		if p.Options.BuilderImage == "" {
			p.Options.BuilderImage = "golang:1.8-alpine"
		}

		if p.Options.AppImage == "" {
			p.Options.AppImage = "alpine:latest"
		}

		if p.Options.BuildOutputDir == "" {
			p.Options.BuildOutputDir = "_output_"
		}

		if p.Options.Verbose {
			logger.Level = logrus.DebugLevel
		} else {
			logger.Level = logrus.WarnLevel
		}

		if p.Options.DockerfileTmpl == "" {
			p.Options.DockerfileTmpl = filepath.Join(os.Getenv("GOPATH"), "src", "github.com/gogap/go-to-docker/builder/dockerfiles_templ/default")
		}

		var isGit bool
		var revisionBranch, revisionID string

		if revisionBranch, revisionID, isGit, err = getRevision(p.Options.WorkDir); err != nil {
			return
		} else if isGit {

			if len(p.Options.RevisionBranch) == 0 {
				p.Options.RevisionBranch = revisionBranch
			}

			p.Options.RevisionID = revisionID

			branchHasTags := false
			if p.Options.BranchTagsConfig.Branchs != nil {
				if branchTag, exist := p.Options.BranchTagsConfig.Branchs[p.Options.RevisionBranch]; exist {

					p.Options.RegistryUsername = branchTag.Username
					p.Options.RegistryPassword = branchTag.Password
					p.Options.RegistryHost = branchTag.Server
					p.Options.RegistryOrg = branchTag.Organization
					if len(branchTag.Tags) > 0 {
						branchHasTags = true
						p.Options.AppImageTags = append(p.Options.AppImageTags, branchTag.Tags...)
					}
				}
			}

			if !branchHasTags {
				p.Options.AppImageTags = append(p.Options.AppImageTags, p.Options.RevisionBranch, p.Options.RevisionBranch+"-"+p.Options.RevisionID)
			}
		} else if len(p.Options.AppImageTags) == 0 {
			p.Options.AppImageTags = []string{"latest"}
		}
	})

	return
}

// docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.6 go build -v
func (p *Builder) BuildApp() (err error) {

	if err = p.initOptions(); err != nil {
		return
	}

	cwd, _ := os.Getwd()
	if cwd != p.Options.WorkDir {
		os.Chdir(p.Options.WorkDir)
	}

	defer func() {
		if cwd != p.Options.WorkDir {
			os.Chdir(cwd)
		}
	}()

	buildpath := filepath.Join(p.Options.BuildOutputDir, p.Options.AppName)

	buildCMD := ""
	if p.Options.BuilderImage == "local" {
		logger.Debugln("use local go build")
		buildCMD = fmt.Sprintf("go build -o %s", buildpath)
	} else {
		if len(p.Options.BuilderImageUser) == 0 {
			buildCMD = fmt.Sprintf(`docker run --rm -v %s:/usr/src/myapp -v %s:/go -w /usr/src/myapp %s go build -o %s`, p.Options.WorkDir, p.Options.GoPath, p.Options.BuilderImage, buildpath)
		} else {
			buildCMD = fmt.Sprintf(`docker run --rm -u %s -v %s:/usr/src/myapp -v %s:/go -w /usr/src/myapp %s go build -o %s`, p.Options.BuilderImageUser, p.Options.WorkDir, p.Options.GoPath, p.Options.BuilderImage, buildpath)
		}
	}

	if p.Options.Verbose {
		buildCMD += " -v"
	}

	logger.Debugln(buildCMD)

	if err = execCommandToShow(p.Options.WorkDir, buildCMD); err != nil {
		return
	}

	var resPaths []string

	if len(p.Options.Resources) > 0 {
		for i := 0; i < len(p.Options.Resources); i++ {
			var paths []string
			if paths, err = filepath.Glob(p.Options.Resources[i]); err != nil {
				return
			}

			var washedPaths []string
			for j := 0; j < len(paths); j++ {
				washedpath := paths[j]
				if filepath.IsAbs(paths[j]) {
					if washedpath, err = filepath.Rel(p.Options.WorkDir, paths[j]); err != nil {
						return
					}
				}
				washedPaths = append(washedPaths, washedpath)
			}
			resPaths = append(resPaths, washedPaths...)
		}
	}

	for i := 0; i < len(resPaths); i++ {

		confPath := filepath.Join(p.Options.BuildOutputDir, resPaths[i])
		confDir, _ := filepath.Split(confPath)
		if err = os.MkdirAll(confDir, 0755); err != nil {
			return
		}

		logger.Debugf("copying file %s to %s", resPaths[i], confPath)
		if err = copyfile(resPaths[i], confPath); err != nil {
			return
		}
	}

	return
}

// BuildImage is for build app image
// docker build -t
// step 1: build template
// step 2: build app image
func (p *Builder) BuildImage() (err error) {
	if err = p.initOptions(); err != nil {
		return
	}

	if len(p.Options.RegistryOrg) == 0 {
		err = errors.New("docker registry organization could not be empty")
		return
	}

	cwd, _ := os.Getwd()
	if cwd != p.Options.WorkDir {
		os.Chdir(p.Options.WorkDir)
	}

	defer func() {
		if cwd != p.Options.WorkDir {
			os.Chdir(cwd)
		}
	}()

	var fi os.FileInfo
	if fi, err = os.Stat(p.Options.BuildOutputDir); err != nil {
		return
	}

	if !fi.IsDir() {
		err = errors.New("output path should be a dir, not a file")
		return
	}

	binpath := filepath.Join(p.Options.BuildOutputDir, p.Options.AppName)
	if fi, err = os.Stat(binpath); err != nil {
		if os.IsNotExist(err) {
			err = errors.New("please build app first")
			return
		}
		return
	}

	if fi.IsDir() {
		err = errors.New(binpath + " should be an executable file")
		return
	}

	logger.Debugf("using Dockerfile template of %s", p.Options.DockerfileTmpl)

	var tmplbuf []byte
	if tmplbuf, err = ioutil.ReadFile(p.Options.DockerfileTmpl); err != nil {
		return
	}

	var tmpl *template.Template
	if tmpl, err = template.New(p.Options.AppImage).Parse(string(tmplbuf)); err != nil {
		return
	}

	buf := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buf, p.Options); err != nil {
		return
	}

	dockerfileContent := buf.Bytes()

	// docker build -t xxxx .
	baseTagName := filepath.Join(p.Options.RegistryHost, p.Options.RegistryOrg, p.Options.AppName)

	var tags = ""
	for i := 0; i < len(p.Options.AppImageTags); i++ {
		tags = tags + fmt.Sprintf(" -t %s:%s", baseTagName, p.Options.AppImageTags[i])
	}

	tags = strings.TrimSpace(tags)

	dockerfilePath := filepath.Join(p.Options.BuildOutputDir, "Dockerfile")
	if err = ioutil.WriteFile(dockerfilePath, dockerfileContent, 0644); err != nil {
		return
	}

	os.RemoveAll(filepath.Join(p.Options.BuildOutputDir, ".docker"))

	buildCMD := fmt.Sprintf("docker build %s .", tags)

	logger.Debugln(buildCMD)

	if err = execCommandToShow(p.Options.BuildOutputDir, buildCMD); err != nil {
		return
	}

	return
}

// docker push [OPTIONS] NAME[:TAG]
func (p *Builder) PushImage() (err error) {
	if err = p.initOptions(); err != nil {
		return
	}

	if len(p.Options.RegistryOrg) == 0 {
		err = errors.New("docker registry organization could not be empty")
		return
	}

	var tmpDockerconf string
	if tmpDockerconf, err = filepath.Abs(filepath.Join(p.Options.BuildOutputDir, ".docker")); err != nil {
		return
	}

	dockerInDockerFMT := "docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock -v " + tmpDockerconf + ":/root/.docker docker:dind %s"

	if len(p.Options.RegistryUsername) > 0 {
		cmdLogin := fmt.Sprintf(dockerInDockerFMT, fmt.Sprintf("docker login -u %s -p %s %s", p.Options.RegistryUsername, p.Options.RegistryPassword, p.Options.RegistryHost))

		if err = execCommandToShow("", cmdLogin); err != nil {
			return
		}

		if len(p.Options.DockerInDockerUser) > 0 {
			// change own
			dockerInDockerFMT = "docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock -v " + tmpDockerconf + ":/root/.docker docker:dind %s"
			cmdChown := fmt.Sprintf(dockerInDockerFMT, fmt.Sprintf("chown -R %s /root/.docker", p.Options.DockerInDockerUser))

			if err = execCommandToShow("", cmdChown); err != nil {
				return
			}

			defer func() {
				if e := os.RemoveAll(tmpDockerconf); e != nil {
					logger.Warnln(e)
				}
			}()
		}

	}

	baseTagName := filepath.Join(p.Options.RegistryHost, p.Options.RegistryOrg, p.Options.AppName)

	for i := 0; i < len(p.Options.AppImageTags); i++ {
		pushCMD := fmt.Sprintf(dockerInDockerFMT, fmt.Sprintf("docker push %s:%s", baseTagName, p.Options.AppImageTags[i]))
		logger.Debugln(pushCMD)

		if err = execCommandToShow("", pushCMD); err != nil {
			return
		}

	}

	return
}

func (p *Builder) PushTrigger() (err error) {
	if err = p.initOptions(); err != nil {
		return
	}

	if len(p.Options.TriggerURIs) == 0 {
		return
	}

	var httpTriggers []string
	for i := 0; i < len(p.Options.TriggerURIs); i++ {
		if strings.Index(p.Options.TriggerURIs[i], "http") == 0 ||
			strings.Index(p.Options.TriggerURIs[i], "https") == 0 {
			httpTriggers = append(httpTriggers, p.Options.TriggerURIs[i])
		}
	}

	if err = pushHttpTriggers(httpTriggers); err != nil {
		return
	}

	return
}

func pushHttpTriggers(uris []string) (err error) {
	if len(uris) == 0 {
		return
	}

	for i := 0; i < len(uris); i++ {

		var resp *http.Response
		if resp, err = http.DefaultClient.Get(uris[i]); err != nil {
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("uri: %s\nstatus code: %d, body: \n%s\n", uris[i], resp.StatusCode, string(body))
			return
		}

		logger.Infof("trigger: %s", uris[i])
		logger.Debugf("body:\n%s", string(body))
	}

	return
}

func (p *Builder) ClearApp() (err error) {

	if err = p.initOptions(); err != nil {
		return
	}

	if p.Options.BuildOutputDir == "/" ||
		p.Options.BuildOutputDir == "" {
		return
	}

	cwd, _ := os.Getwd()
	if cwd != p.Options.WorkDir {
		os.Chdir(p.Options.WorkDir)
	}

	defer func() {
		if cwd != p.Options.WorkDir {
			os.Chdir(cwd)
		}
	}()

	if _, err = os.Stat(p.Options.BuildOutputDir); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}

	if err = os.RemoveAll(p.Options.BuildOutputDir); err != nil {
		return
	}

	return
}

func (p *Builder) ClearImage() (err error) {

	if err = p.initOptions(); err != nil {
		return
	}

	if len(p.Options.RegistryOrg) == 0 {
		err = errors.New("docker registry organization could not be empty")
		return
	}

	baseTagName := filepath.Join(p.Options.RegistryHost, p.Options.RegistryOrg, p.Options.AppName)

	var images []string
	for i := 0; i < len(p.Options.AppImageTags); i++ {
		image := fmt.Sprintf("%s:%s", baseTagName, p.Options.AppImageTags[i])
		images = append(images, image)
	}

	rmiCMD := "docker rmi " + strings.Join(images, " ")
	logger.Debugln(rmiCMD)

	var out []byte
	if out, err = execCommand("", rmiCMD); err != nil {
		err = errors.New(string(out))
		return
	}

	return
}
