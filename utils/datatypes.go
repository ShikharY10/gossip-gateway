package utils

type NEWUSER struct {
	MsgId      string `bson:"msgid,omitempty"`
	Name       string `bson:"name,omitempty"`
	Age        string `bson:"age,omitempty"`
	PhoneNo    string `bson:"phone_no,omitempty"`
	Email      string `bson:"email,omitempty"`
	ProfilePic string `bson:"profile_pic,omitempty"`
	MainKey    string `bson:"main_key,omitempty"`
	Gender     string `bson:"gender,omitempty"`
	Password   string `bson:"password,omitempty"`
}

type MsgFormat struct {
	Sid string `bson:"snum,omitempty"`
	Msg string `bsin:"msg,omitempty"`
}

type TransportMsg struct {
	Tid string `json:"tid"`
	Msg []byte `json:"msg"`
}

type EngineName struct {
	Names []string `json:"names"`
}
