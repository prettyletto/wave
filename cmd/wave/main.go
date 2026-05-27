package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prettyletto/wave/internal/dispatch"
	"github.com/prettyletto/wave/internal/playerctl"
	"github.com/prettyletto/wave/internal/wpctl"
)

func main() {
	client := playerctl.New(nil)
	wpclient := wpctl.New(nil)
	dispatcher := dispatch.New(client, wpclient, os.Stdout)

	if err := dispatcher.Dispatch(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
