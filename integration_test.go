// +build integration

package kat

import (
	"testing"

	"github.com/facebookgo/ensure"
)

const prodURL = "https://kickass.to/"

func TestIntegrateSingleResult(t *testing.T) {
	const target = "Ubuntu 14.10 Desktop 64bit ISO"
	client, err := NewClient(ClientRawURL(prodURL))
	ensure.Nil(t, err)
	actual, err := client.Search(target)
	ensure.Nil(t, err)
	ensure.Subset(t, actual, []Result{
		{
			Name:  "Ubuntu 14.10 Desktop 64bit ISO",
			Size:  "1.08 GB",
			Files: 1,
		},
	})
}

func TestIntegrateNoResults(t *testing.T) {
	client, err := NewClient(ClientRawURL(prodURL))
	ensure.Nil(t, err)
	actual, err := client.Search("8B924976-8B90-400C-84F8-70F7E1FDF617")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(actual), 0)
}
