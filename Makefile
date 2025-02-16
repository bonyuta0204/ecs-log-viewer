DIST     := bin
BINARIES := $(DIST)/ecs-log-viewer
SOURCES  := $(shell find . -name '*.go')

.PHONY: all build clean run

$(DIST):
	mkdir -p $(DIST)

$(BINARIES): $(DIST) $(SOURCES)
	go build -o $(BINARIES) ./cmd/ecs-log-viewer

build: $(BINARIES)

run: build
	$(BINARIES)

clean:
	rm -rf $(DIST)