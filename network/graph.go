package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type IP string

type Graph interface {
	AddNode(node IP, id string) bool
	AddEdge(src IP, target IP) error
	EncodeDOT() ([]byte, error)
}

type network struct {
	sync.Mutex
	graph map[IP]*node
}

type node struct {
	Id       string       `json:"id"`
	Outgoing map[IP]*edge `json:"outgoing"`
}

type edge struct {
	Count       uint         `json:"count"`
	StatusCodes map[int]uint `json:"status_codes"`
}

func NewGraph() Graph {
	return &network{
		graph: make(map[IP]*node),
	}
}

// returns true if node was added, false if it already exists
func (n *network) AddNode(src IP, id string) bool {
	n.Lock()
	defer n.Unlock()

	if node, ok := n.graph[src]; ok {
		if node.Id == "" {
			node.Id = id
		}
		return false
	}

	n.graph[src] = &node{Id: id, Outgoing: make(map[IP]*edge)}
	return true
}

// returns an error if src or target is not in the graph
func (n *network) AddEdge(src IP, target IP) error {
	n.Lock()
	defer n.Unlock()

	if _, ok := n.graph[src]; !ok {
		return errors.New(fmt.Sprintf("source IP %s not found", src))
	}
	if _, ok := n.graph[target]; !ok {
		return errors.New(fmt.Sprintf("target IP %s not found", target))
	}

	node := n.graph[src]
	if e, ok := node.Outgoing[target]; ok {
		e.Count++
	} else {
		node.Outgoing[target] = &edge{Count: 1}
	}

	return nil
}

func (n *network) generateDOTGraph(buffer *bytes.Buffer) (*bytes.Buffer, error) {
	var err error
	n.Lock()
	defer n.Unlock()

	for k, v := range n.graph {
		if v.Id != "" {
			_, err = buffer.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\"];\n", k, v.Id))
			if err != nil {
				return nil, err
			}
		}

		buffer, err = v.generateDOTGraph(k, buffer)
		if err != nil {
			return nil, err
		}
	}
	return buffer, nil
}

func (n *network) EncodeDOT() ([]byte, error) {
	buffer := bytes.NewBufferString("digraph cfnetwork {\n")
	dotEncoding, err := n.generateDOTGraph(buffer)
	if err != nil {
		return []byte{}, err
	}

	_, err = dotEncoding.WriteString("}\n")
	if err != nil {
		return []byte{}, err
	}

	return dotEncoding.Bytes(), nil
}

func (n *network) MarshalJSON() ([]byte, error) {
	n.Lock()
	defer n.Unlock()
	return json.Marshal(n.graph)
}

func (n *node) generateDOTGraph(nodeName IP, buffer *bytes.Buffer) (*bytes.Buffer, error) {
	for k, e := range n.Outgoing {
		_, err := buffer.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%d\"];", nodeName, k, e.Count))
		if err != nil {
			return nil, err
		}
		_, err = buffer.WriteRune('\n')
		if err != nil {
			return nil, err
		}
	}
	return buffer, nil
}
