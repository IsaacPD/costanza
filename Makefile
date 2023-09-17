PROJECT=costanza
OUT_DIR=out
PROTO_DIR=proto
REPOPATH ?= github.com/isaacpd/$(PROJECT)
BUILD_PACKAGE = $(REPOPATH)/cmd/$(PROJECT)

GOPATH ?= $(shell go env GOPATH)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GO_FILES = $(shell find . -type f -name '*.go')
PYTHON_FILES = $(shell find . -type f -name '*.{py,ipynb,pyi}')
PROTO_FILES = $(shell find . -type f -name '*.proto')

GO_PROTO_OUT = pkg
PYTHON_PROTO_OUT = python/proto

ifeq ($(GOOS),windows)
CP = copy /y
INSTALL_DIR = C:\bin
PROJECT = $(PROJECT).exe
else
CP = cp
INSTALL_DIR = /usr/bin
endif

$(OUT_DIR)/$(PROJECT): $(PROTO_FILES) $(GO_FILES)
	protoc --go-grpc_out=$(GO_PROTO_OUT) --go_out=$(GO_PROTO_OUT) $(PROTO_DIR)/chat.proto
	go build -o $(OUT_DIR)/$(PROJECT) $(BUILD_PACKAGE)

chat_server: $(PROTO_FILES) $(PYTHON_FILES)
	python3 -m grpc_tools.protoc --python_out=$(PYTHON_PROTO_OUT) --pyi_out=$(PYTHON_PROTO_OUT) --grpc_python_out=$(PYTHON_PROTO_OUT) --proto_path=$(PROTO_DIR) chat.proto

install: $(OUT_DIR)/$(PROJECT)
	$(CP) "$(OUT_DIR)/$(PROJECT)" "$(INSTALL_DIR)"
