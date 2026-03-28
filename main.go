package main

import (
	"os"

	importusers "github.com/osu-acm/acm-votes/import_users"
	"github.com/osu-acm/acm-votes/server"
)

func nOrDefault(list []string, n int, s string) string {
	if n >= len(list) {
		return s
	}
	return list[n]
}

func main() {
	subProgram := nOrDefault(os.Args, 1, "server")
	switch subProgram {
	case "import_users":
		importusers.Main()
	case "server":
		server.Main()
	}
}
