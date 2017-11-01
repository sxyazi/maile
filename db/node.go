package db

import (
	"errors"
	com "github.com/mlgaku/back/common"
	. "github.com/mlgaku/back/service"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Node struct {
	Id     bson.ObjectId `fill:"u" json:"id" bson:"_id,omitempty"`
	Name   string        `fill:"iu" json:"name" validate:"required,max=30,alphanum"`                            // 名字
	Title  string        `fill:"iu" json:"title" validate:"required,max=30"`                                    // 标题
	Sort   int64         `fill:"iu" json:"sort,omitempty" bson:",omitempty" validate:"omitempty,numeric"`       // 排序
	Desc   string        `fill:"iu" json:"desc,omitempty" bson:",omitempty" validate:"omitempty,min=5,max=300"` // 描述
	Parent bson.ObjectId `fill:"iu" json:"parent,omitempty" bson:",omitempty"`                                  // 父节点 ID
}

// 获得 Node 实例
func NewNode(body []byte, typ string) (*Node, error) {
	node := &Node{}
	if err := com.ParseJSON(body, typ, node); err != nil {
		panic(err)
	}

	return node, nil
}

// 添加
func (*Node) Add(db *Database, node *Node) error {
	if err := com.NewVali().Struct(node); err != "" {
		return errors.New(err)
	}

	return db.C("node").Insert(node)
}

// 保存
func (*Node) Save(db *Database, id bson.ObjectId, node *Node) error {
	if id == "" {
		return errors.New("节点ID不能为空")
	}

	set, err := com.Extract(node, "u")
	if err != nil {
		return err
	}

	return db.C("node").UpdateId(id, bson.M{"$set": set})
}

// 查找所有
func (*Node) FindAll(db *Database) (*[]Node, error) {
	node := &[]Node{}
	if err := db.C("node").Find(bson.M{}).All(node); err != nil {
		return nil, err
	}

	return node, nil
}

// 节点ID是否存在
func (*Node) IdExists(db *Database, id bson.ObjectId) (bool, error) {
	if c, err := db.C("node").FindId(id).Count(); err != nil {
		return false, err
	} else if c != 1 {
		return false, nil
	}

	return true, nil
}

// 节点名是否存在
func (*Node) NameExists(db *Database, name string) (bool, error) {
	if name == "" {
		return false, errors.New("节点名不能为空")
	}

	if c, err := db.C("node").Find(bson.M{"name": name}).Count(); err != nil {
		return false, err
	} else if c < 1 {
		return false, nil
	}

	return true, nil
}

// 是否有子节点存在
func (*Node) HasChild(db *Database, id bson.ObjectId) (bool, error) {
	if c, err := db.C("node").Find(bson.M{"parent": id}).Count(); err != nil {
		return false, err
	} else if c < 1 {
		return false, nil
	}

	return true, nil
}

// 节点下是否有主题存在
func (*Node) HasTopic(db *Database, id bson.ObjectId) (bool, error) {
	if c, err := db.C("topic").Find(bson.M{"node": id}).Count(); err != nil {
		return false, err
	} else if c < 1 {
		return false, nil
	}

	return true, nil
}

// 通过ID或名称查找
func (*Node) FindByIdOrName(db *Database, node *Node) error {
	var q *mgo.Query
	if node.Id != "" {
		q = db.C("node").FindId(node.Id)
	} else if node.Name != "" {
		q = db.C("node").Find(bson.M{"name": node.Name})
	} else {
		return errors.New("ID 或名称不能为空")
	}

	return q.One(node)
}

func (*Node) RemoveById(db *Database, id bson.ObjectId) error {
	if id == "" {
		return errors.New("ID 不能为空")
	}

	return db.C("node").RemoveId(id)
}
