
serve: ${GOBIN}/present
	present

${GOBIN}/present:
	go get golang.org/x/tools/cmd/present

.PHONY: serve
