PROJECT=costanza
OUT_DIR=out
REPOPATH ?= github.com/isaacpd/$(PROJECT)
BUILD_PACKAGE = $(REPOPATH)/src/$(PROJECT)


all:
	go build -o $(OUT_DIR)/$(PROJECT).exe $(BUILD_PACKAGE)

install: all
	copy /y "out\$(PROJECT).exe" "C:\bin"