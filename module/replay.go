package module

import (
	"encoding/json"
	"github.com/mlgaku/back/db"
	. "github.com/mlgaku/back/service"
	. "github.com/mlgaku/back/types"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Replay struct {
	Db db.Replay
}

func (*Replay) parse(body []byte) (*db.Replay, error) {
	replay := &db.Replay{}
	return replay, json.Unmarshal(body, replay)
}

// 添加新回复
func (r *Replay) New(bd *Database, ps *Pubsub, ses *Session, req *Request) Value {
	replay, _ := r.parse(req.Body)
	replay.Author = ses.Get("user_id").(bson.ObjectId)

	topic := &db.Topic{}
	if err := topic.Find(bd, replay.Topic, topic); err != nil {
		return &Fail{Msg: err.Error()}
	}

	// 添加回复
	if err := r.Db.Add(bd, replay); err != nil {
		return &Fail{Msg: err.Error()}
	}

	// 更新最后回复
	topic.UpdateReplay(bd, replay.Topic, ses.Get("user_name").(string))

	// 回复人不是主题作者时添加通知
	if replay.Author != topic.Author {
		new(db.Notice).Add(bd, &db.Notice{
			Type:       1,
			Time:       time.Now(),
			Master:     topic.Author,
			User:       ses.Get("user_name").(string),
			TopicID:    replay.Topic,
			TopicTitle: topic.Title,
		})
	}

	ps.Publish(&Prot{Mod: "replay", Act: "list"})
	ps.Publish(&Prot{Mod: "notice", Act: "list"})
	return &Succ{}
}

// 获取回复列表
func (r *Replay) List(db *Database, req *Request) Value {
	var s struct {
		Page  int
		Topic bson.ObjectId
	}
	if err := json.Unmarshal(req.Body, &s); err != nil {
		return &Fail{Msg: err.Error()}
	}

	replay, err := r.Db.Paginate(db, s.Topic, s.Page)
	if err != nil {
		return &Fail{Msg: err.Error()}
	}

	return &Succ{Data: replay}
}
