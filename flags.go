package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var (
	AppNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "Build output app name",
	}

	WorkDirFlag = cli.StringFlag{
		Name:   "workdir, d",
		Usage:  "Change workdir to this path",
		EnvVar: "PWD",
	}

	BuilderImageFlag = cli.StringFlag{
		Name:   "builder-image, bi",
		Value:  "golang:1.8-alpine",
		EnvVar: "GTD_BUILDER_IMAGE",
		Usage:  "Builder image",
	}

	BuilderImageUserFlag = cli.StringFlag{
		Name:   "builder-image-user, biu",
		EnvVar: "GTD_BUILDER_IMAGE_USER",
		Usage:  "Builder image user (format: <name|uid>[:<group|gid>])",
	}

	ResFlag = cli.StringSliceFlag{
		Name:  "res",
		Usage: "App related resources, app will depends on these files, e.g: *.conf",
	}

	RegistryFlag = cli.StringFlag{
		Name:   "registry, r",
		Usage:  "The registry host to build and push",
		Value:  "",
		EnvVar: "GTD_REGISTRY",
	}

	OrgFlag = cli.StringFlag{
		Name:   "organization, o",
		Usage:  "Which registry organization you will push",
		Value:  "",
		EnvVar: "GTD_ORG",
	}

	TagFlag = cli.StringSliceFlag{
		Name:  "tag, t",
		Usage: "Build image with these tags",
	}

	ExposeFlag = cli.StringFlag{
		Name:  "args, a",
		Usage: "Args for render Dockerfile template, it should be JSON format",
	}

	AppImageUserFlag = cli.StringFlag{
		Name:   "app-image-user, aiu",
		EnvVar: "GTD_APP_IMAGE_USER",
		Usage:  "App image user (format: <name|uid>[:<group|gid>])",
	}

	AppImageFlag = cli.StringFlag{
		Name:   "app-image, ai",
		Value:  "alpine:latest",
		EnvVar: "GTD_APP_IMAGE",
		Usage:  "App run with this image",
	}

	TemplateFlag = cli.StringFlag{
		Name:  "template",
		Value: fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), "github.com/gogap/go-to-docker/builder/dockerfiles_tmpl/default"),
		Usage: "Build app docker image by this Dockerfile template",
	}

	URIFlag = cli.StringSliceFlag{
		Name:  "uri, u",
		Usage: "Trigger URI, supported: [HTTP-GET]",
	}

	BranchTagsConfigFlag = cli.StringFlag{
		Name:  "branch-tags-config",
		Usage: "revision branch name to docker's Tags config filepath",
	}

	DockerInDockerUserFlag = cli.StringFlag{
		Name:   "dind-user, du",
		EnvVar: "GTD_DIND_USER",
		Usage:  "Docker in docker user (format: <name|uid>[:<group|gid>])",
	}

	VerboseFlag = cli.BoolFlag{
		Name:  "verbose",
		Usage: "Print debug info",
	}

	FakeRevisionBranch = cli.StringFlag{
		Name:  "fake-branch, fb",
		Usage: "Sometimes we need build other branch's code and push to specific docker revision branch",
	}

	GoPathFlag = cli.StringFlag{
		Name:   "gopath",
		EnvVar: "GOPATH",
	}
)

var (
	BuildAppFlags = []cli.Flag{
		AppNameFlag,
		WorkDirFlag,
		BuilderImageFlag,
		BuilderImageUserFlag,
		ResFlag,
		VerboseFlag,
		GoPathFlag,
	}

	BuildImageFlags = []cli.Flag{
		AppNameFlag,
		WorkDirFlag,
		RegistryFlag,
		OrgFlag,
		TagFlag,
		ExposeFlag,
		AppImageFlag,
		AppImageUserFlag,
		TemplateFlag,
		BranchTagsConfigFlag,
		DockerInDockerUserFlag,
		FakeRevisionBranch,
		VerboseFlag,
		GoPathFlag,
	}

	BuildAllFlags = joinFlags(BuildAppFlags, BuildImageFlags)

	PushImageFlags = []cli.Flag{
		AppNameFlag,
		WorkDirFlag,
		RegistryFlag,
		OrgFlag,
		TagFlag,
		BranchTagsConfigFlag,
		FakeRevisionBranch,
		VerboseFlag,
	}

	PushTriggerFlags = []cli.Flag{
		URIFlag,
		VerboseFlag,
	}

	PushAllFlags = joinFlags(PushImageFlags, PushTriggerFlags)

	AllFlags = joinFlags(BuildAllFlags, PushAllFlags)

	ClearAppFlags = []cli.Flag{
		WorkDirFlag,
		VerboseFlag,
	}

	ClearImageFlags = []cli.Flag{
		AppNameFlag,
		WorkDirFlag,
		RegistryFlag,
		OrgFlag,
		TagFlag,
		VerboseFlag,
	}

	ClearAllFlags = joinFlags(ClearAppFlags, ClearImageFlags)
)

func joinFlags(a, b []cli.Flag) []cli.Flag {
	c := []cli.Flag{}

	cache := map[string]bool{}

	for i := 0; i < len(a); i++ {
		if _, exist := cache[a[i].GetName()]; exist {
			continue
		}
		cache[a[i].GetName()] = true
		c = append(c, a[i])
	}

	for i := 0; i < len(b); i++ {
		if _, exist := cache[b[i].GetName()]; exist {
			continue
		}
		cache[b[i].GetName()] = true
		c = append(c, b[i])
	}

	return c
}
