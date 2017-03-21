package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gogap/go-to-docker/builder"
	"github.com/gogap/logrus_mate"
	"github.com/urfave/cli"
)

var (
	logger = logrus_mate.Logger()
)

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Name = "go-to-docker"
	app.Usage = "a tool for build your app and push images to docker registry"
	app.HelpName = "go-to-docker"

	app.Commands = []cli.Command{
		{
			Name:  "build",
			Usage: "Build app and image",
			Subcommands: []cli.Command{
				{
					Name:   "app",
					Usage:  "Build golang application by docker image",
					Action: cmdBuildApp,
					Flags:  BuildAppFlags,
				},
				{
					Name:   "image",
					Usage:  "Build app image with dockerfile template",
					Action: cmdBuildImage,
					Flags:  BuildImageFlags,
				},
				{
					Name:   "all",
					Usage:  "Build app and image",
					Action: cmdBuildAll,
					Flags:  BuildAllFlags,
				},
			},
		},
		{
			Name:  "push",
			Usage: "Build image and trigger",
			Subcommands: []cli.Command{
				{
					Name:   "image",
					Usage:  "push image to docker registry",
					Action: cmdPushImage,
					Flags:  PushImageFlags,
				},
				{
					Name:   "trigger",
					Usage:  "push the trigger after image pushed, it could be us in scene of continuous integration",
					Action: cmdPushTrigger,
					Flags:  PushTriggerFlags,
				},
				{
					Name:   "all",
					Usage:  "push image and trigger",
					Action: cmdPushAll,
					Flags:  PushAllFlags,
				},
			},
		},
		{
			Name:   "all",
			Usage:  "Build app and image, then push image and trigger",
			Action: cmdAll,
			Flags:  AllFlags,
		},
		{
			Name:  "clear",
			Usage: "Clear app's build output and image",
			Subcommands: []cli.Command{
				{
					Name:   "app",
					Action: cmdClearApp,
					Flags:  ClearAppFlags,
				},
				{
					Name:   "image",
					Action: cmdClearImage,
					Flags:  ClearImageFlags,
				},
				{
					Name:   "all",
					Action: cmdClearAll,
					Flags:  ClearAllFlags,
				},
			},
		},
	}

	var err error
	err = app.Run(os.Args)

	if err != nil {
		logger.Errorln(err)
	}
}

func cmdBuildApp(c *cli.Context) (err error) {

	appName := c.String("name")
	workdir := c.String("workdir")
	image := c.String("builder-image")
	user := c.String("builder-image-user")

	resources := c.StringSlice("res")
	verbose := c.Bool("verbose")
	gopath := c.String("gopath")

	if len(gopath) == 0 {
		gopath = os.Getenv("GOPATH")
	}

	if err = os.Setenv("GOPATH", gopath); err != nil {
		return
	}

	if appName == "" {
		appName = getDefaultAppName(workdir)
	}

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose:          verbose,
			BuilderImage:     image,
			AppImage:         "",
			WorkDir:          workdir,
			AppName:          appName,
			DockerfileTmpl:   "",
			AppImageTags:     nil,
			Exposes:          nil,
			AppArgs:          nil,
			Resources:        resources,
			BuilderImageUser: user,
			GoPath:           gopath,
		},
	}

	if err = bder.BuildApp(); err != nil {
		return
	}

	return
}

func cmdBuildImage(c *cli.Context) (err error) {

	appName := c.String("name")
	workdir := c.String("workdir")
	image := c.String("app-image")
	user := c.String("app-image-user")
	registry := c.String("registry")
	organization := c.String("organization")
	tag := c.StringSlice("tag")
	template := c.String("template")
	expose := c.StringSlice("expose")
	branchTagConfigFilename := c.String("branch-tags-config")
	fakeBranchName := c.String("fake-branch")
	verbose := c.Bool("verbose")

	if appName == "" {
		appName = getDefaultAppName(workdir)
	}

	var branchTagsConfig builder.BranchTagsConfig
	if len(branchTagConfigFilename) > 0 {
		if branchTagsConfig, err = loadBranchTagConfig(branchTagConfigFilename); err != nil {
			return
		}
	}

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose:          verbose,
			BuilderImage:     "",
			AppImage:         image,
			AppImageUser:     user,
			WorkDir:          workdir,
			AppName:          appName,
			RegistryHost:     registry,
			RegistryOrg:      organization,
			AppImageTags:     tag,
			DockerfileTmpl:   template,
			Exposes:          expose,
			AppArgs:          nil,
			Resources:        nil,
			BranchTagsConfig: branchTagsConfig,
			RevisionBranch:   fakeBranchName,
		},
	}

	if err = bder.BuildImage(); err != nil {
		return
	}

	return
}

func cmdClearApp(c *cli.Context) (err error) {

	workdir := c.String("workdir")
	verbose := c.Bool("verbose")

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose: verbose,
			WorkDir: workdir,
		},
	}

	if err = bder.ClearApp(); err != nil {
		return
	}

	return
}

func cmdClearImage(c *cli.Context) (err error) {

	verbose := c.Bool("verbose")
	appName := c.String("name")
	tag := c.StringSlice("tag")
	registry := c.String("registry")
	organization := c.String("organization")
	workdir := c.String("workdir")
	fakeBranchName := c.String("fake-branch")

	if appName == "" {
		appName = getDefaultAppName(workdir)
	}

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose:        verbose,
			AppImageTags:   tag,
			AppName:        appName,
			RegistryHost:   registry,
			RegistryOrg:    organization,
			RevisionBranch: fakeBranchName,
		},
	}

	if err = bder.ClearImage(); err != nil {
		return
	}

	return
}

func cmdClearAll(c *cli.Context) (err error) {

	if err = cmdClearApp(c); err != nil {
		return
	}

	if err = cmdClearImage(c); err != nil {
		return
	}

	return
}

func cmdPushImage(c *cli.Context) (err error) {
	verbose := c.Bool("verbose")
	appName := c.String("name")
	tag := c.StringSlice("tag")
	registry := c.String("registry")
	organization := c.String("organization")
	workdir := c.String("workdir")
	dockerInDockerUser := c.String("dind-user")
	branchTagConfigFilename := c.String("branch-tags-config")
	fakeBranchName := c.String("fake-branch")

	if appName == "" {
		appName = getDefaultAppName(workdir)
	}

	var branchTagsConfig builder.BranchTagsConfig
	if len(branchTagConfigFilename) > 0 {
		if branchTagsConfig, err = loadBranchTagConfig(branchTagConfigFilename); err != nil {
			return
		}
	}

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose:            verbose,
			AppImageTags:       tag,
			AppName:            appName,
			RegistryHost:       registry,
			RegistryOrg:        organization,
			BranchTagsConfig:   branchTagsConfig,
			DockerInDockerUser: dockerInDockerUser,
			RevisionBranch:     fakeBranchName,
		},
	}

	if err = bder.PushImage(); err != nil {
		return
	}

	return
}

func cmdPushTrigger(c *cli.Context) (err error) {
	uri := c.StringSlice("uri")
	verbose := c.Bool("verbose")

	bder := &builder.Builder{
		Options: builder.BuildOptions{
			Verbose:     verbose,
			TriggerURIs: uri,
		},
	}

	if err = bder.PushTrigger(); err != nil {
		return
	}

	return
}

func cmdPushAll(c *cli.Context) (err error) {
	if err = cmdPushImage(c); err != nil {
		return
	}

	if err = cmdPushTrigger(c); err != nil {
		return
	}

	return
}

func cmdBuildAll(c *cli.Context) (err error) {
	if err = cmdBuildApp(c); err != nil {
		return
	}

	if err = cmdBuildImage(c); err != nil {
		return
	}

	return
}

func cmdAll(c *cli.Context) (err error) {
	if err = cmdBuildApp(c); err != nil {
		return
	}

	if err = cmdBuildImage(c); err != nil {
		return
	}

	if err = cmdPushImage(c); err != nil {
		return
	}

	if err = cmdPushTrigger(c); err != nil {
		return
	}

	return
}

func getDefaultAppName(cwd string) (name string) {
	if cwd == "" {
		name = "app"
		return
	}

	name = filepath.Base(cwd)

	return
}

func loadBranchTagConfig(filename string) (config builder.BranchTagsConfig, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	buf := bytes.NewBuffer(data)

	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	if err = decoder.Decode(&config); err != nil {
		return
	}

	return
}
