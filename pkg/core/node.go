package core

import (
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/model"
)

func getNodes(db *genji.DB, nodeIDs []string) ([]model.Node, error) {
	res, err := db.Query("SELECT * from NODE WHERE id IN ?", nodeIDs)
	if err != nil {
		return nil, err
	}
	uas := make([]model.Node, 0, 5)
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var ua model.Node
		err = document.StructScan(d, &ua)
		if err != nil {
			return err
		}
		uas = append(uas, ua)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return uas, nil
}
