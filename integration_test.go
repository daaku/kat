// +build integration

package kat

import (
	"testing"

	"github.com/facebookgo/ensure"
)

const prodURL = "http://kickass.cd/"

func TestIntegrateSingleResult(t *testing.T) {
	const target = "Ubuntu 16.04.1 LTS Desktop 64-bit"
	client, err := NewClient(ClientRawURL(prodURL))
	ensure.Nil(t, err)
	actual, err := client.Search(target)
	ensure.Nil(t, err)
	ensure.NotDeepEqual(t, len(actual), 0)
}

func TestIntegrateNoResults(t *testing.T) {
	client, err := NewClient(ClientRawURL(prodURL))
	ensure.Nil(t, err)
	actual, err := client.Search("8B924976-8B90-400C-84F8-70F7E1FDF617")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, len(actual), 0)
}
