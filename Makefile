
serve: ${GOBIN}/present
	present -notes

${GOBIN}/present:
	go get golang.org/x/tools/cmd/present

.PHONY: serve
