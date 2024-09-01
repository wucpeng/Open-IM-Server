package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	promePkg "Open_IM/pkg/common/prometheus"
	pbMsg "Open_IM/pkg/proto/msg"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	go_redis "github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (d *DataBases) BatchInsertChat2Mongo(userID string, msgList []*pbMsg.MsgDataToMQ, operationID string, currentMaxSeq uint64) error {
	if len(msgList) > GetSingleGocMsgNum() {
		return errors.New("too large") //5000
	}
	uuids := make([]string, 0)
	mapUidMsgs := make(map[string][]MsgInfo)
	var err error
	for _, m := range msgList {
		currentMaxSeq++
		sMsg := MsgInfo{}
		sMsg.SendTime = m.MsgData.SendTime
		m.MsgData.Seq = uint32(currentMaxSeq)
		if m.MsgData.Seq == 0 {
			m.MsgData.Seq = uint32(currentMaxSeq)
		}
		log.Info(operationID, "msg node ", m.MsgData.Seq, currentMaxSeq, m.MsgData.ClientMsgID)
		if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
			return utils.Wrap(err, "")
		}
		uuid := getSeqUid(userID, m.MsgData.Seq)
		if v, ok := mapUidMsgs[uuid]; ok {
			mapUidMsgs[uuid] = append(v, sMsg)
		} else {
			uuids = append(uuids, uuid)
			ms := make([]MsgInfo, 0)
			ms = append(ms, sMsg)
			mapUidMsgs[uuid] = ms
		}
	}
	log.Info(operationID, utils.GetSelfFuncName(), len(msgList), currentMaxSeq, uuids)
	ctx := context.Background()
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
	for _, uuid := range uuids {
		if msgs, ok := mapUidMsgs[uuid]; ok {
			log.Info(operationID, utils.GetSelfFuncName(), "try update or insert", uuid, len(msgs))
			filter := bson.M{"uid": uuid}
			err := c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": bson.M{"$each": msgs}}}).Err()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					filter := bson.M{"uid": uuid}
					sChat := UserChat{}
					sChat.UID = uuid
					sChat.Msg = msgs
					if _, err = c.InsertOne(ctx, &sChat); err != nil {
						promePkg.PromeInc(promePkg.MsgInsertMongoFailedCounter)
						log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
						return utils.Wrap(err, "")
					} else {
						log.Info(operationID, utils.GetSelfFuncName(), "insert new", uuid, len(msgs))
					}
					promePkg.PromeInc(promePkg.MsgInsertMongoSuccessCounter)
				} else {
					promePkg.PromeInc(promePkg.MsgInsertMongoFailedCounter)
					log.Error(operationID, "FindOneAndUpdate failed ", err.Error(), filter)
					return utils.Wrap(err, "")
				}
			} else {
				promePkg.PromeInc(promePkg.MsgInsertMongoSuccessCounter)
			}
		}
	}
	return nil
}

// redis
func (d *DataBases) BatchInsertChat2Cache(insertID string, msgList []*pbMsg.MsgDataToMQ, operationID string) (error, uint64) {
	//newTime := getCurrentTimestampByMill()
	lenList := len(msgList)
	if lenList > GetSingleGocMsgNum() { //5000
		return errors.New("too large"), 0
	}
	if lenList < 1 {
		return errors.New("too short as 0"), 0
	}
	// judge sessionType to get seq
	var currentMaxSeq uint64
	var err error
	if msgList[0].MsgData.SessionType == constant.SuperGroupChatType {
		currentMaxSeq, err = d.GetGroupMaxSeq(insertID)
		log.Debug(operationID, "constant.SuperGroupChatType  lastMaxSeq before add ", currentMaxSeq, "userID ", insertID, err)
	} else {
		currentMaxSeq, err = d.GetUserMaxSeq(insertID)
		log.Debug(operationID, "constant.SingleChatType  lastMaxSeq before add ", currentMaxSeq, "userID ", insertID, len(msgList), err)
	}
	if err != nil && err != go_redis.Nil {
		promePkg.PromeInc(promePkg.SeqGetFailedCounter)
		return utils.Wrap(err, ""), 0
	}
	promePkg.PromeInc(promePkg.SeqGetSuccessCounter)

	lastMaxSeq := currentMaxSeq
	for _, m := range msgList {
		currentMaxSeq++
		m.MsgData.Seq = uint32(currentMaxSeq)
	}
	//log.Debug(operationID, "SetMessageToCache ", insertID, len(msgList))
	err, failedNum := d.SetMessageToCache(msgList, insertID, operationID)
	if err != nil {
		promePkg.PromeAdd(promePkg.MsgInsertRedisFailedCounter, failedNum)
		log.Error(operationID, "setMessageToCache failed, continue ", err.Error(), len(msgList), insertID)
	} else {
		promePkg.PromeInc(promePkg.MsgInsertRedisSuccessCounter)
	}
	//log.Debug(operationID, "batch to redis  cost time ", getCurrentTimestampByMill()-newTime, insertID, len(msgList))
	if msgList[0].MsgData.SessionType == constant.SuperGroupChatType {
		err = d.SetGroupMaxSeq(insertID, currentMaxSeq)
	} else {
		err = d.SetUserMaxSeq(insertID, currentMaxSeq)
	}
	if err != nil {
		promePkg.PromeInc(promePkg.SeqSetFailedCounter)
		log.NewError(operationID, utils.GetSelfFuncName(), "redis err", insertID, currentMaxSeq, err.Error())
	} else {
		promePkg.PromeInc(promePkg.SeqSetSuccessCounter)
	}
	return utils.Wrap(err, ""), lastMaxSeq
}
