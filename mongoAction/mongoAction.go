package mongoAction

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"gbGATEWAY/gbp"
	"gbGATEWAY/utils"
	"log"
	"strconv"
	"time"

	"github.com/fatih/color"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

type Mongo struct {
	Ctx            context.Context
	Client         *mongo.Client
	UserCollection *mongo.Collection
	MsgCollection  *mongo.Collection
}

func (m *Mongo) Init(mongoIP string, username string, password string) {
	var cred options.Credential
	cred.Username = username
	cred.Password = password

	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI("mongodb://" + mongoIP + ":27017").SetAuth(cred)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	uCollection := client.Database("Users").Collection("UserDatas")
	mCollection := client.Database("messages").Collection("userMsg")
	m.Ctx = ctx
	m.Client = client

	m.UserCollection = uCollection
	m.MsgCollection = mCollection
	color.Green("Mongo connected!")
	// fmt.Println("Mongo client connected!")
}

func (m *Mongo) InsertMsg(msg []byte) (string, string, error) {
	var msgp gbp.ChatPayload
	err := proto.Unmarshal(msg, &msgp)
	if err != nil {
		log.Println("[ProtoUNMError] : ", err.Error())
		return "", "", err
	}
	_id, _ := primitive.ObjectIDFromHex(msgp.Tid)
	ts := strconv.FormatInt(time.Now().UnixNano(), 10)
	mLoc := base64.StdEncoding.EncodeToString([]byte(ts))
	var mp gbp.MsgFormat
	mp.Msg = msgp.Msg
	mp.Sid = msgp.Sid
	mp.Mloc = mLoc
	_, err = m.MsgCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": _id},
		bson.M{"$set": bson.M{"msg." + mLoc: mp}},
	)
	if err != nil {
		log.Println("[MongoUpdateError] : ", err.Error())
		return "", "", err
	}
	return msgp.Tid, mLoc, nil
}

func (m *Mongo) GetMainKey(id string) (string, error) {
	cursor, err := m.UserCollection.Find(
		context.TODO(),
		bson.M{"msgid": id},
	)
	if err != nil {
		log.Println("[MONGOGETERROR] : ", err.Error())
	}
	var userd []utils.NEWUSER
	err = cursor.All(context.TODO(), &userd)
	if err != nil {
		log.Println("[MONGOCURSORERROR] : ", err.Error())
	}
	if len(userd) == 0 {
		return "", errors.New("no user found!")
	} else {
		return userd[0].MainKey, nil
	}

}

func DeleteMsg(mDB Mongo, Tid string, MsgId string) {
	_id, _ := primitive.ObjectIDFromHex(Tid)
	r, err := mDB.MsgCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": _id},
		bson.M{"$unset": bson.M{"msg." + MsgId: 1}},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.MatchedCount)
}

func (m *Mongo) GetMsg(id string) (*utils.TransportMsg, error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("P:::-> ", err.Error())
	}
	filter := bson.M{"_id": _id}
	cursor, err := m.MsgCollection.Find(m.Ctx, filter)
	if err != nil {
		return nil, err
	}
	var userMsg []utils.TransportMsg
	err = cursor.All(m.Ctx, &userMsg)
	if err != nil {
		log.Println("cursor error: ", err.Error())
	}
	if len(userMsg) > 0 {
		return &userMsg[0], nil
	} else {
		return nil, errors.New("no data is found")
	}
}

func (m *Mongo) UpdateMsgStatus(_id string, status int) (int, error) {
	id, _ := primitive.ObjectIDFromHex(_id)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"sts": status}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, err
	}
	i := result.MatchedCount
	return int(i), nil

}

func (m *Mongo) DeleteMsg(_id string) error {
	filter := bson.M{"_id": _id}
	_, err := m.UserCollection.DeleteOne(m.Ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) AddUser(user utils.NEWUSER) (string, error) {

	result, err := m.UserCollection.InsertOne(m.Ctx, user)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	fmt.Println("Adduser")
	id := result.InsertedID.(primitive.ObjectID)
	return id.String(), nil
}

func (m *Mongo) DeleteUser(filter primitive.M) error {
	_, err := m.UserCollection.DeleteOne(m.Ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mongo) DeleteUserByEmail(email string) error {
	filter := bson.M{"email": email}
	err := m.DeleteUser(filter)
	return err
}

func (m *Mongo) DeleteUserByPhoneNo(phoneno string) error {
	filter := bson.M{"phoneno": phoneno}
	err := m.DeleteUser(filter)
	return err
}

func (m *Mongo) UpdateUserName(id string, name string) (int64, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": bson.M{"name": name}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, nil
	}
	i := result.MatchedCount
	return i, nil
}

func (m *Mongo) UpdateUserAge(id int, age string) (int64, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"age": age}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, nil
	}
	i := result.MatchedCount
	return i, nil
}

func (m *Mongo) UpdateUserPhoneNo(id int, phoneno string) (int64, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"phone_no": phoneno}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, nil
	}
	i := result.MatchedCount
	return i, nil
}

func (m *Mongo) UpdateUserEmail(id int, email string) (int64, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"email": email}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, nil
	}
	i := result.MatchedCount
	return i, nil
}

func (m *Mongo) UpdateUserProfilePic(id int, profilePic string) (int64, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"profile_pic": profilePic}}
	result, err := m.UserCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return -1, nil
	}
	i := result.MatchedCount
	return i, nil
}

func (m *Mongo) ReadUserData(filter primitive.M) (*utils.NEWUSER, error) {
	cursor, err := m.UserCollection.Find(m.Ctx, filter)
	if err != nil {
		return nil, err
	}
	var userData []utils.NEWUSER
	cursor.All(m.Ctx, &userData)
	return &userData[0], nil
}

func (m *Mongo) ReadUserDataById(id string) (*utils.NEWUSER, error) {
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	udata, err := m.ReadUserData(filter)
	if err != nil {
		return nil, err
	}
	return udata, nil
}

func (m *Mongo) ReadUserDataByMNo(number string) (*utils.NEWUSER, error) {
	filter := bson.M{"phoneno": number}
	udata, err := m.ReadUserData(filter)
	if err != nil {
		return nil, err
	}
	return udata, nil
}

func (m *Mongo) ReadUserDataByEmail(email string) (*utils.NEWUSER, error) {
	filter := bson.M{"email": email}
	udata, err := m.ReadUserData(filter)
	if err != nil {
		return nil, err
	}
	return udata, nil
}
