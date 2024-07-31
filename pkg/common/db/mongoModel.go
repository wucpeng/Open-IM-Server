package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	pbMsg "Open_IM/pkg/proto/msg"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/sortkeys"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"sync"
	"time"
)

const cChat = "msg"
const singleGocMsgNum = 5000

func GetSingleGocMsgNum() int {
	return singleGocMsgNum
}

type MsgInfo struct {
	SendTime int64
	Msg      []byte
}

type UserChat struct {
	UID string
	//ListIndex int `bson:"index"`
	Msg []MsgInfo
}
type UserChat2 struct {
	ID  string `bson:"_id"`
	UID string
	//ListIndex int `bson:"index"`
	Msg []MsgInfo
}

type GroupMember_x struct {
	GroupID string
	UIDList []string
}

// deleteMsgByLogic
func (d *DataBases) DelMsgBySeqList(userID string, seqList []uint32, operationID string) (totalUnexistSeqList []uint32, err error) {
	//log.Debug(operationID, utils.GetSelfFuncName(), "args ", userID, seqList)
	sortkeys.Uint32s(seqList)
	suffixUserID2SubSeqList := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(userID, seqList)
	//log.Info(operationID, "DelMsgBySeqList:", suffixUserID2SubSeqList)
	lock := sync.Mutex{}
	var wg sync.WaitGroup
	wg.Add(len(suffixUserID2SubSeqList))
	for k, v := range suffixUserID2SubSeqList {
		go func(suffixUserID string, subSeqList []uint32, operationID string) {
			defer wg.Done()
			unexistSeqList, err := d.DelMsgBySeqListInOneDoc(suffixUserID, subSeqList, operationID)
			if err != nil {
				log.Error(operationID, "DelMsgBySeqListInOneDoc failed ", err.Error(), suffixUserID, subSeqList)
				return
			}
			lock.Lock()
			totalUnexistSeqList = append(totalUnexistSeqList, unexistSeqList...)
			lock.Unlock()
		}(k, v, operationID)
	}
	//log.Info(operationID, "DelMsgBySeqList:", totalUnexistSeqList)
	return totalUnexistSeqList, err
}

func (d *DataBases) DelMsgBySeqListInOneDoc(suffixUserID string, seqList []uint32, operationID string) ([]uint32, error) {
	//log.Info(operationID, utils.GetSelfFuncName(), "args ", suffixUserID, seqList)
	seqMsgList, indexList, unexistSeqList, err := d.GetMsgAndIndexBySeqListInOneMongo2(suffixUserID, seqList, operationID)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	for i, v := range seqMsgList {
		if err := d.ReplaceMsgByIndex(suffixUserID, v, operationID, indexList[i]); err != nil {
			return nil, utils.Wrap(err, "")
		}
	}
	return unexistSeqList, nil
}

func (d *DataBases) ReplaceMsgByIndex(suffixUserID string, msg *open_im_sdk.MsgData, operationID string, seqIndex int) error {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), suffixUserID, *msg)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	//log.NewDebug(operationID, utils.GetSelfFuncName(), seqIndex, s)
	msg.Status = constant.MsgDeleted
	bytes, err := proto.Marshal(msg)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto marshal failed ", err.Error(), msg.String())
		return utils.Wrap(err, "")
	}
	s := fmt.Sprintf("msg.%d.msg", seqIndex)
	updateResult, err := c.UpdateOne(ctx, bson.M{"uid": suffixUserID}, bson.M{"$set": bson.M{s: bytes}})
	log.NewDebug(operationID, utils.GetSelfFuncName(), updateResult)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "UpdateOne", err.Error())
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) ReplaceMsgBySeq(uid string, msg *open_im_sdk.MsgData, operationID string) error {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid, *msg)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	uid = getSeqUid(uid, msg.Seq)
	seqIndex := getMsgIndex(msg.Seq)
	s := fmt.Sprintf("msg.%d.msg", seqIndex)
	//log.NewDebug(operationID, utils.GetSelfFuncName(), seqIndex, s)
	bytes, err := proto.Marshal(msg)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto marshal", err.Error())
		return utils.Wrap(err, "")
	}

	updateResult, err := c.UpdateOne(
		ctx, bson.M{"uid": uid},
		bson.M{"$set": bson.M{s: bytes}})
	log.NewDebug(operationID, utils.GetSelfFuncName(), updateResult)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "UpdateOne", err.Error())
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) GetMsgBySeqList(uid string, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid, seqList)
	var hasSeqList []uint32
	singleCount := 0
	session := d.mgoSession.Clone()
	if session == nil {
		return nil, errors.New("session == nil")
	}
	defer session.Close()
	c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(uid, seqList)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.Find(bson.M{"uid": seqUid}).One(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, value, uid, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, uid, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionMessageBySeqList(diff)
		seqMsg = append(seqMsg, exceptionMSg...)

	}
	return seqMsg, nil
}

func (d *DataBases) GetUserMsgListByIndex(ID string, index int64) (*UserChat, error) { //cron_task/clear_msg.go
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	regex := fmt.Sprintf("^%s", ID)
	findOpts := options.Find().SetLimit(1).SetSkip(index).SetSort(bson.M{"uid": 1})
	var msgs []UserChat
	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
	if err != nil {
		return nil, err
	}
	err = cursor.Decode(&msgs)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	if len(msgs) > 0 {
		return &msgs[0], err
	} else {
		return nil, errors.New("get msg list failed")
	}
}

func (d *DataBases) DelMongoMsgs(IDList []string) error { //cron_task/clear_msg.go
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	_, err := c.DeleteMany(ctx, bson.M{"uid": bson.M{"$in": IDList}})
	return err
}

func (d *DataBases) ReplaceMsgToBlankByIndex(suffixID string, index int) error { //cron_task/clear_msg.go
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	userChat := &UserChat{}
	err := c.FindOne(ctx, bson.M{"uid": suffixID}).Decode(&userChat)
	if err != nil {
		return err
	}
	for i, msg := range userChat.Msg {
		if i <= index {
			msg.Msg = nil
			msg.SendTime = 0
		}
	}
	_, err = c.UpdateOne(ctx, bson.M{"uid": suffixID}, userChat)
	return err
}

func (d *DataBases) GetNewestMsg(ID string) (msg *MsgInfo, err error) { //cron-clear_msg.go
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	regex := fmt.Sprintf("^%s", ID)
	findOpts := options.Find().SetLimit(1).SetSort(bson.M{"uid": -1})
	var userChats []UserChat
	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
	if err != nil {
		return nil, err
	}
	err = cursor.Decode(&userChats)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	if len(userChats) > 0 {
		if len(userChats[0].Msg) > 0 {
			return &userChats[0].Msg[len(userChats[0].Msg)], nil
		}
		return nil, errors.New("len(userChats[0].Msg) < 0")
	}
	return nil, errors.New("len(userChats) < 0")
}

func (d *DataBases) GetMsgBySeqListMongo2(uid string, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) { //pull_message.go
	var hasSeqList []uint32
	singleCount := 0
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)

	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(uid, seqList)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, value, uid, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, uid, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionMessageBySeqList(diff)
		seqMsg = append(seqMsg, exceptionMSg...)
	}
	return seqMsg, nil
}

func (d *DataBases) GetMsgById(uid string, clientMsgId string, operationID string) (seqMsg *open_im_sdk.MsgData, err error) {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
	maxSeq, err := d.GetUserMaxSeq(uid)
	if err == redis.Nil {
		return seqMsg, nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid,  maxSeq, seqUsers)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)

	sChat := UserChat{}
	for _, seqUid := range seqUsers {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, uid, err.Error())
			continue
		}
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, err.Error())
				return nil, err
			}
			if msg.ClientMsgID == clientMsgId {
				return msg, err
			}
		}
	}
	return seqMsg, nil
}

func (d *DataBases) GetGroupAllMsgList(uid string, groupID string, startTime int64, endTime int64, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid, groupID)
	maxSeq, err := d.GetUserMaxSeq(uid)
	if err == redis.Nil {
		return seqMsg, nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)

	sChat := UserChat{}
	for _, seqUid := range seqUsers {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, uid, err.Error())
			continue
		}
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, err.Error())
				return nil, err
			}
			if groupID != msg.GroupID {
				continue
			}
			//log.NewDebug(operationID, utils.GetSelfFuncName(), msg.ContentType, msg.SendTime, msg.Status, msg.GroupID, string(msg.Content))
			if msg.Status == constant.MsgDeleted {
				continue
			}
			if startTime != 0 && msg.SendTime < startTime {
				continue
			}
			if endTime != 0 && msg.SendTime > endTime {
				break
			}
			if msg.ContentType == constant.Text || msg.ContentType == constant.Custom || msg.ContentType == constant.AtText || msg.ContentType == constant.AdvancedRevoke {
				// || (msg.ContentType > 1500 && msg.ContentType < 1600) {
				//log.NewError(operationID, utils.GetSelfFuncName(), msg.ContentType, msg.SendTime, string(msg.Content))
				seqMsg = append(seqMsg, msg)
			}
		}
	}
	//log.NewInfo(operationID, utils.GetSelfFuncName(), len(seqMsg))
	return seqMsg, nil
}

func (d *DataBases) GetSingleAllMsgList(uid string, userId string, startTime int64, endTime int64, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
	maxSeq, err := d.GetUserMaxSeq(uid)
	if err == redis.Nil {
		return seqMsg, nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
	//log.NewInfo(operationID, utils.GetSelfFuncName(), uid,  maxSeq, seqUsers)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)

	sChat := UserChat{}
	for _, seqUid := range seqUsers {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, uid, err.Error())
			continue
		}
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, err.Error())
				return nil, err
			}
			if msg.Status == constant.MsgDeleted {
				continue
			}
			if msg.SessionType != constant.SingleChatType {
				continue
			}
			if startTime != 0 && msg.SendTime < startTime {
				continue
			}
			if endTime != 0 && msg.SendTime > endTime {
				break
			}
			isSingle := (msg.SendID == userId && msg.RecvID == uid) || (msg.SendID == uid && msg.RecvID == userId)
			if !isSingle {
				continue
			}
			if msg.ContentType == constant.Text || msg.ContentType == constant.Custom || msg.ContentType == constant.AtText || msg.ContentType == constant.AdvancedRevoke {
				//log.NewInfo(operationID, utils.GetSelfFuncName(), string(msg.Content))
				seqMsg = append(seqMsg, msg)
			}
		}
	}
	return seqMsg, nil
}

func (d *DataBases) GetSuperGroupMsgBySeqListMongo(groupID string, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	var hasSeqList []uint32
	singleCount := 0
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)

	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(groupID, seqList)
	//log.NewInfo(operationID, "GetSuperGroupMsgBySeqListMongo", groupID, seqList, m)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqGroupID", seqUid, value, groupID, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, groupID, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionSuperGroupMessageBySeqList(diff, groupID)
		seqMsg = append(seqMsg, exceptionMSg...)
	}
	return seqMsg, nil
}

func (d *DataBases) GetMsgAndIndexBySeqListInOneMongo2(suffixUserID string, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, indexList []int, unexistSeqList []uint32, err error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	sChat := UserChat{}
	if err = c.FindOne(ctx, bson.M{"uid": suffixUserID}).Decode(&sChat); err != nil {
		log.NewError(operationID, "not find seqUid", suffixUserID, err.Error())
		return nil, nil, nil, utils.Wrap(err, "")
	}
	singleCount := 0
	var hasSeqList []uint32
	for i := 0; i < len(sChat.Msg); i++ {
		msg := new(open_im_sdk.MsgData)
		if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
			log.NewError(operationID, "Unmarshal err", msg.String(), err.Error())
			return nil, nil, nil, err
		}
		if isContainInt32(msg.Seq, seqList) {
			indexList = append(indexList, i)
			seqMsg = append(seqMsg, msg)
			hasSeqList = append(hasSeqList, msg.Seq)
			singleCount++
			if singleCount == len(seqList) {
				break
			}
		}
	}
	for _, i := range seqList {
		if isContainInt32(i, hasSeqList) {
			continue
		}
		unexistSeqList = append(unexistSeqList, i)
	}
	return seqMsg, indexList, unexistSeqList, nil
}

func genExceptionMessageBySeqList(seqList []uint32) (exceptionMsg []*open_im_sdk.MsgData) {
	for _, v := range seqList {
		msg := new(open_im_sdk.MsgData)
		msg.Seq = v
		exceptionMsg = append(exceptionMsg, msg)
	}
	return exceptionMsg
}

func genExceptionSuperGroupMessageBySeqList(seqList []uint32, groupID string) (exceptionMsg []*open_im_sdk.MsgData) {
	for _, v := range seqList {
		msg := new(open_im_sdk.MsgData)
		msg.Seq = v
		msg.GroupID = groupID
		msg.SessionType = constant.SuperGroupChatType
		exceptionMsg = append(exceptionMsg, msg)
	}
	return exceptionMsg
}

func (d *DataBases) SaveUserChat(uid string, sendTime int64, m *pbMsg.MsgDataToDB) error {
	var seqUid string
	//newTime := getCurrentTimestampByMill()
	session := d.mgoSession.Clone()
	if session == nil {
		return errors.New("session == nil")
	}
	defer session.Close()
	//log.NewDebug("", "get mgoSession cost time", getCurrentTimestampByMill()-newTime)
	c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	seqUid = getSeqUid(uid, m.MsgData.Seq)
	n, err := c.Find(bson.M{"uid": seqUid}).Count()
	if err != nil {
		return err
	}
	//log.NewDebug("", "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
	sMsg := MsgInfo{}
	sMsg.SendTime = sendTime
	if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
		return err
	}
	if n == 0 {
		sChat := UserChat{}
		sChat.UID = seqUid
		sChat.Msg = append(sChat.Msg, sMsg)
		err = c.Insert(&sChat)
		if err != nil {
			return err
		}
	} else {
		err = c.Update(bson.M{"uid": seqUid}, bson.M{"$push": bson.M{"msg": sMsg}})
		if err != nil {
			return err
		}
	}
	//log.NewDebug("", "insert mgo data cost time", getCurrentTimestampByMill()-newTime)
	return nil
}

func (d *DataBases) DelUserChatMongo2(uid string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	filter := bson.M{"uid": uid}

	delTime := time.Now().Unix() - int64(config.Config.Mongo.DBRetainChatRecords)*24*3600
	if _, err := c.UpdateOne(ctx, filter, bson.M{"$pull": bson.M{"msg": bson.M{"sendtime": bson.M{"$lte": delTime}}}}); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) CleanUpUserMsgFromMongo(userID string, operationID string) error {
	ctx := context.Background()
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	maxSeq, err := d.GetUserMaxSeq(userID)
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return utils.Wrap(err, "")
	}

	seqUsers := getSeqUserIDList(userID, uint32(maxSeq))
	//log.NewInfo(operationID, "getSeqUserIDList", seqUsers)
	_, err = c.DeleteMany(ctx, bson.M{"uid": bson.M{"$in": seqUsers}})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return utils.Wrap(err, "")
}

// const cTag = "tag"
// const cSendLog = "send_log"
// const cWorkMoment = "work_moment"
// const cSuperGroup = "super_group"
// const cUserToSuperGroup = "user_to_super_group"
const cGroup = "group"
const cCommentMsg = "comment_msg"

func getCurrentTimestampByMill() int64 {
	return time.Now().UnixNano() / 1e6
}

func getSeqUid(uid string, seq uint32) string {
	seqSuffix := seq / singleGocMsgNum
	return indexGen(uid, seqSuffix)
}

func GetSeqUid(uid string, seq uint32) string {
	return getSeqUid(uid, seq)
}

func getSeqUserIDList(userID string, maxSeq uint32) []string {
	seqMaxSuffix := maxSeq / singleGocMsgNum
	var seqUserIDList []string
	for i := 0; i <= int(seqMaxSuffix); i++ {
		seqUserID := indexGen(userID, uint32(i))
		seqUserIDList = append(seqUserIDList, seqUserID)
	}
	return seqUserIDList
}

func getSeqSuperGroupID(groupID string, seq uint32) string {
	seqSuffix := seq / singleGocMsgNum
	return superGroupIndexGen(groupID, seqSuffix)
}

func getMsgIndex(seq uint32) int {
	seqSuffix := seq / singleGocMsgNum
	var index uint32
	if seqSuffix == 0 {
		index = (seq - seqSuffix*singleGocMsgNum) - 1
	} else {
		index = seq - seqSuffix*singleGocMsgNum
	}
	return int(index)
}

func isContainInt32(target uint32, List []uint32) bool {
	for _, element := range List {
		if target == element {
			return true
		}
	}
	return false
}

func indexGen(uid string, seqSuffix uint32) string {
	return uid + ":" + strconv.FormatInt(int64(seqSuffix), 10)
}

func isNotContainInt32(target uint32, List []uint32) bool {
	for _, i := range List {
		if i == target {
			return false
		}
	}
	return true
}

func (d *DataBases) DelUserChat(uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	//
	//delTime := time.Now().Unix() - int64(config.Config.Mongo.DBRetainChatRecords)*24*3600
	//if err := c.Update(bson.M{"uid": uid}, bson.M{"$pull": bson.M{"msg": bson.M{"sendtime": bson.M{"$lte": delTime}}}}); err != nil {
	//	return err
	//}
	//
	//return nil
}

//
//func (d *DataBases) MgoUserCount() (int, error) {
//	return 0, nil
//	//session := d.mgoSession.Clone()
//	//if session == nil {
//	//	return 0, errors.New("session == nil")
//	//}
//	//defer session.Close()
//	//
//	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
//	//
//	//return c.Find(nil).Count()
//}
//
//func (d *DataBases) MgoSkipUID(count int) (string, error) {
//	return "", nil
//	//session := d.mgoSession.Clone()
//	//if session == nil {
//	//	return "", errors.New("session == nil")
//	//}
//	//defer session.Close()
//	//
//	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
//	//
//	//sChat := UserChat{}
//	//c.Find(nil).Skip(count).Limit(1).One(&sChat)
//	//return sChat.UID, nil
//}

func (d *DataBases) GetGroupMember(groupID string) []string {
	return nil
	//groupInfo := GroupMember_x{}
	//groupInfo.GroupID = groupID
	//groupInfo.UIDList = make([]string, 0)
	//
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return groupInfo.UIDList
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//if err := c.Find(bson.M{"groupid": groupInfo.GroupID}).One(&groupInfo); err != nil {
	//	return groupInfo.UIDList
	//}
	//
	//return groupInfo.UIDList
}

func (d *DataBases) AddGroupMember(groupID, uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//n, err := c.Find(bson.M{"groupid": groupID}).Count()
	//if err != nil {
	//	return err
	//}
	//
	//if n == 0 {
	//	groupInfo := GroupMember_x{}
	//	groupInfo.GroupID = groupID
	//	groupInfo.UIDList = append(groupInfo.UIDList, uid)
	//	err = c.Insert(&groupInfo)
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	err = c.Update(bson.M{"groupid": groupID}, bson.M{"$addToSet": bson.M{"uidlist": uid}})
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//return nil
}

func (d *DataBases) DelGroupMember(groupID, uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//if err := c.Update(bson.M{"groupid": groupID}, bson.M{"$pull": bson.M{"uidlist": uid}}); err != nil {
	//	return err
	//}
	//
	//return nil
}

// deleted by wg 2022-12-12
//func (d *DataBases) SaveUserChatMongo2(uid string, sendTime int64, m *pbMsg.MsgDataToDB) error {
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	newTime := getCurrentTimestampByMill()
//	operationID := ""
//	seqUid := getSeqUid(uid, m.MsgData.Seq)
//	filter := bson.M{"uid": seqUid}
//	var err error
//	sMsg := MsgInfo{}
//	sMsg.SendTime = sendTime
//	if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
//		return utils.Wrap(err, "")
//	}
//	err = c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": sMsg}}).Err()
//	log.NewWarn(operationID, "get mgoSession cost time", getCurrentTimestampByMill()-newTime)
//	if err != nil {
//		sChat := UserChat{}
//		sChat.UID = seqUid
//		sChat.Msg = append(sChat.Msg, sMsg)
//		if _, err = c.InsertOne(ctx, &sChat); err != nil {
//			log.NewDebug(operationID, "InsertOne failed", filter)
//			return utils.Wrap(err, "")
//		}
//	} else {
//		//log.NewDebug(operationID, "FindOneAndUpdate ok", filter)
//	}
//
//	//log.NewDebug(operationID, "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
//	return nil
//}

//
//func (d *DataBases) SaveUserChatListMongo2(uid string, sendTime int64, msgList []*pbMsg.MsgDataToDB) error {
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	newTime := getCurrentTimestampByMill()
//	operationID := ""
//	seqUid := ""
//	msgListToMongo := make([]MsgInfo, 0)
//
//	for _, m := range msgList {
//		seqUid = getSeqUid(uid, m.MsgData.Seq)
//		var err error
//		sMsg := MsgInfo{}
//		sMsg.SendTime = sendTime
//		if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
//			return utils.Wrap(err, "")
//		}
//		msgListToMongo = append(msgListToMongo, sMsg)
//	}
//
//	filter := bson.M{"uid": seqUid}
//	log.NewDebug(operationID, "filter ", seqUid)
//	err := c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": bson.M{"$each": msgListToMongo}}}).Err()
//	log.NewWarn(operationID, "get mgoSession cost time", getCurrentTimestampByMill()-newTime)
//	if err != nil {
//		sChat := UserChat{}
//		sChat.UID = seqUid
//		sChat.Msg = msgListToMongo
//
//		if _, err = c.InsertOne(ctx, &sChat); err != nil {
//			log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
//			return utils.Wrap(err, "")
//		}
//	} else {
//		log.NewDebug(operationID, "FindOneAndUpdate ok", filter)
//	}
//
//	log.NewDebug(operationID, "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
//	return nil
//}

// deleteMsgByLogic delete by wg 2022-12-09
//func (d *DataBases) DelMsgLogic(uid string, seqList []uint32, operationID string) error {
//	sortkeys.Uint32s(seqList)
//	seqMsgs, err := d.GetMsgBySeqListMongo2(uid, seqList, operationID)
//	if err != nil {
//		return utils.Wrap(err, "")
//	}
//	for _, seqMsg := range seqMsgs {
//		log.NewDebug(operationID, utils.GetSelfFuncName(), *seqMsg)
//		seqMsg.Status = constant.MsgDeleted
//		if err = d.ReplaceMsgBySeq(uid, seqMsg, operationID); err != nil {
//			log.NewError(operationID, utils.GetSelfFuncName(), "ReplaceMsgListBySeq error", err.Error())
//		}
//	}
//	return nil
//}

//deleted by wg 2022-12-12
//func (d *DataBases) GetMinSeqFromMongo(uid string) (MinSeq uint32, err error) {
//	return 1, nil
//	//var i, NB uint32
//	//var seqUid string
//	//session := d.mgoSession.Clone()
//	//if session == nil {
//	//	return MinSeq, errors.New("session == nil")
//	//}
//	//defer session.Close()
//	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
//	//MaxSeq, err := d.GetUserMaxSeq(uid)
//	//if err != nil && err != redis.ErrNil {
//	//	return MinSeq, err
//	//}
//	//NB = uint32(MaxSeq / singleGocMsgNum)
//	//for i = 0; i <= NB; i++ {
//	//	seqUid = indexGen(uid, i)
//	//	n, err := c.Find(bson.M{"uid": seqUid}).Count()
//	//	if err == nil && n != 0 {
//	//		if i == 0 {
//	//			MinSeq = 1
//	//		} else {
//	//			MinSeq = uint32(i * singleGocMsgNum)
//	//		}
//	//		break
//	//	}
//	//}
//	//return MinSeq, nil
//}
//
//func (d *DataBases) GetMinSeqFromMongo2(uid string) (MinSeq uint32, err error) {
//	return 1, nil
//}
