package dev

import (
	"errors"
	"fmt"
	"github.com/open-horizon/anax/i18n"
)

const MAKE_FILE = "Makefile"
const DOCKER_FILE = "Dockerfile"
const SERVICE_FILE = "service.sh"

var BASE_OS_VALUE = map[string]string{"amd64": "alpine:latest", "arm": "arm32v6/alpine:latest", "arm64": "arm64v8/alpine:latest"}

const DEFAULT_BASE_OS_VALUE = "alpine:latest"

const MAKE_FILE_CONTENT = `# Make targets for building the IBM example helloworld edge service

DOCKER_ENGINE := $(shell command -v docker 2> /dev/null)
ifndef DOCKER_ENGINE
    DOCKER_ENGINE := $(shell command -v podman 2> /dev/null)
    ifndef DOCKER_ENGINE
       $(error "Neither docker nor podman is available. Please install one of them")
    endif
endif

# This imports the variables from horizon/hzn.json. You can ignore these lines, but do not remove them.
-include PROJECT_DIR/.hzn.json.tmp.mk

# Default ARCH to the architecture of this machines (as horizon/golang describes it)
export ARCH ?= $(shell hzn architecture)

# Build the docker image for the current architecture
build:
	$(DOCKER_ENGINE) build -t $(DOCKER_IMAGE_BASE)_$(ARCH):$(SERVICE_VERSION) -f ./Dockerfile.$(ARCH) .

# Build the docker image for 3 architectures
build-all-arches:
	ARCH=amd64 $(MAKE) build
	ARCH=arm $(MAKE) build
	ARCH=arm64 $(MAKE) build

clean:
	-$(DOCKER_ENGINE) rmi $(DOCKER_IMAGE_BASE)_$(ARCH):$(SERVICE_VERSION) 2> /dev/null || :

clean-all-archs:
	ARCH=amd64 $(MAKE) clean
	ARCH=arm $(MAKE) clean
	ARCH=arm64 $(MAKE) clean

## This imports the variables from horizon/hzn.json. You can ignore these lines, but do not remove them.
PROJECT_DIR/.hzn.json.tmp.mk: PROJECT_DIR/hzn.json
	hzn util configconv -f $< > $@

.PHONY: build clean`

const DOCKER_FILE_CONTENT = `FROM BASE_OS

COPY *.sh /
WORKDIR /
CMD /service.sh`

const SERVICE_FILE_CONTENT = `#!/bin/sh

# Very simple Horizon sample edge service.

while true; do
  echo "$HZN_NODE_ID says: Hello ${HW_WHO}!"
  sleep 3
done
`

// It creates a sample Dockerfile for the service container image in the file system.
// For now we only handle the first image.
func CreateServiceImageFiles(directory string, hznenv_dir string) error {
	// get message printer
	msgPrinter := i18n.GetMessagePrinter()

	if err := CreateDockerFiles(directory); err != nil {
		return errors.New(msgPrinter.Sprintf("error creating %v for service image. %v", DOCKER_FILE, err))
	}
	if err := CreateMakeFile(directory, hznenv_dir); err != nil {
		return errors.New(msgPrinter.Sprintf("error creating %v for service image. %v", MAKE_FILE, err))
	}
	if err := CreateServiceFile(directory); err != nil {
		return errors.New(msgPrinter.Sprintf("error creating %v for service image. %v", SERVICE_FILE, err))
	} else {
		return nil
	}
}

// This function creates 3 Dockerfiles, for amd64, arm and arm64.
func CreateDockerFiles(directory string) error {
	// different arch requires different base os for Dockerfile,
	for arch, base_os := range BASE_OS_VALUE {
		fn := fmt.Sprintf("%v.%v", DOCKER_FILE, arch)
		if err := CreateFileWithConent(directory, fn, DOCKER_FILE_CONTENT, map[string]string{"BASE_OS": base_os}, false); err != nil {
			return err
		}
	}
	return nil
}

func CreateMakeFile(directory string, hznenv_dir string) error {
	return CreateFileWithConent(directory, MAKE_FILE, MAKE_FILE_CONTENT, map[string]string{"PROJECT_DIR": hznenv_dir}, false)
}

func CreateServiceFile(directory string) error {
	return CreateFileWithConent(directory, SERVICE_FILE, SERVICE_FILE_CONTENT, nil, true)
}
