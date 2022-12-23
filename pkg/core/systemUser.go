package core

import (
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/handewo/gojump/pkg/model"
)

func (c *Core) GetSystemUsersByUserIdAndAssetId(userID string, assetID string) (sysUsers []model.SystemUser, err error) {
	sysIDs := make([]string, 0, 5)
	err = queryOneFieldFromDb(c.db, &sysIDs, "SELECT sysuserid FROM ASSETUSERINFO where userid = ? and assetid = ?", userID, assetID)
	if err != nil {
		return nil, err
	}

	sysUsers, err = getSystemUser(c.db, sysIDs)
	return
}

func getSystemUser(db *genji.DB, ids []string) ([]model.SystemUser, error) {
	sysUsers := make([]model.SystemUser, 0, 5)
	res, err := db.Query("Select * from SYSTEMUSER WHERE id IN ?", ids)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	err = res.Iterate(func(d types.Document) error {
		var su model.SystemUser
		err = document.StructScan(d, &su)
		sysUsers = append(sysUsers, su)
		return err
	})
	return sysUsers, err
}
