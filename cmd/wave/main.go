package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prettyletto/wave/internal/dispatch"
	"github.com/prettyletto/wave/internal/playerctl"
)

func main() {
	client := playerctl.New(nil)
	dispatcher := dispatch.New(client, os.Stdout)

	if err := dispatcher.Dispatch(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
