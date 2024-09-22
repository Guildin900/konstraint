package konstraint

import (
	"os"

	"github.com/Guildin900/konstraint/porno/commands"
)

func Run() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
