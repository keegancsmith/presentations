GOBIN:=${PWD}/.bin

serve: ${GOBIN}/present
	present -notes

${GOBIN}/present:
	go install golang.org/x/tools/cmd/present

.PHONY: serve
