package main

import (
	"os"

	importusers "github.com/osu-acm/acm-votes/import_users"
	initalizeelection "github.com/osu-acm/acm-votes/initalize_election"
	"github.com/osu-acm/acm-votes/server"
	tallyballot "github.com/osu-acm/acm-votes/tally_ballot"
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
	case "tally":
		tallyballot.Main()
	case "initalize_election":
		initalizeelection.Main()
	case "import_users":
		importusers.Main()
	case "server":
		server.Main()
	}
}
