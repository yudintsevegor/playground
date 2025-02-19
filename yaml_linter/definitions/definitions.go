package definitions

import "embed"

var (
	//go:embed *.yaml
	EmbedFS embed.FS
)
