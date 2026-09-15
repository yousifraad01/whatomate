// Sentinel module: it exists only so that "go build ./...", "go vet ./..." and
// "go test ./..." run from the repository root do not descend into
// frontend/node_modules, where some npm packages ship Go sources (for example
// flatted/golang). The frontend itself is a Node project; nothing here is
// compiled by Go.
module github.com/shridarpatil/whatomate/frontend

go 1.25
