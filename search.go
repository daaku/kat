// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/daaku/kat"
	"github.com/pkg/errors"
)

func Main() error {
	katURL := flag.String("url", "http://kickass.cd/", "kat URL")
	query := flag.String("query", "", "query string to search for")
	flag.Parse()

	client, err := kat.NewClient(kat.ClientRawURL(*katURL))
	if err != nil {
		return errors.Wrap(err, "invalid client")
	}
	results, err := client.Search(*query)
	if err != nil {
		return errors.Wrap(err, "search failed")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
			r.Name, r.Size, r.Age, r.Seed, r.Leech)
	}
	w.Flush()
	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
