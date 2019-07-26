package util

import (
	"fmt"
	"os"
)

func ExitError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s:", os.Args[0])

	if msg != "" {
		fmt.Fprintf(os.Stderr, " %s", msg)

		if err != nil {
			fmt.Fprintf(os.Stderr, ":")
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, " %s", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
