all: devel testsuite install clean

testsuite: test race benchmark

install: deps
	go install github.com/universonic/ivy-utils/cmd/...

deps:
	glide update

test: deps
	go test -cpu 1,4 -timeout 5m github.com/universonic/ivy-utils/...

race: deps
	go test -race -cpu 1,4 -timeout 7m github.com/universonic/ivy-utils/...

benchmark: deps
	go test -bench . -cpu 1,4 -timeout 10m github.com/universonic/ivy-utils/...

build: deps
	go build github.com/universonic/ivy-utils/cmd/...

clean:
	go clean -i github.com/universonic/ivy-utils/...

devel:
	go version
	go get -u github.com/Masterminds/glide

.PHONY: \
	testsuite \
	install \
	deps \
	test \
	race \
	benchmark \
	build \
	clean \