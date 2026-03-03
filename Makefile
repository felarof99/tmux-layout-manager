BINARY := layouts
GOBIN := $(shell go env GOPATH)/bin
VERSION ?= 0.1.0
FISH_FUNCTIONS ?= $(HOME)/.config/fish/functions

.PHONY: build install completions fish clean

build:
	go build -ldflags "-X layouts/cmd.Version=$(VERSION)" -o $(BINARY) .

install: build
	cp $(BINARY) $(GOBIN)/$(BINARY)
	codesign --force --sign - $(GOBIN)/$(BINARY)
	@echo "Installed $(BINARY) to $(GOBIN)/$(BINARY)"

completions: install
	$(GOBIN)/$(BINARY) completion fish > ~/.config/fish/completions/$(BINARY).fish

fish:
	mkdir -p $(FISH_FUNCTIONS)
	cp ly.fish $(FISH_FUNCTIONS)/ly.fish
	@echo "Installed ly alias to $(FISH_FUNCTIONS)/ly.fish"

clean:
	rm -f $(BINARY)
