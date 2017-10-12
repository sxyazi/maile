package module

import (
	"encoding/json"
	com "github.com/mlgaku/back/common"
	. "github.com/mlgaku/back/service"
	. "github.com/mlgaku/back/types"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name     string        `json:"name" validate:"required,min=4,max=15,alphanum"`
	Email    string        `json:"email" validate:"required,min=8,max=30,email"`
	Password string        `json:"password,omitempty" validate:"required,min=8,max=20,alphanum"`

	RegIP   string `json:"reg_ip,omitempty" bson:"reg_ip"`
	RegTime int64  `json:"reg_time,omitempty" bson:"reg_time"`
}

func (*User) parse(body []byte) (*User, error) {
	user := &User{}
	return user, json.Unmarshal(body, user)
}

// 注册
func (u *User) Reg(db *Database, req *Request, conf *Config) Value {
	user, _ := u.parse(req.Body)

	if err := com.NewVali().Struct(user); err != "" {
		return &Fail{Msg: err}
	}

	user.Password = com.Sha1(user.Password, conf.Secret.Salt)
	user.RegIP, _ = com.IPAddr(req.Client.Http.RemoteAddr)
	user.RegTime = time.Now().Unix()

	if err := db.C("user").Insert(user); err != nil {
		return &Fail{Msg: err.Error()}
	}

	return &Succ{}
}

// 登录
func (u *User) Login(db *Database, req *Request, ses *Session, conf *Config) Value {
	user, _ := u.parse(req.Body)

	err := com.NewVali().Each(com.Iter(user.Name, user.Password), []string{"required"})
	if err != "" {
		return &Fail{Msg: err}
	}

	if _, ok := u.Check(db, req).(*Succ); ok {
		return &Fail{Msg: "用户名不存在"}
	}

	result := &User{}
	if err := db.C("user").Find(bson.M{"name": user.Name}).One(result); err != nil {
		return &Fail{Msg: "未知错误"}
	}

	if result.Password != com.Sha1(user.Password, conf.Secret.Salt) {
		return &Fail{Msg: "用户名与密码不匹配"}
	}

	// 保存状态
	ses.Set("user_id", result.Id)
	ses.Set("user_name", result.Name)
	ses.Set("user_email", result.Email)

	return &Succ{Data: &User{
		Id:    result.Id,
		Name:  result.Name,
		Email: result.Email,
	}}
}

// 检查用户名是否已被注册
func (u *User) Check(db *Database, req *Request) Value {
	user, _ := u.parse(req.Body)

	if err := com.NewVali().Var(user.Name, "required"); err != "" {
		return &Fail{Msg: err}
	}

	if c, _ := db.C("user").Find(bson.M{"name": user.Name}).Count(); c > 0 {
		return &Fail{Msg: "用户名已存在"}
	}

	return &Succ{}
}
