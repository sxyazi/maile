package common

import "gopkg.in/mgo.v2"

type Database struct {
	Session *mgo.Session
}
