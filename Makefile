GOBIN:=${PWD}/.bin

all: README.md serve

serve: ${GOBIN}/present
	present -notes

${GOBIN}/present:
	go install golang.org/x/tools/cmd/present

README.md: $(wildcard */*.slide)
	./gen-readme.sh

.PHONY: all serve
