package reflect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenizer(t *testing.T) {
	require.Equal(t, []string{"", "foo"}, tokenize("/foo"))
	require.Equal(t, []string{"", "foo", "bar", "baz"}, tokenize("/foo/bar/baz"))
	require.Equal(t, []string{"foo", "bar", "baz"}, tokenize("foo/bar/baz"))
	require.Equal(t, []string{"foo"}, tokenize("foo"))

	// with quoting to support azure rm type names: e.g. Microsoft.Network/virtualNetworks
	require.Equal(t, []string{"", "foo"}, tokenize("/'fo'o"))
	require.Equal(t, []string{"", "foo/bar", "baz"}, tokenize("/'foo/bar'/baz"))
	require.Equal(t, []string{"foo", "bar/baz"}, tokenize("foo/'bar/baz'"))
	require.Equal(t, []string{"foo"}, tokenize("'foo'"))
}
