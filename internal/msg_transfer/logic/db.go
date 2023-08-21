package logic

import (
	"Open_IM/pkg/common/db"
	pbMsg "Open_IM/pkg/proto/msg"
)

func saveUserChatList(userID string, msgList []*pbMsg.MsgDataToMQ, operationID string) (error, uint64) {
	//log.Info(operationID, utils.GetSelfFuncName(), "args ", userID, len(msgList))
	return db.DB.BatchInsertChat2Cache(userID, msgList, operationID)
}