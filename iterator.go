package smt

import (
	"bytes"
	"errors"
)

var (
	ErrIteratorEnd  = errors.New("iterator end")
	ErrNodeNotFound = errors.New("failed to fetch the node")
)

// Iterator is a key value iterator that traverses the tree
type Iterator interface {
	Next() bool
	Leaf() bool
	NodeBlob() []byte
	LeafKey() []byte
	LeafValue() ([]byte, error)
}

type iteratorState struct {
	hash         []byte
	rawNode      []byte
	leftVisited  bool
	rightVisited bool
}

// iterator traverses the sparse merkle tree pre-order, this implements Iterator interface
type iterator struct {
	trie  *SparseMerkleTree
	stack []*iteratorState
	err   error
}

// init initializes the iterator with root node
func (i *iterator) init() (*iteratorState, error) {
	node, err := i.trie.nodes.Get(i.trie.root)
	if err != nil {
		return nil, err
	}

	return &iteratorState{
		hash:         i.trie.root,
		rawNode:      node,
		leftVisited:  false,
		rightVisited: false,
	}, nil
}

// Next moves the iterator by one step
func (i *iterator) Next() bool {
	if i.err != nil {
		return false
	}

	node, err := i.next(true)
	i.err = err

	if err != nil {
		return false
	}

	i.push(node)

	return true
}

func (i *iterator) next(descend bool) (*iteratorState, error) {
	// initialize with root node if stack is empty
	if len(i.stack) == 0 {
		state, err := i.init()
		if err != nil {
			return nil, err
		}

		return state, nil
	}

	if !descend {
		i.pop()
	}

	for len(i.stack) > 0 {
		parent := i.stack[len(i.stack)-1]

		// find the next child node
		child, shouldPop, err := i.nextChild(parent)
		if err != nil {
			return nil, err
		}

		if !shouldPop {
			return child, nil
		}

		// pop since no child is available
		i.pop()
	}

	return nil, ErrIteratorEnd
}

func (i *iterator) pop() {
	i.stack[len(i.stack)-1] = nil
	i.stack = i.stack[:len(i.stack)-1]
}

func (i *iterator) push(node *iteratorState) {
	i.stack = append(i.stack, node)
}

// nextChild returns the next child node, follows the pre-order traversal.
func (i *iterator) nextChild(nodeState *iteratorState) (child *iteratorState, pop bool, err error) {
	if i.trie.th.isLeaf(nodeState.rawNode) {
		return nodeState, true, nil
	}

	leftChild, rightChild := i.trie.th.parseNode(nodeState.rawNode)

	if !nodeState.leftVisited {
		// Check for empty left subtree
		if !bytes.Equal(i.trie.th.zeroValue, leftChild) {
			leftNode, err := i.trie.nodes.Get(leftChild)
			if err != nil {
				return nil, false, ErrNodeNotFound
			}

			nodeState.leftVisited = true

			return &iteratorState{
				hash:         leftChild,
				rawNode:      leftNode,
				leftVisited:  false,
				rightVisited: false,
			}, false, nil
		} else {
			nodeState.leftVisited = true
		}
	}

	if !nodeState.rightVisited {
		// Check for empty left subtree
		if !bytes.Equal(i.trie.th.zeroValue, rightChild) {
			rightNode, err := i.trie.nodes.Get(rightChild)
			if err != nil {
				return nil, false, ErrNodeNotFound
			}

			nodeState.rightVisited = true

			return &iteratorState{
				hash:         rightChild,
				rawNode:      rightNode,
				leftVisited:  false,
				rightVisited: false,
			}, false, nil
		} else {
			nodeState.rightVisited = true
		}
	}

	return nodeState, true, nil
}

// Leaf returns true iff the current node is a leaf node
func (i *iterator) Leaf() bool {
	node := i.stack[len(i.stack)-1]

	return i.trie.th.isLeaf(node.rawNode)
}

// NodeBlob returns the raw node
func (i *iterator) NodeBlob() []byte {
	return i.stack[len(i.stack)-1].rawNode
}

// LeafKey returns the hashed key associated with the leaf node
func (i *iterator) LeafKey() []byte {
	leaf := i.stack[len(i.stack)-1]

	key, _ := i.trie.th.parseLeaf(leaf.rawNode)

	return key
}

// LeafValue returns the actual value of the leaf node
func (i *iterator) LeafValue() ([]byte, error) {
	leaf := i.stack[len(i.stack)-1]

	path, _ := i.trie.th.parseLeaf(leaf.rawNode)

	return i.trie.values.Get(path)
}
