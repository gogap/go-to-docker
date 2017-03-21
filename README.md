go-to-docker
============

Build your go program into docker image and push it to your own registry


> Your golang program code must be in a git repo

#### Install go-to-docker

```bash
go get github.com/gogap/go-to-docker
go install github.com/gogap/go-to-docker
```

```bash
> go-to-docker

NAME:
   go-to-docker - a tool for build your app and push images to docker registry

USAGE:
   go-to-docker [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     build    Build app and image
     push     Build image and trigger
     all      Build app and image, then push image and trigger
     clear    Clear app's build output and image
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```


#### Build App

##### Example

```bash
## dir: $GOPATH/src/gogap/example

go-to-docker build app
```

the app will build into `_output_` dir


#### help

```bash
go-to-docker build app --help
NAME:
   go-to-docker build app - Build golang application by docker image

USAGE:
   go-to-docker build app [command options] [arguments...]

OPTIONS:
   --name value                             Build output app name
   --workdir value, -d value                Change workdir to this path [$PWD]
   --builder-image value, --bi value        Builder image (default: "golang:1.8-alpine") [$GTD_BUILDER_IMAGE]
   --builder-image-user value, --biu value  Builder image user (format: <name|uid>[:<group|gid>]) [$GTD_BUILDER_IMAGE_USER]
   --res value                              App related resources, app will depends on these files, e.g: *.conf
   --verbose                                Print debug info
   --gopath value                            [$GOPATH]
```

#### Build Image

##### example

`branchs.conf`

```json
{
	"branchs":{
		"master":{
			"server":"registry.cn-beijing.aliyuncs.com",
			"username":"aliyun-account",
			"password":"aliyun-docker-registry-password",
			"organization":"aliyun-registry-namespace",
			"tags":[]
		}
	}
}
```

```bash
## dir: $GOPATH/src/gogap/example

go-to-docker build app --branch-tags-config ./branchs.conf
```

```bash
docker images

REPOSITORY                                      TAG                 IMAGE ID            CREATED             SIZE
registry.cn-beijing.aliyuncs.com/zeal/example   master              f0590d74dc99        12 minutes ago      5.54 MB
registry.cn-beijing.aliyuncs.com/zeal/example   master-00000000     f0590d74dc99        12 minutes ago      5.54 MB
```

##### help

```bash
> go-to-docker build image --help


NAME:
   go-to-docker build image - Build app image with dockerfile template

USAGE:
   go-to-docker build image [command options] [arguments...]

OPTIONS:
   --name value                         Build output app name
   --workdir value, -d value            Change workdir to this path [$PWD]
   --registry value, -r value           The registry host to build and push [$GTD_REGISTRY]
   --organization value, -o value       Which registry organization you will push [$GTD_ORG]
   --tag value, -t value                Build image with these tags
   --args value, -a value               Args for render Dockerfile template, it should be JSON format
   --app-image value, --ai value        App run with this image (default: "alpine:latest") [$GTD_APP_IMAGE]
   --app-image-user value, --aiu value  App image user (format: <name|uid>[:<group|gid>]) [$GTD_APP_IMAGE_USER]
   --template value                     Build app docker image by this Dockerfile template (default: "/gopath/src/github.com/gogap/go-to-docker/builder/dockerfiles_tmpl/default")
   --branch-tags-config value           revision branch name to docker's Tags config filepath
   --dind-user value, --du value        Docker in docker user (format: <name|uid>[:<group|gid>]) [$GTD_DIND_USER]
   --fake-branch value, --fb value      Sometimes we need build other branch's code and push to specific docker revision branch
   --verbose                            Print debug info
   --gopath value                        [$GOPATH]
```


#### Build all by one command

```bash
## dir: $GOPATH/src/gogap/example
go-to-docker build all --branch-tags-config ./branchs.conf
```


#### Push image to registry
```bash
## dir: $GOPATH/src/gogap/example
go-to-docker push image --branch-tags-config ./branchs.conf
```

#### Build, push by one command
```bash
## dir: $GOPATH/src/gogap/example
go-to-docker all --branch-tags-config ./branchs.conf
```