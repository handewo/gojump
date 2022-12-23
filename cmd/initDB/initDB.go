package main

import (
	"flag"

	"github.com/genjidb/genji"
	"github.com/handewo/gojump/pkg/initDB"
	"github.com/handewo/gojump/pkg/log"
)

var path string

func init() {
	flag.StringVar(&path, "d", "gojumpdb", "database path")
}
func main() {
	db, err := genji.Open(path)
	if err != nil {
		log.Fatal.Fatal(err)
	}

	initDB.InitSchema(db)
	initDB.InitData(db)
	initDB.InsertTestData(db)
	log.Info.Print("Init database finished")
}
