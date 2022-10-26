package smt

import (
	"crypto/sha256"
	"github.com/stretchr/testify/require"
	"testing"
)

type kvs struct{ k, v string }

var testdata = []kvs{
	{"aardvark", "c"},
	{"bar", "b"},
	{"barb", "bd"},
	{"bars", "be"},
	{"fab", "z"},
	{"foo", "a"},
	{"foos", "aa"},
	{"food", "ab"},
	{"jars", "d"},
}

func TestIterator_LeafNodes(t *testing.T) {
	nodeStore := NewSimpleMap()
	valueStore := NewSimpleMap()

	// Initialise the tree
	tree := NewSparseMerkleTree(nodeStore, valueStore, sha256.New())

	// Add test data to tree
	for _, entry := range testdata {
		_, err := tree.Update([]byte(entry.k), []byte(entry.v))
		require.NoError(t, err)
	}

	it := tree.NewIterator()

	leafNodes := make(map[string]string)

	for it.Next() {
		if it.Leaf() {
			value, err := it.LeafValue()
			require.NoError(t, err)

			leafNodes[string(it.LeafKey())] = string(value)
		}
	}

	for _, entry := range testdata {
		value, ok := leafNodes[string(tree.th.path([]byte(entry.k)))]
		// check for the key
		require.True(t, ok, "key not found")
		// check for the value
		require.Equal(t, value, entry.v)
	}
}
