package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/model"
)

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

func (c *Core) QueryAllAsset() ([]string, error) {
	assets, err := queryStructsFromDb[model.Asset](c.db, "SELECT * FROM ASSET")
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, 10)
	for _, v := range assets {
		ps := strings.Join(v.Protocols, ",")
		s := fmt.Sprintf("%4s|%10s|%10s|%39s|%10s|%6v|%10s|%15s|%s",
			v.ID, v.Name, v.Hostname, v.IP, v.Os, v.IsActive, v.Platform, ps, v.Comment)
		res = append(res, s)
	}
	return res, nil
}

func (c *Core) QueryAssetUserInfo() ([]string, error) {
	aus, err := queryStructsFromDb[model.AssetUserInfo](c.db, "SELECT * FROM ASSETUSERINFO")
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, 10)
	for _, v := range aus {
		var ea string
		ea = time.Unix(v.ExpireAt, 0).Format(common.LogFormat)
		if v.ExpireAt == 0 {
			ea = "9999-12-31 23:59:59"
		}
		si := strings.Join(v.SysUserID, ",")
		s := fmt.Sprintf("%4s|%7s|%8s|%s|%12v|%s",
			v.ID, v.UserID, v.AssetID, ea, v.NeedConfirm, si)
		res = append(res, s)
	}
	return res, nil
}
