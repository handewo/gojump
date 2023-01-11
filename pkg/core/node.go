package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) getNodes(nodeIDs []string) ([]model.Node, error) {
	v, err := c.db.QueryStructs(model.NodeType, "SELECT * from NODE WHERE id IN ?", nodeIDs)
	if err != nil {
		return nil, err
	}

	nodes, ok := v.([]model.Node)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return nodes, nil
}

func (c *Core) QueryAllNode() ([]string, error) {
	v, err := c.db.QueryStructs(model.NodeType, "SELECT * FROM NODE")
	if err != nil {
		return nil, err
	}

	nodes, ok := v.([]model.Node)
	if !ok {
		return nil, errors.New("invalid value type")
	}

	res := make([]string, 0, 10)
	for _, v := range nodes {
		as := strings.Join(v.AssetIDs, ",")
		s := fmt.Sprintf("%4s|%10s|%10s|%s", v.ID, v.Name, v.Key, as)
		res = append(res, s)
	}
	return res, nil
}
