DIST     := bin
BINARIES := $(DIST)/ecs-log-viewer
SOURCES  := $(wildcard *.go)

.PHONY: build clean run

build: $(BINARIES)
	go build -o $(DIST)/ecs-log-viewer

run: build
	$(DIST)/ecs-log-viewer

clean:
	rm -rf $(DIST)
