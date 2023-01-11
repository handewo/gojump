package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/handewo/gojump/pkg/common"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) GetAssetById(assetID string) (model.Asset, error) {
	asset := model.Asset{}
	err := c.db.QueryStruct(&asset, "SELECT * from ASSET WHERE id = ?", assetID)
	return asset, err
}

func (c *Core) GetAllUserPermsAssets(nodeIDs []string) ([]map[string]interface{}, model.NodeList, error) {
	nas, err := c.getNodes(nodeIDs)
	if err != nil {
		return nil, nil, err
	}

	aids := make([]string, 0, 10)
	for _, v := range nas {
		aids = append(aids, v.AssetIDs...)
	}

	ats, err := c.getAssets(aids)
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

func (c *Core) getAssets(assetIDs []string) ([]model.Asset, error) {
	v, err := c.db.QueryStructs(model.AssetType, "SELECT * from ASSET WHERE id IN ?", assetIDs)
	if err != nil {
		return nil, err
	}

	assets, ok := v.([]model.Asset)
	if !ok {
		return nil, errors.New("invalid value type")
	}
	return assets, nil
}

func (c *Core) QueryAllAsset() ([]string, error) {
	v, err := c.db.QueryStructs(model.AssetType, "SELECT * FROM ASSET")
	if err != nil {
		return nil, err
	}

	assets, ok := v.([]model.Asset)
	if !ok {
		return nil, errors.New("invalid value type")
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
	v, err := c.db.QueryStructs(model.AssetUserInfoType, "SELECT * FROM ASSETUSERINFO")
	if err != nil {
		return nil, err
	}

	aus, ok := v.([]model.AssetUserInfo)
	if !ok {
		return nil, errors.New("invalid value type")
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
