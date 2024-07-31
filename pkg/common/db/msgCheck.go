package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// 1 查看输出日志 统计内容类型数量
func (d *DataBases) UserMsgLogs(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
	maxSeq, err := d.GetUserMaxSeq(uid)
	if err == redis.Nil {
		return seqMsg, nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	regex := fmt.Sprintf("^%s", uid)
	findOpts := options.Find().SetSort(bson.M{"uid": 1})
	var userChats []UserChat
	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &userChats); err != nil {
		return nil, err
	}
	mapContentCount := make(map[int32]int)
	for _, userChat := range userChats {
		cursor.Decode(&userChat)
		log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
		for i := 0; i < len(userChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", uid, err.Error())
				return nil, err
			}
			log.NewError(operationID, "UserMsgLogs", i, msg.Seq, msg.ContentType, utils.UnixMillSecondToTime(msg.SendTime))
			if v, ok := mapContentCount[msg.ContentType]; ok {
				mapContentCount[msg.ContentType] = v + 1
			} else {
				mapContentCount[msg.ContentType] = 1
			}
		}
	}
	for k, v := range mapContentCount {
		log.Error(operationID, utils.GetSelfFuncName(), "contentCount", k, v)
	}
	return nil, nil
}

// 2 过滤掉系统消息并重置seq值 替换
func (d *DataBases) ResetSystemMsgList(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
	maxSeq, err := d.GetUserMaxSeq(uid)
	if err == redis.Nil {
		return seqMsg, nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	regex := fmt.Sprintf("^%s", uid)
	findOpts := options.Find().SetSort(bson.M{"uid": 1})
	var userChats []UserChat2
	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &userChats); err != nil {
		return nil, err
	}
	allMsgs := make([]open_im_sdk.MsgData, 0)
	sendTimes := make([]int64, 0)
	for _, userChat := range userChats {
		cursor.Decode(&userChat)
		log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
		for i := 0; i < len(userChat.Msg); i++ {
			if userChat.Msg[i].SendTime == 0 {
				continue
			}
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", uid, err.Error())
				return nil, err
			}
			if msg.ContentType < 1000 || msg.ContentType == 1501 {
				allMsgs = append(allMsgs, *msg)
				sendTimes = append(sendTimes, userChat.Msg[i].SendTime)
			}
		}
		log.NewError(operationID, userChat.ID, userChat.UID)
		modifyUidId := fmt.Sprintf("modify_%s", userChat.UID)
		log.NewError(operationID, "userChat", userChat.ID, userChat.UID, modifyUidId)
		objID, _ := primitive.ObjectIDFromHex(userChat.ID)
		_, err = c.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"uid": modifyUidId}})
	}
	var seq uint32 = 1
	mapUserChat := make(map[string]UserChat)
	//uuid := getSeqUid(uid, uint32(i))
	for i := 0; i < len(allMsgs); i++ {
		allMsgs[i].Seq = seq
		msg, err := proto.Marshal(&allMsgs[i])
		if err != nil {
			log.NewError(operationID, "userChat", err.Error())
			return nil, err
		}
		msgInfo := MsgInfo{
			Msg:      msg,
			SendTime: sendTimes[i],
		}
		uuid := getSeqUid(uid, seq)
		if us, ok := mapUserChat[uuid]; ok {
			us.Msg = append(us.Msg, msgInfo)
			mapUserChat[uuid] = us
		} else {
			nuc := UserChat{
				UID: uuid,
				Msg: []MsgInfo{},
			}
			nuc.Msg = append(nuc.Msg, msgInfo)
			mapUserChat[uuid] = nuc
		}
		seq++
	}
	for k, v := range mapUserChat {
		if _, err = c.InsertOne(ctx, &v); err != nil {
			log.NewError(operationID, "InsertOne failed", k, len(v.Msg), err.Error())
			return nil, nil
		}
	}
	if len(sendTimes) == 0 {
		seq = 0
	}
	err = d.SetUserMaxSeq(uid, uint64(seq-1))
	if err != nil {
		log.NewError(operationID, "SetUserMaxSeq", uid, seq, err.Error())
		return nil, err
	}

	return nil, nil
}

// 3 对比seq与消息中seq对应关系
func (d *DataBases) CheckUserSeq(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	regex := fmt.Sprintf("^%s", uid)
	findOpts := options.Find().SetLimit(1).SetSort(bson.M{"uid": -1})
	var userChats []UserChat
	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &userChats); err != nil {
		return nil, err
	}
	var msgMaxSeq uint32 = 0
	for _, userChat := range userChats {
		cursor.Decode(&userChat)
		log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
		for i := 0; i < len(userChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", uid, err.Error())
				return nil, err
			}
			if msgMaxSeq < msg.Seq {
				msgMaxSeq = msg.Seq
			}
		}
	}
	maxSeq, err := d.GetUserMaxSeq(uid)
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq, msgMaxSeq)
	if err == redis.Nil {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "redis nil", maxSeq, msgMaxSeq)
	} else if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "redis err", maxSeq, msgMaxSeq, err.Error())
	} else {
		if maxSeq != uint64(msgMaxSeq) {
			log.NewError(operationID, utils.GetSelfFuncName(), "redis seq no match", uid, maxSeq, msgMaxSeq)
		}
	}
	return nil, nil
}

// 1
//func (d *DataBases) CheckGroupAllMsgList(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
//	maxSeq, err := d.GetUserMaxSeq(uid)
//	if err == redis.Nil {
//		return seqMsg, nil
//	}
//	if err != nil {
//		return nil, utils.Wrap(err, "")
//	}
//	seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq, seqUsers)
//
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	regex := fmt.Sprintf("^%s", uid)
//	findOpts := options.Find().SetSort(bson.M{"uid": 1})
//	var userChats []UserChat2
//	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
//	if err != nil {
//		return nil, err
//	}
//	if err = cursor.All(context.TODO(), &userChats); err != nil {
//		return nil, err
//	}
//	mapSeq := make(map[uint32]int)
//	mapSeqMsg := make(map[uint32]open_im_sdk.MsgData)
//	for _, userChat := range userChats {
//		cursor.Decode(&userChat)
//		//log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
//		var currentMsgs []MsgInfo
//		for i := 0; i < len(userChat.Msg); i++ {
//			msg := new(open_im_sdk.MsgData)
//			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
//				log.NewError(operationID, "Unmarshal err", uid, err.Error())
//				return nil, err
//			}
//			if val, ok := mapSeq[msg.Seq]; ok {
//				//log.NewError(operationID, "seq dup", msg.Seq, msg.ClientMsgID, msg.ServerMsgID, msg.ContentType, msg.SendTime, msg.CreateTime, val)
//				mapSeq[msg.Seq] = val + 1
//				if msg.ContentType < 1000 {
//					currentMsgs = append(currentMsgs, userChat.Msg[i])
//					if data, ok1 := mapSeqMsg[msg.Seq]; ok1 {
//						log.NewError(operationID, "seq dup msg", data.Seq, data.ClientMsgID, data.ServerMsgID, data.ContentType, data.SendTime, data.CreateTime)
//					}
//				}
//			} else {
//				mapSeq[msg.Seq] = 1
//				currentMsgs = append(currentMsgs, userChat.Msg[i])
//				mapSeqMsg[msg.Seq] = *msg
//			}
//		}
//		//log.NewError(operationID, userChat.ID, userChat.UID, len(userChat.Msg), len(currentMsgs))
//		if len(currentMsgs) != len(userChat.Msg) {
//			objID, _ := primitive.ObjectIDFromHex(userChat.ID)
//			_, err = c.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"msg": currentMsgs}})
//			if err != nil {
//				log.NewError(operationID, "update err", userChat.ID, userChat.UID, err.Error())
//			} else {
//				log.NewError(operationID, "update success", userChat.ID, userChat.UID)
//			}
//		}
//	}
//	return nil, nil
//}

// 2 处理重复 seq
//func (d *DataBases) ResizeGroupAllMsgList(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
//	maxSeq, err := d.GetUserMaxSeq(uid)
//	if err == redis.Nil {
//		return seqMsg, nil
//	}
//	if err != nil {
//		return nil, utils.Wrap(err, "")
//	}
//	seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq, seqUsers)
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	regex := fmt.Sprintf("^%s", uid)
//	findOpts := options.Find().SetSort(bson.M{"uid": 1})
//	var userChats []UserChat
//	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
//	if err != nil {
//		return nil, err
//	}
//	if err = cursor.All(context.TODO(), &userChats); err != nil {
//		return nil, err
//	}
//	mapSeq := make(map[uint32]int)
//	mapSeqData := make(map[uint32]MsgInfo)
//	mapSeqMsg := make(map[uint32]open_im_sdk.MsgData)
//	for _, userChat := range userChats {
//		cursor.Decode(&userChat)
//		log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
//		for i := 0; i < len(userChat.Msg); i++ {
//			msg := new(open_im_sdk.MsgData)
//			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
//				log.NewError(operationID, "Unmarshal err", uid, err.Error())
//				return nil, err
//			}
//			if val, ok := mapSeq[msg.Seq]; ok {
//				log.NewError(operationID, "seq dup", msg.Seq, msg.ClientMsgID, msg.ServerMsgID, msg.ContentType, msg.SendTime, msg.CreateTime, val)
//				mapSeq[msg.Seq] = val + 1
//				if msg.ContentType < 1000 {
//					if data, ok1 := mapSeqMsg[msg.Seq]; ok1 {
//						log.NewError(operationID, "seq dup msg", data.Seq, data.ClientMsgID, data.ServerMsgID, data.ContentType, data.SendTime, data.CreateTime)
//						if data.ContentType > 1000 || (data.ContentType > msg.ContentType && msg.ContentType == 110) {
//							mapSeqData[msg.Seq] = userChat.Msg[i]
//							mapSeqMsg[msg.Seq] = *msg
//						}
//					}
//				}
//			} else {
//				mapSeq[msg.Seq] = 1
//				mapSeqData[msg.Seq] = userChat.Msg[i]
//				mapSeqMsg[msg.Seq] = *msg
//			}
//		}
//	}
//	mapUserChat := make(map[string]UserChat)
//	for i := 1; i <= int(maxSeq); i++ {
//		uuid := getSeqUid(uid, uint32(i))
//		if us, ok2 := mapUserChat[uuid]; ok2 {
//			if data, ok := mapSeqData[uint32(i)]; ok {
//				us.Msg = append(us.Msg, data)
//			} else {
//				emptyData := MsgInfo{
//					SendTime: 0,
//					Msg:      nil,
//				}
//				us.Msg = append(us.Msg, emptyData)
//			}
//			mapUserChat[uuid] = us
//		} else {
//			data3 := UserChat{
//				UID: uuid,
//				Msg: []MsgInfo{},
//			}
//			if data, ok := mapSeqData[uint32(i)]; ok {
//				data3.Msg = append(data3.Msg, data)
//			} else {
//				emptyData := MsgInfo{
//					SendTime: 0,
//					Msg:      nil,
//				}
//				data3.Msg = append(data3.Msg, emptyData)
//			}
//			mapUserChat[uuid] = data3
//		}
//	}
//	for k, v := range mapUserChat {
//		modifyUidId := fmt.Sprintf("modify_%s", v.UID)
//		log.NewError(operationID, "mapUserChat", k, len(v.Msg), v.UID, modifyUidId)
//		//_, err = c.UpdateOne(sCtx, bson.M{"user_id": userID}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
//		_, err = c.UpdateOne(ctx, bson.M{"uid": v.UID}, bson.M{"$set": bson.M{"uid": modifyUidId}})
//		if err == nil {
//			if _, err = c.InsertOne(ctx, &v); err != nil {
//				log.NewError(operationID, "InsertOne failed", k, len(v.Msg), err.Error())
//				return nil, nil
//			}
//		}
//	}
//	return nil, nil
//}

// 4 本地更新消息
//func (d *DataBases) RemoveGroupSystemMsgList(uid string, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid)
//	maxSeq, err := d.GetUserMaxSeq(uid)
//	if err == redis.Nil {
//		return seqMsg, nil
//	}
//	if err != nil {
//		return nil, utils.Wrap(err, "")
//	}
//	//seqUsers := getSeqUserIDList(uid, uint32(maxSeq))
//	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, maxSeq)
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	regex := fmt.Sprintf("^%s", uid)
//	findOpts := options.Find().SetSort(bson.M{"uid": 1})
//	var userChats []UserChat2
//	cursor, err := c.Find(ctx, bson.M{"uid": bson.M{"$regex": regex}}, findOpts)
//	if err != nil {
//		return nil, err
//	}
//	if err = cursor.All(context.TODO(), &userChats); err != nil {
//		return nil, err
//	}
//	for _, userChat := range userChats {
//		cursor.Decode(&userChat)
//		//log.NewInfo(operationID, utils.GetSelfFuncName(), "range", userChat.UID, len(userChat.Msg))
//		currentMsgs := make([]MsgInfo, 0)
//		for i := 0; i < len(userChat.Msg); i++ {
//			if userChat.Msg[i].SendTime == 0 {
//				continue
//			}
//			msg := new(open_im_sdk.MsgData)
//			if err = proto.Unmarshal(userChat.Msg[i].Msg, msg); err != nil {
//				log.NewError(operationID, "Unmarshal err", uid, err.Error())
//				return nil, err
//			}
//			if msg.ContentType < 1000 || msg.ContentType == 1501 {
//				currentMsgs = append(currentMsgs, userChat.Msg[i])
//			}
//		}
//		//log.NewError(operationID, userChat.ID, userChat.UID, len(userChat.Msg), len(currentMsgs))
//		if len(currentMsgs) != len(userChat.Msg) {
//			//log.NewError(operationID, "update user chat")
//			objID, _ := primitive.ObjectIDFromHex(userChat.ID)
//			_, err = c.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"msg": currentMsgs}})
//			if err != nil {
//				log.NewError(operationID, "update err", userChat.ID, userChat.UID, err.Error())
//			} else {
//				log.NewError(operationID, "update success", userChat.ID, userChat.UID)
//			}
//		}
//	}
//	return nil, nil
//}
