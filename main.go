package kontrainst

import (
	"os"

	"github.com/Guildin900/konstraint/internal/commands"
)

func Run() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
