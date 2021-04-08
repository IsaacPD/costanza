PROJECT=costanza
OUT_DIR=out
REPOPATH ?= github.com/isaacpd/$(PROJECT)
BUILD_PACKAGE = $(REPOPATH)/src/$(PROJECT)

GOPATH ?= $(shell go env GOPATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GO_FILES = $(shell find . -type f -name '*.go')

ifeq ($(GOOS),windows)
CP = copy /y
INSTALL_DIR = C:\bin
PROJECT = $(PROJECT).exe
else
CP = cp
INSTALL_DIR = /usr/bin
endif

$(OUT_DIR)/$(PROJECT): $(GO_FILES)
	go build -o $(OUT_DIR)/$(PROJECT) $(BUILD_PACKAGE)

install: $(OUT_DIR)/$(PROJECT)
	$(CP) "$(OUT_DIR)\$(PROJECT)" "$(INSTALL_DIR)"