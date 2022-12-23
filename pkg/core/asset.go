package core

import (
	"errors"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) GetAssetByName(name string) ([]model.Asset, error) {
	//TODO
	return nil, errors.New("TODO")
}

func (c *Core) GetUserNodeAssets(userId, nodeId string,
	params model.PaginationParam) (res model.PaginationResponse, err error) {
	//TODO
	return model.PaginationResponse{}, errors.New("TODO")
}

func (c *Core) GetUserPermsAssets(userId string,
	params model.PaginationParam) (res model.PaginationResponse, err error) {
	//TODO
	return model.PaginationResponse{}, errors.New("TODO")
}

func (c *Core) GetAssetById(assetID string) (asset model.Asset, err error) {
	res, err := c.db.Query("Select * from ASSET WHERE id = ?", assetID)
	if err != nil {
		return asset, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		err = document.StructScan(d, &asset)
		return err
	})
	return asset, err
}

func (c *Core) GetAllUserPermsAssets(nodeIDs []string) ([]map[string]interface{}, model.NodeList, error) {
	nas, err := getNodes(c.db, nodeIDs)
	if err != nil {
		return nil, nil, err
	}

	aids := make([]string, 0, 10)
	for _, v := range nas {
		aids = append(aids, v.AssetIDs...)
	}

	ats, err := getAssets(c.db, aids)
	if err != nil {
		return nil, nil, err
	}

	res := make([]map[string]interface{}, 0, 10)
	for _, v := range ats {
		mp := make(map[string]interface{})
		// internal ID
		mp["id"] = v.ID
		mp["hostname"] = v.Hostname
		mp["ip"] = v.IP
		mp["platform"] = v.Platform
		mp["comment"] = v.Comment
		res = append(res, mp)
	}

	return res, model.NodeList(nas), err
}

func (c *Core) GetUserPermAssetsByIP(userId, assetIP string) (assets []model.Asset, err error) {
	// TODO
	return nil, errors.New("TODO")
}

func getAssets(db *genji.DB, assetIDs []string) ([]model.Asset, error) {
	res, err := db.Query("Select * from ASSET WHERE id IN ?", assetIDs)
	if err != nil {
		return nil, err
	}
	as := make([]model.Asset, 0, 5)
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var a model.Asset
		err = document.StructScan(d, &a)
		if err != nil {
			return err
		}
		as = append(as, a)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return as, nil
}
