package msg

import (
	"Open_IM/pkg/common/constant"
	rocksCache "Open_IM/pkg/common/db/rocks_cache"
	"Open_IM/pkg/utils"
	"context"
	go_redis "github.com/go-redis/redis/v8"
	"time"

	commonDB "Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"

	promePkg "Open_IM/pkg/common/prometheus"
)

func (rpc *rpcChat) GetMaxAndMinSeq(_ context.Context, in *open_im_sdk.GetMaxAndMinSeqReq) (*open_im_sdk.GetMaxAndMinSeqResp, error) {
	//log.NewInfo(in.OperationID, "rpc getMaxAndMinSeq is arriving", in.String())
	resp := new(open_im_sdk.GetMaxAndMinSeqResp)
	m := make(map[string]*open_im_sdk.MaxAndMinSeq)
	var maxSeq, minSeq uint64
	var err1, err2 error
	maxSeq, err1 = commonDB.DB.GetUserMaxSeq(in.UserID)
	minSeq, err2 = commonDB.DB.GetUserMinSeq(in.UserID)
	if (err1 != nil && err1 != go_redis.Nil) || (err2 != nil && err2 != go_redis.Nil) {
		log.NewError(in.OperationID, "getMaxSeq from redis error", in.String())
		if err1 != nil {
			log.NewError(in.OperationID, utils.GetSelfFuncName(), err1.Error())
		}
		if err2 != nil {
			log.NewError(in.OperationID, utils.GetSelfFuncName(), err2.Error())
		}
		resp.ErrCode = 200
		resp.ErrMsg = "redis get err"
		return resp, nil
	}
	resp.MaxSeq = uint32(maxSeq)
	resp.MinSeq = uint32(minSeq)
	for _, groupID := range in.GroupIDList {
		x := new(open_im_sdk.MaxAndMinSeq)
		maxSeq, _ := commonDB.DB.GetGroupMaxSeq(groupID)
		minSeq, _ := commonDB.DB.GetGroupUserMinSeq(groupID, in.UserID)
		x.MaxSeq = uint32(maxSeq)
		x.MinSeq = uint32(minSeq)
		m[groupID] = x
	}
	resp.GroupMaxAndMinSeq = m
	return resp, nil
}

func (rpc *rpcChat) PullMessageBySeqList(_ context.Context, in *open_im_sdk.PullMessageBySeqListReq) (*open_im_sdk.PullMessageBySeqListResp, error) {
	log.NewInfo(in.OperationID, "rpc PullMessageBySeqList is arriving", in.String())
	resp := new(open_im_sdk.PullMessageBySeqListResp)
	//m := make(map[string]*open_im_sdk.MsgDataList)
	redisMsgList, failedSeqList, err := commonDB.DB.GetMessageListBySeq(in.UserID, in.SeqList, in.OperationID)
	if err != nil {
		if err != go_redis.Nil {
			promePkg.PromeAdd(promePkg.MsgPullFromRedisFailedCounter, len(failedSeqList))
			log.Error(in.OperationID, "get message from redis exception", err.Error(), failedSeqList)
		} else {
			log.Info(in.OperationID, "get message from redis is nil", failedSeqList, err.Error())
		}
		msgList, err1 := commonDB.DB.GetMsgBySeqListMongo2(in.UserID, failedSeqList, in.OperationID)
		if err1 != nil {
			promePkg.PromeAdd(promePkg.MsgPullFromMongoFailedCounter, len(failedSeqList))
			log.Error(in.OperationID, "PullMessageBySeqList data error", in.String(), err.Error())
			resp.ErrCode = 201
			resp.ErrMsg = err.Error()
			return resp, nil
		} else {
			promePkg.PromeAdd(promePkg.MsgPullFromMongoSuccessCounter, len(msgList))
			log.Info(in.OperationID, "get message from mongo",  len(msgList))
			redisMsgList = append(redisMsgList, msgList...)
			resp.List = redisMsgList
		}
	} else {
		promePkg.PromeAdd(promePkg.MsgPullFromRedisSuccessCounter, len(redisMsgList))
		resp.List = redisMsgList
	}
	// TODO 处理客服账号小红点
	if in.UserID == "111111111111111111111112" && len(resp.List) > 0 {
		t1 := time.Now()
		//log.Info(in.OperationID, utils.GetSelfFuncName(), "custom process", len(resp.List))
		for _, v := range resp.List {
			isUnread := utils.GetSwitchFromOptions(v.Options, constant.IsUnreadCount)
			if isUnread {
				var conId string
				if v.SessionType == constant.SingleChatType {
					if in.UserID == v.RecvID {
						conId = utils.GetConversationIDBySessionType(v.SendID, constant.SingleChatType)
					} else {
						conId = utils.GetConversationIDBySessionType(v.RecvID, constant.SingleChatType)
					}
				} else if v.SessionType == constant.GroupChatType {
					conId = utils.GetConversationIDBySessionType(v.GroupID, constant.GroupChatType)
				}
				if len(conId) > 0 {
					conversation, err := rocksCache.GetConversationFromCache(in.UserID, conId)
					if err == nil {
						if conversation.UpdateUnreadCountTime > v.SendTime {
							utils.SetSwitchFromOptions(v.Options, constant.IsUnreadCount, false)
						}
					} else {
						log.Error(in.OperationID, "custom process error", conId, err.Error())
					}
				}
			}
		}
		log.NewError(in.OperationID, "custom cost", time.Since(t1))
	}
	return resp, nil
}

type MsgFormats []*open_im_sdk.MsgData

// Implement the sort.Interface interface to get the number of elements method
func (s MsgFormats) Len() int {
	return len(s)
}

//Implement the sort.Interface interface comparison element method
func (s MsgFormats) Less(i, j int) bool {
	return s[i].SendTime < s[j].SendTime
}

//Implement the sort.Interface interface exchange element method
func (s MsgFormats) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
