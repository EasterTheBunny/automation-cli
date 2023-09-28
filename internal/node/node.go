package node

import (
	"fmt"
	"io"

	"github.com/docker/docker/client"
)

type ChainlinkNode struct {
	Name           string
	Network        string
	PostgresImage  string
	ChainlinkImage string
	GroupName      string
	Address        string

	client    *client.Client
	writer    io.Writer
	postgres  nodeContainer
	chainlink nodeContainer
}

func (n *ChainlinkNode) URL() string {
	return fmt.Sprintf("http://localhost:%d", n.chainlink.port)
}
