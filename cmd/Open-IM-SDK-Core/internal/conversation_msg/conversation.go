package conversation_msg

import (
	"errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	_ "open_im_sdk/internal/common"
	"open_im_sdk/open_im_sdk_callback"
	"open_im_sdk/pkg/common"
	"open_im_sdk/pkg/constant"
	"open_im_sdk/pkg/db/model_struct"
	"open_im_sdk/pkg/log"
	sdk "open_im_sdk/pkg/sdk_params_callback"
	"open_im_sdk/pkg/server_api_params"
	"open_im_sdk/pkg/utils"
	"open_im_sdk/sdk_struct"
	"sort"
	"time"
)

func (c *Conversation) getAllConversationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetAllConversationListCallback {
	conversationList, err := c.db.GetAllConversationList()
	common.CheckDBErrCallback(callback, err, operationID)
	return conversationList
}

func (c *Conversation) getConversationListSplit(callback open_im_sdk_callback.Base, offset, count int, operationID string) sdk.GetConversationListSplitCallback {
	conversationList, err := c.db.GetConversationListSplit(offset, count)
	common.CheckDBErrCallback(callback, err, operationID)
	return conversationList
}

func (c *Conversation) setConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationIDList []string, opt int, operationID string) {
	apiReq := server_api_params.BatchSetConversationsReq{}
	apiResp := server_api_params.BatchSetConversationsResp{}
	apiReq.OperationID = operationID
	apiReq.OwnerUserID = c.loginUserID
	apiReq.NotificationType = constant.ConversationChangeNotification
	var conversations []server_api_params.Conversation
	for _, conversationID := range conversationIDList {
		localConversation, err := c.db.GetConversation(conversationID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), "GetConversation failed", err.Error())
			continue
		}
		conversations = append(conversations, server_api_params.Conversation{
			OwnerUserID:      c.loginUserID,
			ConversationID:   conversationID,
			ConversationType: localConversation.ConversationType,
			UserID:           localConversation.UserID,
			GroupID:          localConversation.GroupID,
			RecvMsgOpt:       int32(opt),
			IsPinned:         localConversation.IsPinned,
			IsPrivateChat:    localConversation.IsPrivateChat,
			AttachedInfo:     localConversation.AttachedInfo,
			Ex:               localConversation.Ex,
		})
	}
	apiReq.Conversations = conversations
	c.p.PostFatalCallback(callback, constant.BatchSetConversationRouter, apiReq, &apiResp, apiReq.OperationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "output: ", apiResp)
	c.SyncConversations(operationID)
}

func (c *Conversation) setConversation(callback open_im_sdk_callback.Base, apiReq *server_api_params.ModifyConversationFieldReq, conversationID string, localConversation *model_struct.LocalConversation, operationID string) {
	apiResp := server_api_params.ModifyConversationFieldResp{}
	apiReq.OwnerUserID = c.loginUserID
	apiReq.OperationID = operationID
	apiReq.ConversationID = conversationID
	apiReq.ConversationType = localConversation.ConversationType
	apiReq.UserID = localConversation.UserID
	apiReq.GroupID = localConversation.GroupID
	apiReq.UserIDList = []string{c.loginUserID}
	c.p.PostFatalCallback(callback, constant.ModifyConversationFieldRouter, apiReq, nil, apiReq.OperationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "request success, output: ", apiResp)
}

func (c *Conversation) setGlobalRecvMessageOpt(callback open_im_sdk_callback.Base, opt int32, operationID string) {
	apiReq := server_api_params.SetGlobalRecvMessageOptReq{}
	apiReq.OperationID = operationID
	apiReq.GlobalRecvMsgOpt = &opt
	c.p.PostFatalCallback(callback, constant.SetGlobalRecvMessageOptRouter, apiReq, nil, apiReq.OperationID)
	c.user.SyncLoginUserInfo(operationID)
}
func (c *Conversation) setOneConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationID string, opt int, operationID string) {
	apiReq := &server_api_params.ModifyConversationFieldReq{}
	localConversation, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	apiReq.RecvMsgOpt = int32(opt)
	apiReq.FieldType = constant.FieldRecvMsgOpt
	c.setConversation(callback, apiReq, conversationID, localConversation, operationID)
	c.SyncConversations(operationID)
}

func (c *Conversation) setOneConversationPrivateChat(callback open_im_sdk_callback.Base, conversationID string, isPrivate bool, operationID string) {
	apiReq := &server_api_params.ModifyConversationFieldReq{}
	localConversation, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	apiReq.IsPrivateChat = isPrivate
	apiReq.FieldType = constant.FieldIsPrivateChat
	c.setConversation(callback, apiReq, conversationID, localConversation, operationID)
	c.SyncConversations(operationID)
}

func (c *Conversation) setOneConversationPinned(callback open_im_sdk_callback.Base, conversationID string, isPinned bool, operationID string) {
	apiReq := &server_api_params.ModifyConversationFieldReq{}
	localConversation, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	apiReq.IsPinned = isPinned
	apiReq.FieldType = constant.FieldIsPinned
	c.setConversation(callback, apiReq, conversationID, localConversation, operationID)
	c.SyncConversations(operationID)
}
func (c *Conversation) setOneConversationGroupAtType(callback open_im_sdk_callback.Base, conversationID, operationID string) {
	lc, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	if lc.GroupAtType == constant.AtNormal || lc.ConversationType != constant.GroupChatType {
		common.CheckAnyErrCallback(callback, 201, errors.New("conversation don't need to reset"), operationID)
	}
	apiReq := &server_api_params.ModifyConversationFieldReq{}
	localConversation, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	apiReq.GroupAtType = constant.AtNormal
	apiReq.FieldType = constant.FieldGroupAtType
	c.setConversation(callback, apiReq, conversationID, localConversation, operationID)
	c.SyncConversations(operationID)
}
func (c *Conversation) getConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationIDList []string, operationID string) []server_api_params.GetConversationRecvMessageOptResp {
	apiReq := server_api_params.GetConversationsReq{}
	apiReq.OperationID = operationID
	apiReq.OwnerUserID = c.loginUserID
	apiReq.ConversationIDs = conversationIDList
	var resp []server_api_params.GetConversationRecvMessageOptResp
	conversations := c.getMultipleConversation(callback, conversationIDList, operationID)
	for _, conversation := range conversations {
		resp = append(resp, server_api_params.GetConversationRecvMessageOptResp{
			ConversationID: conversation.ConversationID,
			Result:         &conversation.RecvMsgOpt,
		})
	}
	return resp
}

func (c *Conversation) getOneConversation(callback open_im_sdk_callback.Base, sourceID string, sessionType int32, operationID string) *model_struct.LocalConversation {
	conversationID := utils.GetConversationIDBySessionType(sourceID, int(sessionType))
	lc, err := c.db.GetConversation(conversationID)
	if err == nil {
		return lc
	} else {
		var newConversation model_struct.LocalConversation
		newConversation.ConversationID = conversationID
		newConversation.ConversationType = sessionType
		switch sessionType {
		case constant.SingleChatType:
			newConversation.UserID = sourceID
			faceUrl, name, err, isFromSvr := c.friend.GetUserNameAndFaceUrlByUid(sourceID, operationID)
			//	faceUrl, name, err := c.cache.GetUserNameAndFaceURL(sourceID, operationID)
			common.CheckDBErrCallback(callback, err, operationID)
			if isFromSvr {
				c.cache.Update(sourceID, faceUrl, name)
			}
			newConversation.ShowName = name
			newConversation.FaceURL = faceUrl
		case constant.GroupChatType, constant.SuperGroupChatType:
			newConversation.GroupID = sourceID
			g, err := c.full.GetGroupInfoFromLocal2Svr(sourceID, sessionType)
			//g, err := c.db.GetGroupInfoByGroupID(sourceID)
			common.CheckDBErrCallback(callback, err, operationID)
			newConversation.ShowName = g.GroupName
			newConversation.FaceURL = g.FaceURL
		}
		err := c.db.InsertConversation(&newConversation)
		common.CheckDBErrCallback(callback, err, operationID)
		return &newConversation
	}
}
func (c *Conversation) getMultipleConversation(callback open_im_sdk_callback.Base, conversationIDList []string, operationID string) sdk.GetMultipleConversationCallback {
	conversationList, err := c.db.GetMultipleConversation(conversationIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	return conversationList
}

func (c *Conversation) deleteConversation(callback open_im_sdk_callback.Base, conversationID, operationID string) {
	lc, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	var sourceID string
	switch lc.ConversationType {
	case constant.SingleChatType, constant.NotificationChatType:
		sourceID = lc.UserID
	case constant.GroupChatType, constant.SuperGroupChatType:
		sourceID = lc.GroupID
	}
	//Mark messages related to this conversation for deletion
	err = c.db.UpdateMessageStatusBySourceIDController(sourceID, constant.MsgStatusHasDeleted, lc.ConversationType)
	common.CheckDBErrCallback(callback, err, operationID)
	//Reset the session information, empty session
	err = c.db.ResetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	c.doUpdateConversation(common.Cmd2Value{Value: common.UpdateConNode{"", constant.TotalUnreadMessageChanged, ""}})

}
func (c *Conversation) setConversationDraft(callback open_im_sdk_callback.Base, conversationID, draftText, operationID string) {
	if draftText != "" {
		err := c.db.SetConversationDraft(conversationID, draftText)
		common.CheckDBErrCallback(callback, err, operationID)
	} else {
		err := c.db.RemoveConversationDraft(conversationID, draftText)
		common.CheckDBErrCallback(callback, err, operationID)
	}
}
func (c *Conversation) pinConversation(callback open_im_sdk_callback.Base, conversationID string, isPinned bool, operationID string) {
	//lc := db.LocalConversation{ConversationID: conversationID, IsPinned: isPinned}
	//if isPinned {
	c.setOneConversationPinned(callback, conversationID, isPinned, operationID)
	//err := c.db.UpdateConversation(&lc)
	//common.CheckDBErrCallback(callback, err, operationID)
	//} else {
	//	err := c.db.UnPinConversation(conversationID, constant.NotPinned)
	//	common.CheckDBErrCallback(callback, err, operationID)
	//}
}
func (c *Conversation) getServerConversationList(operationID string) (server_api_params.GetAllConversationsResp, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName())
	var req server_api_params.GetAllConversationsReq
	var resp server_api_params.GetAllConversationsResp
	req.OwnerUserID = c.loginUserID
	req.OperationID = operationID
	err := c.p.PostReturn(constant.GetAllConversationsRouter, req, &resp.Conversations)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		return resp, err
	}
	return resp, nil
}
func (c *Conversation) SyncConversations(operationID string) {
	var newConversationList []*model_struct.LocalConversation
	ccTime := time.Now()
	log.NewInfo(operationID, utils.GetSelfFuncName())
	conversationsOnServer, err := c.getServerConversationList(operationID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		return
	}
	conversationsOnLocal, err := c.db.GetAllConversationListToSync()
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
	}
	log.Info(operationID, "get server cost time", time.Since(ccTime))
	conversationsOnLocalTempFormat := common.LocalTransferToTempConversation(conversationsOnLocal)
	conversationsOnServerTempFormat := common.ServerTransferToTempConversation(conversationsOnServer)
	conversationsOnServerLocalFormat := common.TransferToLocalConversation(conversationsOnServer)

	aInBNot, bInANot, sameA, sameB := common.CheckConversationListDiff(conversationsOnServerTempFormat, conversationsOnLocalTempFormat)
	log.Info(operationID, "diff server cost time", time.Since(ccTime))

	log.NewInfo(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	// server有 local没有
	// 可能是其他点开一下生成会话设置免打扰 插入到本地 不回调..
	for _, index := range aInBNot {
		conversation := conversationsOnServerLocalFormat[index]
		var newConversation model_struct.LocalConversation
		newConversation.ConversationID = conversation.ConversationID
		newConversation.ConversationType = conversation.ConversationType
		switch conversation.ConversationType {
		case constant.SingleChatType, constant.NotificationChatType:
			newConversation.UserID = conversation.UserID
			faceUrl, name, err, isFromSvr := c.friend.GetUserNameAndFaceUrlByUid(conversation.UserID, operationID)
			if err != nil {
				log.NewError(operationID, utils.GetSelfFuncName(), "GetUserNameAndFaceUrlByUid error", err.Error())
				continue
			}
			if isFromSvr {
				c.cache.Update(conversation.UserID, faceUrl, name)
			}

			newConversation.ShowName = name
			newConversation.FaceURL = faceUrl
		case constant.GroupChatType:
			newConversation.GroupID = conversation.GroupID
			g, err := c.group.GetGroupInfoFromLocal2Svr(conversation.GroupID)
			if err != nil {
				log.NewError(operationID, utils.GetSelfFuncName(), "GetGroupInfoFromLocal2Svr error", err.Error())
				continue
			}
			newConversation.ShowName = g.GroupName
			newConversation.FaceURL = g.FaceURL
		}
		newConversation.RecvMsgOpt = conversation.RecvMsgOpt
		newConversation.IsPinned = conversation.IsPinned
		newConversation.IsPrivateChat = conversation.IsPrivateChat
		newConversation.GroupAtType = conversation.GroupAtType
		newConversation.IsNotInGroup = conversation.IsNotInGroup
		newConversation.Ex = conversation.Ex
		newConversation.AttachedInfo = conversation.AttachedInfo
		newConversationList = append(newConversationList, &newConversation)
		//err := c.db.InsertConversation(&newConversation)
		//if err != nil {
		//	log.NewError(operationID, utils.GetSelfFuncName(), "InsertConversation error", err.Error(), conversation)
		//	continue
		//}
	}
	//New conversation storage
	err2 := c.db.BatchInsertConversationList(newConversationList)
	if err2 != nil {
		log.Error(operationID, "insert new conversation err:", err2.Error(), newConversationList)
	}
	log.Info(operationID, "insert cost time", time.Since(ccTime))
	// 本地服务器有的会话 以服务器为准更新
	var conversationChangedList []string
	for _, index := range sameA {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "server and client both have", *conversationsOnServerLocalFormat[index])
		err := c.db.UpdateConversationForSync(conversationsOnServerLocalFormat[index])
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), "UpdateConversation failed ", err.Error(), *conversationsOnServerLocalFormat[index])
			continue
		}
		conversationChangedList = append(conversationChangedList, conversationsOnServerLocalFormat[index].ConversationID)
	}
	// callback
	if len(conversationChangedList) > 0 {
		if err = common.TriggerCmdUpdateConversation(common.UpdateConNode{Action: constant.ConChange, Args: conversationChangedList}, c.GetCh()); err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		}
	}
	// local有 server没有 代表没有修改公共字段
	for _, index := range bInANot {
		log.NewDebug(operationID, utils.GetSelfFuncName(), index, conversationsOnLocal[index].ConversationID,
			conversationsOnLocal[index].RecvMsgOpt, conversationsOnLocal[index].IsPinned, conversationsOnLocal[index].IsPrivateChat)
	}
}

func (c *Conversation) SyncOneConversation(conversationID, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "conversationID: ", conversationID)
	// todo
}
func (c *Conversation) getHistoryMessageList(callback open_im_sdk_callback.Base, req sdk.GetHistoryMessageListParams, operationID string, isReverse bool) sdk.GetHistoryMessageListCallback {
	t := time.Now()
	var sourceID string
	var conversationID string
	var startTime int64
	var sessionType int
	var list []*model_struct.LocalChatLog
	var err error
	var messageList sdk_struct.NewMsgList
	var msg sdk_struct.MsgStruct
	var notStartTime bool
	if req.ConversationID != "" {
		conversationID = req.ConversationID
		lc, err := c.db.GetConversation(conversationID)
		if err != nil {
			return nil
		}
		switch lc.ConversationType {
		case constant.SingleChatType, constant.NotificationChatType:
			sourceID = lc.UserID
		case constant.GroupChatType, constant.SuperGroupChatType:
			sourceID = lc.GroupID
			msg.GroupID = lc.GroupID
		}
		sessionType = int(lc.ConversationType)
		if req.StartClientMsgID == "" {
			//startTime = lc.LatestMsgSendTime + TimeOffset
			////startTime = utils.GetCurrentTimestampByMill()
			notStartTime = true
		} else {
			msg.SessionType = lc.ConversationType
			msg.ClientMsgID = req.StartClientMsgID
			m, err := c.db.GetMessageController(&msg)
			common.CheckDBErrCallback(callback, err, operationID)
			startTime = m.SendTime
		}
	} else {
		if req.UserID == "" {
			newConversationID, newSessionType, err := c.getConversationTypeByGroupID(req.GroupID)
			common.CheckDBErrCallback(callback, err, operationID)
			sourceID = req.GroupID
			sessionType = int(newSessionType)
			conversationID = newConversationID
			msg.GroupID = req.GroupID
			msg.SessionType = newSessionType
		} else {
			sourceID = req.UserID
			conversationID = utils.GetConversationIDBySessionType(sourceID, constant.SingleChatType)
			sessionType = constant.SingleChatType
		}
		if req.StartClientMsgID == "" {
			//lc, err := c.db.GetConversation(conversationID)
			//if err != nil {
			//	return nil
			//}
			//startTime = lc.LatestMsgSendTime + TimeOffset
			//startTime = utils.GetCurrentTimestampByMill()
			notStartTime = true
		} else {
			msg.ClientMsgID = req.StartClientMsgID
			m, err := c.db.GetMessageController(&msg)
			common.CheckDBErrCallback(callback, err, operationID)
			startTime = m.SendTime
		}
	}
	log.Debug(operationID, "Assembly parameters cost time", time.Since(t))
	t = time.Now()
	log.Info(operationID, "sourceID:", sourceID, "startTime:", startTime, "count:", req.Count, "not start_time", notStartTime)
	if notStartTime {
		list, err = c.db.GetMessageListNoTimeController(sourceID, sessionType, req.Count, isReverse)
	} else {
		list, err = c.db.GetMessageListController(sourceID, sessionType, req.Count, startTime, isReverse)
	}
	log.Debug(operationID, "db cost time", time.Since(t))
	common.CheckDBErrCallback(callback, err, operationID)
	t = time.Now()
	for _, v := range list {
		temp := sdk_struct.MsgStruct{}
		tt := time.Now()
		temp.ClientMsgID = v.ClientMsgID
		temp.ServerMsgID = v.ServerMsgID
		temp.CreateTime = v.CreateTime
		temp.SendTime = v.SendTime
		temp.SessionType = v.SessionType
		temp.SendID = v.SendID
		temp.RecvID = v.RecvID
		temp.MsgFrom = v.MsgFrom
		temp.ContentType = v.ContentType
		temp.SenderPlatformID = v.SenderPlatformID
		temp.SenderNickname = v.SenderNickname
		temp.SenderFaceURL = v.SenderFaceURL
		temp.Content = v.Content
		temp.Seq = v.Seq
		temp.IsRead = v.IsRead
		temp.Status = v.Status
		temp.AttachedInfo = v.AttachedInfo
		temp.Ex = v.Ex
		err := c.msgHandleByContentType(&temp)
		if err != nil {
			log.Error(operationID, "Parsing data error:", err.Error(), temp)
			continue
		}
		log.Debug(operationID, "internal unmarshal cost time", time.Since(tt))

		switch sessionType {
		case constant.GroupChatType:
			fallthrough
		case constant.SuperGroupChatType:
			temp.GroupID = temp.RecvID
			temp.RecvID = c.loginUserID
		}
		messageList = append(messageList, &temp)
	}
	log.Debug(operationID, "unmarshal cost time", time.Since(t))
	t = time.Now()
	if !isReverse {
		sort.Sort(messageList)
	}
	log.Debug(operationID, "sort cost time", time.Since(t))
	return sdk.GetHistoryMessageListCallback(messageList)
}
func (c *Conversation) getAdvancedHistoryMessageList(callback open_im_sdk_callback.Base, req sdk.GetAdvancedHistoryMessageListParams, operationID string, isReverse bool) sdk.GetAdvancedHistoryMessageListCallback {
	t := time.Now()
	var messageListCallback sdk.GetAdvancedHistoryMessageListCallback
	var sourceID string
	var conversationID string
	var startTime int64

	var sessionType int
	var list []*model_struct.LocalChatLog
	var err error
	var messageList sdk_struct.NewMsgList
	var msg sdk_struct.MsgStruct
	var notStartTime bool
	if req.ConversationID != "" {
		conversationID = req.ConversationID
		lc, err := c.db.GetConversation(conversationID)
		if err != nil {
			messageListCallback.ErrCode = 100
			messageListCallback.ErrMsg = "conversation get err"
			return messageListCallback
		}
		switch lc.ConversationType {
		case constant.SingleChatType, constant.NotificationChatType:
			sourceID = lc.UserID
		case constant.GroupChatType, constant.SuperGroupChatType:
			sourceID = lc.GroupID
			msg.GroupID = lc.GroupID
		}
		sessionType = int(lc.ConversationType)
		if req.StartClientMsgID == "" {
			//startTime = lc.LatestMsgSendTime + TimeOffset
			////startTime = utils.GetCurrentTimestampByMill()
			notStartTime = true
		} else {
			msg.SessionType = lc.ConversationType
			msg.ClientMsgID = req.StartClientMsgID
			m, err := c.db.GetMessageController(&msg)
			common.CheckDBErrCallback(callback, err, operationID)
			startTime = m.SendTime
		}
	} else {
		if req.UserID == "" {
			newConversationID, newSessionType, err := c.getConversationTypeByGroupID(req.GroupID)
			common.CheckDBErrCallback(callback, err, operationID)
			sourceID = req.GroupID
			sessionType = int(newSessionType)
			conversationID = newConversationID
			msg.GroupID = req.GroupID
			msg.SessionType = newSessionType
		} else {
			sourceID = req.UserID
			conversationID = utils.GetConversationIDBySessionType(sourceID, constant.SingleChatType)
			sessionType = constant.SingleChatType
		}
		if req.StartClientMsgID == "" {
			//lc, err := c.db.GetConversation(conversationID)
			//if err != nil {
			//	return nil
			//}
			//startTime = lc.LatestMsgSendTime + TimeOffset
			//startTime = utils.GetCurrentTimestampByMill()
			notStartTime = true
		} else {
			msg.ClientMsgID = req.StartClientMsgID
			m, err := c.db.GetMessageController(&msg)
			common.CheckDBErrCallback(callback, err, operationID)
			startTime = m.SendTime
		}
	}
	log.Debug(operationID, "Assembly parameters cost time", time.Since(t))
	t = time.Now()
	log.Info(operationID, "sourceID:", sourceID, "startTime:", startTime, "count:", req.Count, "not start_time", notStartTime)
	if notStartTime {
		list, err = c.db.GetMessageListNoTimeController(sourceID, sessionType, req.Count, isReverse)
	} else {
		list, err = c.db.GetMessageListController(sourceID, sessionType, req.Count, startTime, isReverse)
	}
	log.Debug(operationID, "db cost time", time.Since(t))
	t = time.Now()
	common.CheckDBErrCallback(callback, err, operationID)
	if len(list) < req.Count && sessionType == constant.SuperGroupChatType {
		seq, _ := c.db.SuperGroupGetNormalMinSeq(sourceID)
		log.Debug(operationID, sourceID+":table min seq is ", seq)
		if seq != 0 && seq != 1 {
			seqList := func(seq uint32) (seqList []uint32) {
				startSeq := int64(seq) - constant.PullMsgNumForReadDiffusion
				if startSeq <= 0 {
					startSeq = 1
				}
				log.Debug(operationID, "pull start is ", startSeq)
				for i := startSeq; i < int64(seq); i++ {
					seqList = append(seqList, uint32(i))
				}
				log.Debug(operationID, "pull seqList is ", seqList)
				return seqList
			}(seq)
			log.Debug(operationID, "pull seqList is ", seqList, len(seqList))
			if len(seqList) > 0 {
				c.pullMessageAndReGetHistoryMessages(sourceID, seqList, notStartTime, isReverse, req.Count, sessionType, startTime, &list, &messageListCallback, operationID)
			}
		}
	} else if len(list) == req.Count && sessionType == constant.SuperGroupChatType {
		maxSeq, minSeq, haveSeqList := func(messages []*model_struct.LocalChatLog) (max, min uint32, seqList []uint32) {
			for _, message := range messages {
				if message.Seq != 0 {
					max = message.Seq
					min = message.Seq
					break
				}
			}
			for i := 0; i < len(messages); i++ {
				if messages[i].Seq != 0 {
					seqList = append(seqList, messages[i].Seq)
				}
				if messages[i].Seq > max {
					max = messages[i].Seq

				}
				if messages[i].Seq < min {
					min = messages[i].Seq
				}
			}
			return max, min, seqList
		}(list)
		log.Debug(operationID, "get message from local db max seq:", maxSeq, "minSeq:", minSeq, "haveSeqList:", haveSeqList, "length:", len(haveSeqList))
		if maxSeq != 0 && minSeq != 0 {
			successiveSeqList := func(max, min uint32) (seqList []uint32) {
				for i := min; i <= max; i++ {
					seqList = append(seqList, i)
				}
				return seqList
			}(maxSeq, minSeq)
			lostSeqList := utils.DifferenceSubset(successiveSeqList, haveSeqList)
			lostSeqListLength := len(lostSeqList)
			log.Debug(operationID, "get lost seqList is :", lostSeqList, "length:", lostSeqListLength)
			if lostSeqListLength > 0 {
				var pullSeqList []uint32
				if lostSeqListLength <= constant.PullMsgNumForReadDiffusion {
					pullSeqList = lostSeqList
				} else {
					pullSeqList = lostSeqList[lostSeqListLength-constant.PullMsgNumForReadDiffusion : lostSeqListLength]
				}
				c.pullMessageAndReGetHistoryMessages(sourceID, pullSeqList, notStartTime, isReverse, req.Count, sessionType, startTime, &list, &messageListCallback, operationID)
			} else {
				if req.LastMinSeq != 0 {
					var thisMaxSeq uint32
					for i := 0; i < len(list); i++ {
						if list[i].Seq != 0 && thisMaxSeq == 0 {
							thisMaxSeq = list[i].Seq
						}
						if list[i].Seq > thisMaxSeq {
							thisMaxSeq = list[i].Seq
						}
					}
					log.Debug(operationID, "get lost LastMinSeq is :", req.LastMinSeq, "thisMaxSeq is :", thisMaxSeq)
					if thisMaxSeq != 0 {
						if thisMaxSeq+1 != req.LastMinSeq {
							startSeq := int64(req.LastMinSeq) - constant.PullMsgNumForReadDiffusion
							if startSeq <= int64(thisMaxSeq) {
								startSeq = int64(thisMaxSeq) + 1
							}
							successiveSeqList := func(max, min uint32) (seqList []uint32) {
								for i := min; i <= max; i++ {
									seqList = append(seqList, i)
								}
								return seqList
							}(req.LastMinSeq-1, uint32(startSeq))
							log.Debug(operationID, "get lost successiveSeqList is :", successiveSeqList, len(successiveSeqList))
							if len(successiveSeqList) > 0 {
								c.pullMessageAndReGetHistoryMessages(sourceID, successiveSeqList, notStartTime, isReverse, req.Count, sessionType, startTime, &list, &messageListCallback, operationID)
							}
						}

					}

				}
			}

		}
	}
	log.Debug(operationID, "pull cost time", time.Since(t))
	t = time.Now()
	var thisMinSeq uint32
	for _, v := range list {
		if v.Seq != 0 && thisMinSeq == 0 {
			thisMinSeq = v.Seq
		}
		if v.Seq < thisMinSeq {
			thisMinSeq = v.Seq
		}
		temp := sdk_struct.MsgStruct{}
		tt := time.Now()
		temp.ClientMsgID = v.ClientMsgID
		temp.ServerMsgID = v.ServerMsgID
		temp.CreateTime = v.CreateTime
		temp.SendTime = v.SendTime
		temp.SessionType = v.SessionType
		temp.SendID = v.SendID
		temp.RecvID = v.RecvID
		temp.MsgFrom = v.MsgFrom
		temp.ContentType = v.ContentType
		temp.SenderPlatformID = v.SenderPlatformID
		temp.SenderNickname = v.SenderNickname
		temp.SenderFaceURL = v.SenderFaceURL
		temp.Content = v.Content
		temp.Seq = v.Seq
		temp.IsRead = v.IsRead
		temp.Status = v.Status
		temp.AttachedInfo = v.AttachedInfo
		temp.Ex = v.Ex
		err := c.msgHandleByContentType(&temp)
		if err != nil {
			log.Error(operationID, "Parsing data error:", err.Error(), temp)
			continue
		}
		log.Debug(operationID, "internal unmarshal cost time", time.Since(tt))

		switch sessionType {
		case constant.GroupChatType:
			fallthrough
		case constant.SuperGroupChatType:
			temp.GroupID = temp.RecvID
			temp.RecvID = c.loginUserID
		}
		messageList = append(messageList, &temp)
	}
	log.Debug(operationID, "unmarshal cost time", time.Since(t))
	t = time.Now()
	if !isReverse {
		sort.Sort(messageList)
	}
	log.Debug(operationID, "sort cost time", time.Since(t))
	messageListCallback.MessageList = messageList
	messageListCallback.LastMinSeq = thisMinSeq
	return messageListCallback
}
func (c *Conversation) pullMessageAndReGetHistoryMessages(sourceID string, seqList []uint32, notStartTime, isReverse bool, count, sessionType int, startTime int64, list *[]*model_struct.LocalChatLog, messageListCallback *sdk.GetAdvancedHistoryMessageListCallback, operationID string) {
	var pullMsgReq server_api_params.PullMessageBySeqListReq
	pullMsgReq.UserID = c.loginUserID
	pullMsgReq.GroupSeqList = make(map[string]*server_api_params.SeqList, 0)
	pullMsgReq.GroupSeqList[sourceID] = &server_api_params.SeqList{SeqList: seqList}

	pullMsgReq.OperationID = operationID
	log.Debug(operationID, "read diffusion group pull message, req: ", pullMsgReq)
	resp, err := c.SendReqWaitResp(&pullMsgReq, constant.WSPullMsgBySeqList, 30, 3, c.loginUserID, operationID)
	if err != nil {
		messageListCallback.ErrCode = 100
		messageListCallback.ErrMsg = err.Error()
		log.Error(operationID, "SendReqWaitResp failed ", err.Error(), constant.WSPullMsgBySeqList, 30, 3, c.loginUserID)
	} else {
		var pullMsgResp server_api_params.PullMessageBySeqListResp
		err = proto.Unmarshal(resp.Data, &pullMsgResp)
		if err != nil {
			messageListCallback.ErrCode = 100
			messageListCallback.ErrMsg = err.Error()
			log.Error(operationID, "pullMsgResp Unmarshal failed ", err.Error())
		} else {
			log.Debug(operationID, "syncMsgFromServerSplit pull msg ", pullMsgReq.String(), pullMsgResp.String())
			if v, ok := pullMsgResp.GroupMsgDataList[sourceID]; ok {
				c.pullMessageIntoTable(v.MsgDataList, operationID)
			}
			if notStartTime {
				*list, err = c.db.GetMessageListNoTimeController(sourceID, sessionType, count, isReverse)
			} else {
				*list, err = c.db.GetMessageListController(sourceID, sessionType, count, startTime, isReverse)
			}
		}

	}
}
func (c *Conversation) pullMessageIntoTable(pullMsgData []*server_api_params.MsgData, operationID string) {

	var insertMsg, specialUpdateMsg []*model_struct.LocalChatLog
	var exceptionMsg []*model_struct.LocalErrChatLog
	var msgReadList, groupMsgReadList, msgRevokeList, newMsgRevokeList sdk_struct.NewMsgList

	log.Info(operationID, "do Msg come here, len: ", len(pullMsgData))
	b := utils.GetCurrentTimestampByMill()

	for _, v := range pullMsgData {
		log.Info(operationID, "do Msg come here, msg detail ", v.RecvID, v.SendID, v.ClientMsgID, v.ServerMsgID, v.Seq, c.loginUserID)
		msg := new(sdk_struct.MsgStruct)
		copier.Copy(msg, v)
		var tips server_api_params.TipsComm
		if v.ContentType >= constant.NotificationBegin && v.ContentType <= constant.NotificationEnd {
			_ = proto.Unmarshal(v.Content, &tips)
			marshaler := jsonpb.Marshaler{
				OrigName:     true,
				EnumsAsInts:  false,
				EmitDefaults: false,
			}
			msg.Content, _ = marshaler.MarshalToString(&tips)
		} else {
			msg.Content = string(v.Content)
		}
		//When the message has been marked and deleted by the cloud, it is directly inserted locally without any conversation and message update.
		if msg.Status == constant.MsgStatusHasDeleted {
			insertMsg = append(insertMsg, c.msgStructToLocalChatLog(msg))
			continue
		}
		msg.Status = constant.MsgStatusSendSuccess
		msg.IsRead = false
		//		log.Info(operationID, "new msg, seq, ServerMsgID, ClientMsgID", msg.Seq, msg.ServerMsgID, msg.ClientMsgID)
		//De-analyze data
		if msg.ClientMsgID == "" {
			exceptionMsg = append(exceptionMsg, c.msgStructToLocalErrChatLog(msg))
			continue
		}
		switch {
		case v.ContentType == constant.ConversationChangeNotification || v.ContentType == constant.ConversationPrivateChatNotification:
			log.Info(operationID, utils.GetSelfFuncName(), v)
			c.DoNotification(v)
		case v.ContentType == constant.SuperGroupUpdateNotification:
			c.full.SuperGroup.DoNotification(v, c.GetCh())
		}
		if v.SendID == c.loginUserID { //seq
			// Messages sent by myself  //if  sent through  this terminal
			m, err := c.db.GetMessageController(msg)
			if err == nil {
				log.Info(operationID, "have message", msg.Seq, msg.ServerMsgID, msg.ClientMsgID, *msg)
				if m.Seq == 0 {
					specialUpdateMsg = append(specialUpdateMsg, c.msgStructToLocalChatLog(msg))

				} else {
					exceptionMsg = append(exceptionMsg, c.msgStructToLocalErrChatLog(msg))
				}
			} else { //      send through  other terminal
				log.Info(operationID, "sync message", msg.Seq, msg.ServerMsgID, msg.ClientMsgID, *msg)
				insertMsg = append(insertMsg, c.msgStructToLocalChatLog(msg))
				switch msg.ContentType {
				case constant.Revoke:
					msgRevokeList = append(msgRevokeList, msg)
				case constant.HasReadReceipt:
					msgReadList = append(msgReadList, msg)
				case constant.GroupHasReadReceipt:
					groupMsgReadList = append(groupMsgReadList, msg)
				case constant.AdvancedRevoke:
					newMsgRevokeList = append(newMsgRevokeList, msg)
				default:
				}
			}
		} else { //Sent by others
			if oldMessage, err := c.db.GetMessageController(msg); err != nil { //Deduplication operation
				log.Warn("test", "trigger msg is ", msg.SenderNickname, msg.SenderFaceURL)
				insertMsg = append(insertMsg, c.msgStructToLocalChatLog(msg))
				switch msg.ContentType {
				case constant.Revoke:
					msgRevokeList = append(msgRevokeList, msg)
				case constant.HasReadReceipt:
					msgReadList = append(msgReadList, msg)
				case constant.GroupHasReadReceipt:
					groupMsgReadList = append(groupMsgReadList, msg)
				case constant.AdvancedRevoke:
					newMsgRevokeList = append(newMsgRevokeList, msg)

				default:
				}

			} else {
				if oldMessage.Seq == 0 {
					specialUpdateMsg = append(specialUpdateMsg, c.msgStructToLocalChatLog(msg))
				} else {
					exceptionMsg = append(exceptionMsg, c.msgStructToLocalErrChatLog(msg))
					log.Warn(operationID, "Deduplication operation ", *c.msgStructToLocalErrChatLog(msg))
				}
			}
		}
	}

	//update message
	err6 := c.db.BatchSpecialUpdateMessageList(specialUpdateMsg)
	if err6 != nil {
		log.Error(operationID, "sync seq normal message err  :", err6.Error())
	}
	b3 := utils.GetCurrentTimestampByMill()
	//Normal message storage
	err1 := c.db.BatchInsertMessageListController(insertMsg)
	if err1 != nil {
		log.Error(operationID, "insert GetMessage detail err:", err1.Error(), len(insertMsg))
		for _, v := range insertMsg {
			e := c.db.InsertMessageController(v)
			if e != nil {
				errChatLog := &model_struct.LocalErrChatLog{}
				copier.Copy(errChatLog, v)
				exceptionMsg = append(exceptionMsg, errChatLog)
				log.Warn(operationID, "InsertMessage operation ", "chat err log: ", errChatLog, "chat log: ", v, e.Error())
			}
		}
	}
	b4 := utils.GetCurrentTimestampByMill()
	log.Debug(operationID, "BatchInsertMessageListController, cost time : ", b4-b3)

	//Exception message storage
	for _, v := range exceptionMsg {
		log.Warn(operationID, "exceptionMsg show: ", *v)
	}

	err2 := c.db.BatchInsertExceptionMsgController(exceptionMsg)
	if err2 != nil {
		log.Error(operationID, "BatchInsertExceptionMsgController err message err  :", err2.Error())

	}
	b8 := utils.GetCurrentTimestampByMill()
	c.DoGroupMsgReadState(groupMsgReadList)
	b9 := utils.GetCurrentTimestampByMill()
	log.Debug(operationID, "DoGroupMsgReadState  cost time : ", b9-b8, "len: ", len(groupMsgReadList))

	c.revokeMessage(msgRevokeList)
	c.newRevokeMessage(newMsgRevokeList)
	b10 := utils.GetCurrentTimestampByMill()
	log.Debug(operationID, "revokeMessage  cost time : ", b10-b9)
	log.Info(operationID, "insert msg, total cost time: ", utils.GetCurrentTimestampByMill()-b, "len:  ", len(pullMsgData))
}

func (c *Conversation) revokeOneMessage(callback open_im_sdk_callback.Base, req sdk.RevokeMessageParams, operationID string) {
	var recvID, groupID string
	var localMessage model_struct.LocalChatLog
	var lc model_struct.LocalConversation
	var conversationID string
	message, err := c.db.GetMessageController((*sdk_struct.MsgStruct)(&req))
	common.CheckDBErrCallback(callback, err, operationID)
	if message.Status != constant.MsgStatusSendSuccess {
		common.CheckAnyErrCallback(callback, 201, errors.New("only send success message can be revoked"), operationID)
	}
	if message.SendID != c.loginUserID {
		common.CheckAnyErrCallback(callback, 201, errors.New("only you send message can be revoked"), operationID)
	}
	//Send message internally
	switch req.SessionType {
	case constant.SingleChatType:
		recvID = req.RecvID
		conversationID = utils.GetConversationIDBySessionType(recvID, constant.SingleChatType)
	case constant.GroupChatType:
		groupID = req.GroupID
		conversationID = utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
	case constant.SuperGroupChatType:
		groupID = req.GroupID
		conversationID = utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType)
	default:
		common.CheckAnyErrCallback(callback, 201, errors.New("SessionType err"), operationID)
	}
	req.Content = message.ClientMsgID
	req.ClientMsgID = utils.GetMsgID(message.SendID)
	req.ContentType = constant.Revoke
	req.SendTime = 0
	req.CreateTime = utils.GetCurrentTimestampByMill()
	options := make(map[string]bool, 5)
	utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
	utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	resp, _ := c.InternalSendMessage(callback, (*sdk_struct.MsgStruct)(&req), recvID, groupID, operationID, &server_api_params.OfflinePushInfo{}, false, options)
	req.ServerMsgID = resp.ServerMsgID
	req.SendTime = resp.SendTime
	req.Status = constant.MsgStatusSendSuccess
	msgStructToLocalChatLog(&localMessage, (*sdk_struct.MsgStruct)(&req))
	err = c.db.InsertMessageController(&localMessage)
	if err != nil {
		log.Error(operationID, "inset into chat log err", localMessage, req)
	}
	err = c.db.UpdateColumnsMessageController(req.Content, groupID, req.SessionType, map[string]interface{}{"status": constant.MsgStatusRevoked})
	if err != nil {
		log.Error(operationID, "update revoke message err", localMessage, req)
	}
	lc.LatestMsg = utils.StructToJsonString(req)
	lc.LatestMsgSendTime = req.SendTime
	lc.ConversationID = conversationID
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: lc}, c.GetCh())
}
func (c *Conversation) newRevokeOneMessage(callback open_im_sdk_callback.Base, req sdk.RevokeMessageParams, operationID string) {
	var recvID, groupID string
	var localMessage model_struct.LocalChatLog
	var revokeMessage sdk_struct.MessageRevoked
	var lc model_struct.LocalConversation
	var conversationID string
	message, err := c.db.GetMessageController((*sdk_struct.MsgStruct)(&req))
	common.CheckDBErrCallback(callback, err, operationID)
	if message.Status != constant.MsgStatusSendSuccess {
		common.CheckAnyErrCallback(callback, 201, errors.New("only send success message can be revoked"), operationID)
	}
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.AdvancedRevoke, operationID)
	revokeMessage.ClientMsgID = message.ClientMsgID
	revokeMessage.RevokerID = c.loginUserID
	revokeMessage.RevokeTime = utils.GetCurrentTimestampBySecond()
	revokeMessage.RevokerNickname = s.SenderNickname
	revokeMessage.SourceMessageSendTime = message.SendTime
	revokeMessage.SessionType = message.SessionType
	revokeMessage.SourceMessageSendID = message.SendID
	revokeMessage.SourceMessageSenderNickname = message.SenderNickname
	//Send message internally
	switch message.SessionType {
	case constant.SingleChatType:
		if message.SendID != c.loginUserID {
			common.CheckAnyErrCallback(callback, 201, errors.New("only you send message can be revoked"), operationID)
		}
		recvID = message.RecvID
		conversationID = utils.GetConversationIDBySessionType(recvID, constant.SingleChatType)
	case constant.GroupChatType:
		if message.SendID != c.loginUserID {
			ownerID, adminIDList, err := c.group.GetGroupOwnerIDAndAdminIDList(message.RecvID, operationID)
			common.CheckDBErrCallback(callback, err, operationID)
			if c.loginUserID == ownerID {
				revokeMessage.RevokerRole = constant.GroupOwner
			} else if utils.IsContain(c.loginUserID, adminIDList) {
				if utils.IsContain(message.SendID, adminIDList) || message.SendID == ownerID {
					common.CheckAnyErrCallback(callback, 201, errors.New("you do not have this permission"), operationID)
				} else {
					revokeMessage.RevokerRole = constant.GroupAdmin
				}

			} else {
				common.CheckAnyErrCallback(callback, 201, errors.New("you do not have this permission"), operationID)
			}
		}
		groupID = message.RecvID
		conversationID = utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
	case constant.SuperGroupChatType:
		if message.SendID != c.loginUserID {
			ownerID, adminIDList, err := c.group.GetGroupOwnerIDAndAdminIDList(message.RecvID, operationID)
			common.CheckDBErrCallback(callback, err, operationID)
			if c.loginUserID == ownerID {
				revokeMessage.RevokerRole = constant.GroupOwner
			} else if utils.IsContain(c.loginUserID, adminIDList) {
				if utils.IsContain(message.SendID, adminIDList) || message.SendID == ownerID {
					common.CheckAnyErrCallback(callback, 201, errors.New("you do not have this permission"), operationID)
				} else {
					revokeMessage.RevokerRole = constant.GroupAdmin
				}

			} else {
				common.CheckAnyErrCallback(callback, 201, errors.New("you do not have this permission"), operationID)
			}
		}
		groupID = message.RecvID
		conversationID = utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType)
	default:
		common.CheckAnyErrCallback(callback, 201, errors.New("SessionType err"), operationID)
	}
	s.Content = utils.StructToJsonString(revokeMessage)
	options := make(map[string]bool, 5)
	utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
	utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	resp, _ := c.InternalSendMessage(callback, &s, recvID, groupID, operationID, &server_api_params.OfflinePushInfo{}, false, options)
	s.ServerMsgID = resp.ServerMsgID
	s.SendTime = message.SendTime //New message takes the old place
	s.Status = constant.MsgStatusSendSuccess
	msgStructToLocalChatLog(&localMessage, &s)
	err = c.db.InsertMessageController(&localMessage)
	if err != nil {
		log.Error(operationID, "inset into chat log err", localMessage, s)
	}
	err = c.db.UpdateColumnsMessageController(message.ClientMsgID, groupID, message.SessionType, map[string]interface{}{"status": constant.MsgStatusRevoked})
	if err != nil {
		log.Error(operationID, "update revoke message err", localMessage, message)
	}
	s.SendTime = resp.SendTime
	lc.LatestMsg = utils.StructToJsonString(s)
	lc.LatestMsgSendTime = s.SendTime
	lc.ConversationID = conversationID
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: lc}, c.GetCh())
}

func (c *Conversation) typingStatusUpdate(callback open_im_sdk_callback.Base, recvID, msgTip, operationID string) {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Typing, operationID)
	s.Content = msgTip
	options := make(map[string]bool, 6)
	utils.SetSwitchFromOptions(options, constant.IsHistory, false)
	utils.SetSwitchFromOptions(options, constant.IsPersistent, false)
	utils.SetSwitchFromOptions(options, constant.IsSenderSync, false)
	utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
	utils.SetSwitchFromOptions(options, constant.IsSenderConversationUpdate, false)
	utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
	utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	c.InternalSendMessage(callback, &s, recvID, "", operationID, &server_api_params.OfflinePushInfo{}, true, options)

}

func (c *Conversation) markC2CMessageAsRead(callback open_im_sdk_callback.Base, msgIDList sdk.MarkC2CMessageAsReadParams, userID, operationID string) {
	var localMessage model_struct.LocalChatLog
	var newMessageIDList []string
	messages, err := c.db.GetMultipleMessage(msgIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	for _, v := range messages {
		if v.IsRead == false && v.ContentType < constant.NotificationBegin && v.SendID != c.loginUserID {
			newMessageIDList = append(newMessageIDList, v.ClientMsgID)
		}
	}
	if len(newMessageIDList) == 0 {
		common.CheckAnyErrCallback(callback, 201, errors.New("message has been marked read or sender is yourself or notification message not support"), operationID)
	}
	conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.HasReadReceipt, operationID)
	s.Content = utils.StructToJsonString(newMessageIDList)
	options := make(map[string]bool, 5)
	utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
	utils.SetSwitchFromOptions(options, constant.IsSenderConversationUpdate, false)
	utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
	utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
	//If there is an error, the coroutine ends, so judgment is not  required
	resp, _ := c.InternalSendMessage(callback, &s, userID, "", operationID, &server_api_params.OfflinePushInfo{}, false, options)
	s.ServerMsgID = resp.ServerMsgID
	s.SendTime = resp.SendTime
	s.Status = constant.MsgStatusFiltered
	msgStructToLocalChatLog(&localMessage, &s)
	err = c.db.InsertMessage(&localMessage)
	if err != nil {
		log.Error(operationID, "inset into chat log err", localMessage, s, err.Error())
	}

	err2 := c.db.UpdateSingleMessageHasRead(userID, newMessageIDList)
	if err2 != nil {
		log.Error(operationID, "update message has read error", newMessageIDList, userID, err2.Error())
	}
	newMessages, err3 := c.db.GetMultipleMessage(newMessageIDList)
	if err3 != nil {
		log.Error(operationID, "get messages error", newMessageIDList, userID, err3.Error())
	}
	for _, v := range newMessages {
		attachInfo := sdk_struct.AttachedInfoElem{}
		_ = utils.JsonStringToStruct(v.AttachedInfo, &attachInfo)
		attachInfo.HasReadTime = s.SendTime
		v.AttachedInfo = utils.StructToJsonString(attachInfo)
		err = c.db.UpdateMessage(v)
		if err != nil {
			log.Error("internal", "setMessageHasReadByMsgID err:", err, "ClientMsgID", v)
			continue
		}
	}
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UpdateLatestMessageChange}, c.GetCh())
	//_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.ch)
}
func (c *Conversation) markGroupMessageAsRead(callback open_im_sdk_callback.Base, msgIDList sdk.MarkGroupMessageAsReadParams, groupID, operationID string) {
	conversationID, conversationType, err := c.getConversationTypeByGroupID(groupID)
	common.CheckAnyErrCallback(callback, 202, err, operationID)
	if len(msgIDList) == 0 {
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UnreadCountSetZero}, c.GetCh())
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
		return
	}
	var localMessage model_struct.LocalChatLog
	allUserMessage := make(map[string][]string, 3)
	messages, err := c.db.GetMultipleMessageController(msgIDList, groupID, conversationType)
	common.CheckDBErrCallback(callback, err, operationID)
	for _, v := range messages {
		log.Debug(operationID, "get group info is test2", v.ClientMsgID, v.SessionType)
		if v.IsRead == false && v.ContentType < constant.NotificationBegin && v.SendID != c.loginUserID {
			if msgIDList, ok := allUserMessage[v.SendID]; ok {
				msgIDList = append(msgIDList, v.ClientMsgID)
				allUserMessage[v.SendID] = msgIDList
			} else {
				allUserMessage[v.SendID] = []string{v.ClientMsgID}
			}
		}
	}
	if len(allUserMessage) == 0 {
		common.CheckAnyErrCallback(callback, 201, errors.New("message has been marked read or sender is yourself or notification message not support"), operationID)
	}

	for userID, list := range allUserMessage {
		s := sdk_struct.MsgStruct{}
		s.GroupID = groupID
		c.initBasicInfo(&s, constant.UserMsgType, constant.GroupHasReadReceipt, operationID)
		s.Content = utils.StructToJsonString(list)
		options := make(map[string]bool, 5)
		utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
		utils.SetSwitchFromOptions(options, constant.IsSenderConversationUpdate, false)
		utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
		utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
		//If there is an error, the coroutine ends, so judgment is not  required
		resp, _ := c.InternalSendMessage(callback, &s, userID, "", operationID, &server_api_params.OfflinePushInfo{}, false, options)
		s.ServerMsgID = resp.ServerMsgID
		s.SendTime = resp.SendTime
		s.Status = constant.MsgStatusFiltered
		msgStructToLocalChatLog(&localMessage, &s)
		err = c.db.InsertMessageController(&localMessage)
		if err != nil {
			log.Error(operationID, "inset into chat log err", localMessage, s, err.Error())
		}
		log.Debug(operationID, "get group info is test3", list, conversationType)
		err2 := c.db.UpdateGroupMessageHasReadController(list, groupID, conversationType)
		if err2 != nil {
			log.Error(operationID, "update message has read err", list, userID, err2.Error())
		}
	}
}

//func (c *Conversation) markMessageAsReadByConID(callback open_im_sdk_callback.Base, msgIDList sdk.MarkMessageAsReadByConIDParams, conversationID, operationID string) {
//	var localMessage db.LocalChatLog
//	var newMessageIDList []string
//	messages, err := c.db.GetMultipleMessage(msgIDList)
//	common.CheckDBErrCallback(callback, err, operationID)
//	for _, v := range messages {
//		if v.IsRead == false && v.ContentType < constant.NotificationBegin && v.SendID != c.loginUserID {
//			newMessageIDList = append(newMessageIDList, v.ClientMsgID)
//		}
//	}
//	if len(newMessageIDList) == 0 {
//		common.CheckAnyErrCallback(callback, 201, errors.New("message has been marked read or sender is yourself"), operationID)
//	}
//	conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
//	s := sdk_struct.MsgStruct{}
//	c.initBasicInfo(&s, constant.UserMsgType, constant.HasReadReceipt, operationID)
//	s.Content = utils.StructToJsonString(newMessageIDList)
//	options := make(map[string]bool, 5)
//	utils.SetSwitchFromOptions(options, constant.IsConversationUpdate, false)
//	utils.SetSwitchFromOptions(options, constant.IsSenderConversationUpdate, false)
//	utils.SetSwitchFromOptions(options, constant.IsUnreadCount, false)
//	utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
//	//If there is an error, the coroutine ends, so judgment is not  required
//	resp, _ := c.InternalSendMessage(callback, &s, userID, "", operationID, &server_api_params.OfflinePushInfo{}, false, options)
//	s.ServerMsgID = resp.ServerMsgID
//	s.SendTime = resp.SendTime
//	s.Status = constant.MsgStatusFiltered
//	msgStructToLocalChatLog(&localMessage, &s)
//	err = c.db.InsertMessage(&localMessage)
//	if err != nil {
//		log.Error(operationID, "inset into chat log err", localMessage, s, err.Error())
//	}
//	err2 := c.db.UpdateMessageHasRead(userID, newMessageIDList, constant.SingleChatType)
//	if err2 != nil {
//		log.Error(operationID, "update message has read error", newMessageIDList, userID, err2.Error())
//	}
//	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UpdateLatestMessageChange}, c.ch)
//	//_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.ch)
//}
func (c *Conversation) insertMessageToLocalStorage(callback open_im_sdk_callback.Base, s *model_struct.LocalChatLog, operationID string) string {
	err := c.db.InsertMessageController(s)
	common.CheckDBErrCallback(callback, err, operationID)
	return s.ClientMsgID
}

func (c *Conversation) clearGroupHistoryMessage(callback open_im_sdk_callback.Base, groupID string, operationID string) {
	_, sessionType, err := c.getConversationTypeByGroupID(groupID)
	common.CheckAnyErrCallback(callback, 202, err, operationID)

	conversationID := utils.GetConversationIDBySessionType(groupID, int(sessionType))
	err = c.db.UpdateMessageStatusBySourceIDController(groupID, constant.MsgStatusHasDeleted, sessionType)
	common.CheckDBErrCallback(callback, err, operationID)
	err = c.db.ClearConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())

}

func (c *Conversation) clearC2CHistoryMessage(callback open_im_sdk_callback.Base, userID string, operationID string) {
	conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
	err := c.db.UpdateMessageStatusBySourceID(userID, constant.MsgStatusHasDeleted, constant.SingleChatType)
	common.CheckDBErrCallback(callback, err, operationID)
	err = c.db.ClearConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
}

func (c *Conversation) deleteMessageFromSvr(callback open_im_sdk_callback.Base, s *sdk_struct.MsgStruct, operationID string) {
	seq, err := c.db.GetMsgSeqByClientMsgIDController(s)
	switch s.SessionType {
	case constant.SingleChatType, constant.GroupChatType:
		var apiReq server_api_params.DeleteMsgReq
		common.CheckDBErrCallback(callback, err, operationID)
		apiReq.SeqList = []uint32{seq}
		apiReq.OpUserID = c.loginUserID
		apiReq.UserID = c.loginUserID
		apiReq.OperationID = operationID
		c.p.PostFatalCallback(callback, constant.DeleteMsgRouter, apiReq, nil, apiReq.OperationID)
	case constant.SuperGroupChatType:
		var apiReq server_api_params.DelSuperGroupMsgReq
		apiReq.UserID = c.loginUserID
		apiReq.IsAllDelete = false
		apiReq.GroupID = s.GroupID
		apiReq.OperationID = operationID
		apiReq.SeqList = []uint32{seq}
		c.p.PostFatalCallback(callback, constant.DeleteSuperGroupMsgRouter, apiReq, nil, apiReq.OperationID)
		return
	}

}

func (c *Conversation) clearMessageFromSvr(callback open_im_sdk_callback.Base, operationID string) {
	var apiReq server_api_params.CleanUpMsgReq
	apiReq.UserID = c.loginUserID
	apiReq.OperationID = operationID
	c.p.PostFatalCallback(callback, constant.ClearMsgRouter, apiReq, nil, apiReq.OperationID)
}

func (c *Conversation) deleteMessageFromLocalStorage(callback open_im_sdk_callback.Base, s *sdk_struct.MsgStruct, operationID string) {
	var conversation model_struct.LocalConversation
	var latestMsg sdk_struct.MsgStruct
	var conversationID string
	var sourceID string
	chatLog := model_struct.LocalChatLog{ClientMsgID: s.ClientMsgID, Status: constant.MsgStatusHasDeleted, SessionType: s.SessionType}

	switch s.SessionType {
	case constant.GroupChatType:
		conversationID = utils.GetConversationIDBySessionType(s.GroupID, constant.GroupChatType)
		sourceID = s.GroupID
	case constant.SingleChatType:
		if s.SendID != c.loginUserID {
			conversationID = utils.GetConversationIDBySessionType(s.SendID, constant.SingleChatType)
			sourceID = s.SendID
		} else {
			conversationID = utils.GetConversationIDBySessionType(s.RecvID, constant.SingleChatType)
			sourceID = s.RecvID
		}
	case constant.SuperGroupChatType:
		conversationID = utils.GetConversationIDBySessionType(s.GroupID, constant.SuperGroupChatType)
		sourceID = s.GroupID
		chatLog.RecvID = s.GroupID
	}
	err := c.db.UpdateMessageController(&chatLog)
	common.CheckDBErrCallback(callback, err, operationID)
	LocalConversation, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	common.JsonUnmarshalCallback(LocalConversation.LatestMsg, &latestMsg, callback, operationID)

	if s.ClientMsgID == latestMsg.ClientMsgID { //If the deleted message is the latest message of the conversation, update the latest message of the conversation
		list, err := c.db.GetMessageListNoTimeController(sourceID, int(s.SessionType), 1, false)
		common.CheckDBErrCallback(callback, err, operationID)

		conversation.ConversationID = conversationID
		if list == nil {
			conversation.LatestMsg = ""
			conversation.LatestMsgSendTime = s.SendTime
		} else {
			copier.Copy(&latestMsg, list[0])
			err := c.msgConvert(&latestMsg)
			if err != nil {
				log.Error(operationID, "Parsing data error:", err.Error(), latestMsg)
			}
			conversation.LatestMsg = utils.StructToJsonString(latestMsg)
			conversation.LatestMsgSendTime = latestMsg.SendTime
		}
		err = c.db.UpdateColumnsConversation(conversation.ConversationID, map[string]interface{}{"latest_msg_send_time": conversation.LatestMsgSendTime, "latest_msg": conversation.LatestMsg})
		if err != nil {
			log.Error("internal", "updateConversationLatestMsgModel err: ", err)
		} else {
			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
		}
	}
}
func (c *Conversation) judgeMultipleSubString(keywordList []string, main string, keywordListMatchType int) bool {
	if len(keywordList) == 0 {
		return true
	}
	if keywordListMatchType == constant.KeywordMatchOr {
		for _, v := range keywordList {
			if utils.KMP(main, v) {
				return true
			}
		}
		return false
	} else {
		for _, v := range keywordList {
			if !utils.KMP(main, v) {
				return false
			}
		}
	}
	return true
}

func (c *Conversation) searchLocalMessages(callback open_im_sdk_callback.Base, searchParam sdk.SearchLocalMessagesParams, operationID string) (r sdk.SearchLocalMessagesCallback) {

	var conversationID, sourceID string
	var startTime, endTime int64
	var list []*model_struct.LocalChatLog
	conversationMap := make(map[string]*sdk.SearchByConversationResult, 10)
	var err error

	if searchParam.SearchTimePosition == 0 {
		endTime = utils.GetCurrentTimestampBySecond()
	} else {
		endTime = searchParam.SearchTimePosition
	}
	if searchParam.SearchTimePeriod != 0 {
		startTime = endTime - searchParam.SearchTimePeriod
	}
	startTime = utils.UnixSecondToTime(startTime).UnixNano() / 1e6
	endTime = utils.UnixSecondToTime(endTime).UnixNano() / 1e6
	if len(searchParam.KeywordList) == 0 && len(searchParam.MessageTypeList) == 0 {
		common.CheckAnyErrCallback(callback, 201, errors.New("keywordlist and messageTypelist all null"), operationID)
	}
	if searchParam.ConversationID != "" {
		if searchParam.PageIndex < 1 || searchParam.Count < 1 {
			common.CheckAnyErrCallback(callback, 201, errors.New("page or count is null"), operationID)
		}
		offset := (searchParam.PageIndex - 1) * searchParam.Count
		localConversation, err := c.db.GetConversation(searchParam.ConversationID)
		common.CheckDBErrCallback(callback, err, operationID)
		switch localConversation.ConversationType {
		case constant.SingleChatType:
			sourceID = localConversation.UserID
		case constant.GroupChatType:
			sourceID = localConversation.GroupID
		case constant.SuperGroupChatType:
			sourceID = localConversation.GroupID
		}
		if len(searchParam.MessageTypeList) != 0 && len(searchParam.KeywordList) == 0 {
			list, err = c.db.SearchMessageByContentTypeController(searchParam.MessageTypeList, sourceID, startTime, endTime, int(localConversation.ConversationType), offset, searchParam.Count)
		} else {
			newContentTypeList := func(list []int) (result []int) {
				for _, v := range list {
					if utils.IsContainInt(v, SearchContentType) {
						result = append(result, v)
					}
				}
				return result
			}(searchParam.MessageTypeList)
			if len(newContentTypeList) == 0 {
				newContentTypeList = SearchContentType
			}
			list, err = c.db.SearchMessageByKeywordController(newContentTypeList, searchParam.KeywordList, searchParam.KeywordListMatchType, sourceID, startTime, endTime, int(localConversation.ConversationType), offset, searchParam.Count)
		}
	} else {
		//Comprehensive search, search all
		if len(searchParam.MessageTypeList) == 0 {
			searchParam.MessageTypeList = SearchContentType
		}
		list, err = c.db.SearchMessageByContentTypeAndKeywordController(searchParam.MessageTypeList, searchParam.KeywordList, searchParam.KeywordListMatchType, startTime, endTime, operationID)
	}
	common.CheckDBErrCallback(callback, err, operationID)
	//localChatLogToMsgStruct(&messageList, list)

	//log.Debug("hahh",utils.KMP("SSSsdf3434","s"))
	//log.Debug("hahh",utils.KMP("SSSsdf3434","g"))
	//log.Debug("hahh",utils.KMP("SSSsdf3434","3434"))
	//log.Debug("hahh",utils.KMP("SSSsdf3434","F3434"))
	//log.Debug("hahh",utils.KMP("SSSsdf3434","SDF3"))
	log.Debug("get raw data length is", len(list))
	for _, v := range list {
		temp := sdk_struct.MsgStruct{}
		temp.ClientMsgID = v.ClientMsgID
		temp.ServerMsgID = v.ServerMsgID
		temp.CreateTime = v.CreateTime
		temp.SendTime = v.SendTime
		temp.SessionType = v.SessionType
		temp.SendID = v.SendID
		temp.RecvID = v.RecvID
		temp.MsgFrom = v.MsgFrom
		temp.ContentType = v.ContentType
		temp.SenderPlatformID = v.SenderPlatformID
		temp.SenderNickname = v.SenderNickname
		temp.SenderFaceURL = v.SenderFaceURL
		temp.Content = v.Content
		temp.Seq = v.Seq
		temp.IsRead = v.IsRead
		temp.Status = v.Status
		temp.AttachedInfo = v.AttachedInfo
		temp.Ex = v.Ex
		err := c.msgHandleByContentType(&temp)
		if err != nil {
			log.Error(operationID, "Parsing data error:", err.Error(), temp)
			continue
		}
		if temp.ContentType == constant.File && !c.judgeMultipleSubString(searchParam.KeywordList, temp.FileElem.FileName, searchParam.KeywordListMatchType) {
			continue
		}
		if temp.ContentType == constant.AtText && !c.judgeMultipleSubString(searchParam.KeywordList, temp.AtElem.Text, searchParam.KeywordListMatchType) {
			continue
		}
		switch temp.SessionType {
		case constant.SingleChatType:
			if temp.SendID == c.loginUserID {
				conversationID = utils.GetConversationIDBySessionType(temp.RecvID, constant.SingleChatType)
			} else {
				conversationID = utils.GetConversationIDBySessionType(temp.SendID, constant.SingleChatType)
			}
		case constant.GroupChatType:
			temp.GroupID = temp.RecvID
			temp.RecvID = c.loginUserID
			conversationID = utils.GetConversationIDBySessionType(temp.GroupID, constant.GroupChatType)
		case constant.SuperGroupChatType:
			temp.GroupID = temp.RecvID
			temp.RecvID = c.loginUserID
			conversationID = utils.GetConversationIDBySessionType(temp.GroupID, constant.SuperGroupChatType)
		}
		if oldItem, ok := conversationMap[conversationID]; !ok {
			searchResultItem := sdk.SearchByConversationResult{}
			localConversation, err := c.db.GetConversation(conversationID)
			if err != nil {
				log.Error(operationID, "get conversation err ", err.Error(), conversationID)
				continue
			}
			searchResultItem.ConversationID = conversationID
			searchResultItem.FaceURL = localConversation.FaceURL
			searchResultItem.ShowName = localConversation.ShowName
			searchResultItem.ConversationType = localConversation.ConversationType
			searchResultItem.MessageList = append(searchResultItem.MessageList, &temp)
			searchResultItem.MessageCount++
			conversationMap[conversationID] = &searchResultItem
		} else {
			oldItem.MessageCount++
			oldItem.MessageList = append(oldItem.MessageList, &temp)
			conversationMap[conversationID] = oldItem
		}
	}
	for _, v := range conversationMap {
		r.SearchResultItems = append(r.SearchResultItems, v)
		r.TotalCount += v.MessageCount

	}
	return r
}

func (c *Conversation) setConversationNotification(msg *server_api_params.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	c.SyncConversations(operationID)
}

func (c *Conversation) DoNotification(msg *server_api_params.MsgData) {
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg)
	if c.msgListener == nil {
		log.Error(operationID, utils.GetSelfFuncName(), "listener == nil")
		return
	}
	go func() {
		c.setConversationNotification(msg, operationID)
	}()
}

func (c *Conversation) delMsgBySeq(seqList []uint32) error {
	var SPLIT = 1000
	for i := 0; i < len(seqList)/SPLIT; i++ {
		if err := c.delMsgBySeqSplit(seqList[i*SPLIT : (i+1)*SPLIT]); err != nil {
			return utils.Wrap(err, "")
		}
	}
	return nil
}

func (c *Conversation) delMsgBySeqSplit(seqList []uint32) error {
	var req server_api_params.DelMsgListReq
	req.SeqList = seqList
	req.OperationID = utils.OperationIDGenerator()
	req.OpUserID = c.loginUserID
	req.UserID = c.loginUserID
	operationID := req.OperationID

	resp, err := c.Ws.SendReqWaitResp(&req, constant.WsDelMsg, 30, 5, c.loginUserID, req.OperationID)
	if err != nil {
		return utils.Wrap(err, "SendReqWaitResp failed")
	}
	var delResp server_api_params.DelMsgListResp
	err = proto.Unmarshal(resp.Data, &delResp)
	if err != nil {
		log.Error(operationID, "Unmarshal failed ", err.Error())
		return utils.Wrap(err, "Unmarshal failed")
	}
	return nil
}

// old WS method
//func (c *Conversation) deleteMessageFromSvr(callback open_im_sdk_callback.Base, s *sdk_struct.MsgStruct, operationID string) {
//	seq, err := c.db.GetMsgSeqByClientMsgID(s.ClientMsgID)
//	common.CheckDBErrCallback(callback, err, operationID)
//	if seq == 0 {
//		err = errors.New("seq == 0 ")
//		common.CheckArgsErrCallback(callback, err, operationID)
//	}
//	seqList := []uint32{seq}
//	err = c.delMsgBySeq(seqList)
//	common.CheckArgsErrCallback(callback, err, operationID)
//}

func (c *Conversation) deleteConversationAndMsgFromSvr(callback open_im_sdk_callback.Base, conversationID, operationID string) {
	local, err := c.db.GetConversation(conversationID)
	common.CheckDBErrCallback(callback, err, operationID)
	log.Debug(operationID, utils.GetSelfFuncName(), *local)
	var seqList []uint32
	switch local.ConversationType {
	case constant.SingleChatType:
		peerUserID := local.UserID
		if peerUserID != c.loginUserID {
			seqList, err = c.db.GetMsgSeqListByPeerUserID(peerUserID)
		} else {
			seqList, err = c.db.GetMsgSeqListBySelfUserID(c.loginUserID)
		}
		log.NewDebug(operationID, utils.GetSelfFuncName(), "seqList: ", seqList)
		common.CheckDBErrCallback(callback, err, operationID)
	case constant.GroupChatType:
		groupID := local.GroupID
		seqList, err = c.db.GetMsgSeqListByGroupID(groupID)
		log.NewDebug(operationID, utils.GetSelfFuncName(), "seqList: ", seqList)
		common.CheckDBErrCallback(callback, err, operationID)
	case constant.SuperGroupChatType:
		var apiReq server_api_params.DelSuperGroupMsgReq
		apiReq.UserID = c.loginUserID
		apiReq.IsAllDelete = true
		apiReq.GroupID = local.GroupID
		apiReq.OperationID = operationID
		c.p.PostFatalCallback(callback, constant.DeleteSuperGroupMsgRouter, apiReq, nil, apiReq.OperationID)
		return

	}
	var apiReq server_api_params.DeleteMsgReq
	apiReq.OpUserID = c.loginUserID
	apiReq.UserID = c.loginUserID
	apiReq.OperationID = operationID
	apiReq.SeqList = seqList
	c.p.PostFatalCallback(callback, constant.DeleteMsgRouter, apiReq, nil, apiReq.OperationID)
	common.CheckArgsErrCallback(callback, err, operationID)
}

func (c *Conversation) deleteAllMsgFromLocal(callback open_im_sdk_callback.Base, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName())
	err := c.db.DeleteAllMessage()
	common.CheckDBErrCallback(callback, err, operationID)
	err = c.db.CleaAllConversation()
	common.CheckDBErrCallback(callback, err, operationID)
	conversationList, err := c.db.GetAllConversationList()
	common.CheckDBErrCallback(callback, err, operationID)
	var cidList []string
	for _, conversation := range conversationList {
		cidList = append(cidList, conversation.ConversationID)
	}
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{Action: constant.ConChange, Args: cidList}, c.GetCh())
	c.doUpdateConversation(common.Cmd2Value{Value: common.UpdateConNode{"", constant.TotalUnreadMessageChanged, ""}})

}

func (c *Conversation) deleteAllMsgFromSvr(callback open_im_sdk_callback.Base, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName())
	seqList, err := c.db.GetAllUnDeleteMessageSeqList()
	log.NewInfo(operationID, utils.GetSelfFuncName(), seqList)
	common.CheckDBErrCallback(callback, err, operationID)
	var apiReq server_api_params.DeleteMsgReq
	apiReq.OpUserID = c.loginUserID
	apiReq.UserID = c.loginUserID
	apiReq.OperationID = operationID
	apiReq.SeqList = seqList
	c.p.PostFatalCallback(callback, constant.DeleteMsgRouter, apiReq, nil, apiReq.OperationID)
}
