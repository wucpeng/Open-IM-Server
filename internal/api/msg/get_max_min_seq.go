package msg

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	commonDB "Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/msg"
	sdk_ws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	//"bytes"
	"context"
	"github.com/gin-gonic/gin"
	//"io/ioutil"
	"net/http"
	"strings"
)

type paramsUserNewestSeq struct {
	ReqIdentifier int    `json:"reqIdentifier" binding:"required"`
	SendID        string `json:"sendID" binding:"required"`
	OperationID   string `json:"operationID" binding:"required"`
	MsgIncr       int    `json:"msgIncr" binding:"required"`
}

func GetSeq(c *gin.Context) {
	params := paramsUserNewestSeq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	token := c.Request.Header.Get("token")
	if ok, err := token_verify.VerifyToken(token, params.SendID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	pbData := sdk_ws.GetMaxAndMinSeqReq{}
	pbData.UserID = params.SendID
	pbData.OperationID = params.OperationID
	grpcConn := getcdv3.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMsgName, pbData.OperationID)
	if grpcConn == nil {
		errMsg := pbData.OperationID + " getcdv3.GetDefaultConn == nil"
		log.NewError(pbData.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	msgClient := pbChat.NewMsgClient(grpcConn)
	reply, err := msgClient.GetMaxAndMinSeq(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserGetSeq rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserGetSeq rpc failed, " + err.Error()})
		return
	}
	log.NewError(params.OperationID, "GetSeq result, ", params, reply)
	c.JSON(http.StatusOK, gin.H{
		"errCode":       reply.ErrCode,
		"errMsg":        reply.ErrMsg,
		"msgIncr":       params.MsgIncr,
		"reqIdentifier": params.ReqIdentifier,
		"data": gin.H{
			"maxSeq": reply.MaxSeq,
			"minSeq": reply.MinSeq,
		},
	})

}

type GroupCalendarReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	StartTime   int64  `json:"startTime"  binding:"required"`
	EndTime     int64  `json:"endTime"  binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
}

func GetGroupCalendar(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))

	params := GroupCalendarReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	if params.GroupID != "" {
		messages, err := commonDB.DB.GetGroupAllMsgList(params.UserID, params.GroupID, params.StartTime, params.EndTime, params.OperationID)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
		var times []int64
		for _, md := range messages {
			times = append(times, md.SendTime)
		}
		c.JSON(http.StatusOK, gin.H{
			"errCode": 0,
			"errMsg":  "",
			"data":    times,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"errCode": 0,
			"errMsg":  "",
			"data":    messages,
		})
	}
}

func reverseSlice(s []*sdk_ws.MsgData) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

type GroupForwardReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Forward     int64  `json:"forward"`
	TimeDesc    int64  `json:"timeDesc"`
	LastMsgId   uint32 `json:"lastMsgId"`
	AssignId    string `json:"assignId"`
	PageSize    int    `json:"pageSize"`
}

func GetGroupForward(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))
	params := GroupForwardReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	messages, err := commonDB.DB.GetGroupAllMsgList(params.UserID, params.GroupID, params.StartTime, params.EndTime, params.OperationID)
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}

	if params.AssignId != "" {
		var messages2 []*sdk_ws.MsgData
		for _, md := range messages {
			if md.SendID == params.AssignId {
				messages2 = append(messages2, md)
			}
		}
		messages = messages2
	}
	if len(messages) < 1 {
		c.JSON(http.StatusOK, gin.H{
			"errCode": 0,
			"errMsg":  "",
			"data":    messages,
		})
		return
	}
	if params.TimeDesc == 1 { //倒序
		reverseSlice(messages)
	}
	if params.LastMsgId != 0 { //数据
		i := -1
		for s, md := range messages {
			if md.Seq == params.LastMsgId {
				i = s
				break
			}
		}
		log.NewInfo(params.OperationID, utils.GetSelfFuncName(), i, len(messages))
		if i != -1 {
			if params.Forward == 1 { //下一页
				len3 := i + params.PageSize + 1
				if len3 > len(messages) {
					len3 = len(messages)
				}
				messages = messages[i+1 : len3]
			} else { //上一页
				//messages = messages[:i]
				if i > params.PageSize {
					messages = messages[i-params.PageSize : i]
				} else {
					messages = messages[:i]
				}
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"errCode": 10001,
				"errMsg":  "",
				"data":    nil,
			})
			return
		}
	} else {
		log.NewInfo(params.OperationID, utils.GetSelfFuncName(), params.PageSize, len(messages))
		if params.Forward == 1 { //下一页
			len1 := len(messages)
			if len1 > params.PageSize {
				len1 = params.PageSize
			}
			messages = messages[:len1]
		} else { //上一页
			lenM := len(messages)
			if lenM > params.PageSize {
				messages = messages[lenM-params.PageSize : lenM]
			} else {
				messages = messages[:lenM]
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": 0,
		"errMsg":  "",
		"data":    messages,
	})
}

type GroupRangeReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	TimeDesc    int64  `json:"timeDesc"`
	LastMsgId   uint32 `json:"lastMsgId"`
	PageSize    int    `json:"pageSize"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

func GetGroupRange(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))
	params := GroupRangeReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	messages, err := commonDB.DB.GetGroupAllMsgList(params.UserID, params.GroupID, params.StartTime, params.EndTime, params.OperationID)
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	if len(messages) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"errCode": 0,
			"errMsg":  "",
			"data":    messages,
		})
		return
	}

	if params.TimeDesc == 1 { //倒序
		reverseSlice(messages)
	}
	if params.LastMsgId != 0 { //数据
		i := -1
		for s, md := range messages {
			if md.Seq == params.LastMsgId {
				i = s
				break
			}
		}
		//log.NewInfo(params.OperationID, utils.GetSelfFuncName(), i, len(messages))
		if i != -1 {
			len3 := i + params.PageSize
			if len3 > len(messages) {
				len3 = len(messages)
			}
			len1 := i - params.PageSize
			if len1 < 0 {
				len1 = 0
			}
			messages = messages[len1:len3]
		} else {
			c.JSON(http.StatusOK, gin.H{
				"errCode": 10001,
				"errMsg":  "",
				"data":    nil,
			})
			return
		}
	} else {
		lenM := len(messages)
		if lenM > params.PageSize {
			messages = messages[:params.PageSize]
		} else {
			messages = messages[:lenM]
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": 0,
		"errMsg":  "",
		"data":    messages,
	})
}

type UserCalendarReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	StartTime   int64  `json:"startTime"  binding:"required"`
	EndTime     int64  `json:"endTime"  binding:"required"`
	SenderID    string `json:"senderID"  binding:"required"`
}

func GetUserCalendar(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))
	params := UserCalendarReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	messages, err := commonDB.DB.GetSingleAllMsgList(params.UserID, params.SenderID, params.StartTime, params.EndTime, params.OperationID)
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	var times []int64
	for _, md := range messages {
		times = append(times, md.SendTime)
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": 0,
		"errMsg":  "",
		"data":    times,
	})
}

type UserForwardReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	SenderID    string `json:"senderID" binding:"required"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Forward     int64  `json:"forward"`
	TimeDesc    int64  `json:"timeDesc"`
	LastMsgId   uint32 `json:"lastMsgId"`
	PageSize    int    `json:"pageSize"`
}

func GetUserForward(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))
	params := UserForwardReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	messages, err := commonDB.DB.GetSingleAllMsgList(params.UserID, params.SenderID, params.StartTime, params.EndTime, params.OperationID)
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	if len(messages) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"errCode": 0,
			"errMsg":  "",
			"data":    messages,
		})
		return
	}
	if params.TimeDesc == 1 { //倒序
		reverseSlice(messages)
	}
	//log.NewInfo(params.OperationID, utils.GetSelfFuncName(), len(messages), params.LastMsgId)
	if params.LastMsgId != 0 { //数据
		i := -1
		for s, md := range messages {
			if md.Seq == params.LastMsgId {
				i = s
				break
			}
		}
		//log.NewInfo(params.OperationID, utils.GetSelfFuncName(), i, len(messages), params.LastMsgId)
		if i != -1 {
			if params.Forward == 1 { //下一页
				len3 := i + params.PageSize + 1
				if len3 > len(messages) {
					len3 = len(messages)
				}
				messages = messages[i+1 : len3]
				//log.NewInfo(params.OperationID, utils.GetSelfFuncName(), i, len3)
			} else { //上一页
				//messages = messages[:i]
				if i > params.PageSize {
					messages = messages[i-params.PageSize : i]
				} else {
					messages = messages[:i]
				}
				//log.NewInfo(params.OperationID, utils.GetSelfFuncName(), i)
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"errCode": 10001,
				"errMsg":  "",
				"data":    nil,
			})
			return
		}
	} else {
		if params.Forward == 1 { //下一页
			len1 := len(messages)
			if len1 > params.PageSize {
				len1 = params.PageSize
			}
			messages = messages[:len1]
		} else { //上一页
			lenM := len(messages)
			if lenM > params.PageSize {
				messages = messages[lenM-params.PageSize : lenM]
			} else {
				messages = messages[:lenM]
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": 0,
		"errMsg":  "",
		"data":    messages,
	})
}

type GroupAtReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	AtMsgId     string `json:"atMsgId" binding:"required"`
	EndMsgId    string `json:"endMsgId" binding:"required"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

func GetGroupAt(c *gin.Context) {
	//bodyData, _ := ioutil.ReadAll(c.Request.Body)
	//log.NewInfo("2222", utils.GetSelfFuncName(), "Body", string(bodyData))
	//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))

	params := GroupAtReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	var messages []*sdk_ws.MsgData
	messages, err := commonDB.DB.GetGroupAllMsgList(params.UserID, params.GroupID, params.StartTime, params.EndTime, params.OperationID)
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "GetAllMsgList failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}
	resp := api.GroupAtResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}}
	reverseSlice(messages)
	var atMsgSite int
	var atMsgSeq uint32
	for i, md := range messages {
		if params.AtMsgId == md.ClientMsgID {
			atMsgSite = i
			atMsgSeq = md.Seq
			break
		}
	}
	var endMsgSite int
	var endMsgSeq uint32
	for i, md := range messages {
		if params.EndMsgId == md.ClientMsgID {
			endMsgSite = i
			endMsgSeq = md.Seq
			break
		}
	}
	log.NewInfo(params.OperationID, utils.GetSelfFuncName(), atMsgSite, atMsgSeq, endMsgSite, endMsgSeq)
	resp.AtMsgSite = atMsgSite
	resp.AtMsgSeq = atMsgSeq
	resp.EndMsgSite = endMsgSite
	resp.EndMsgSeq = endMsgSeq

	c.JSON(http.StatusOK, resp)
}

type CheckUserMsg struct {
	OperationID string `json:"operationID"  binding:"required"`
	UserID      string `json:"userID"  binding:"required"`
	Type        int32  `json:"type"  binding:"required"`
}

func CheckUserMongoMsg(c *gin.Context) {
	params := CheckUserMsg{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	//if ok, err := token_verify.VerifyToken(c.Request.Header.Get("token"), params.UserID); !ok {
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
	//	return
	//}
	log.NewError(params.OperationID, utils.GetSelfFuncName(), params.Type, params.UserID)
	if params.Type == 1 {
		_, err := commonDB.DB.UserMsgLogs(params.UserID, params.OperationID)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "UserMsgLogs failed", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
	} else if params.Type == 2 {
		_, err := commonDB.DB.ResetSystemMsgList(params.UserID, params.OperationID)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "ResetSystemMsgList failed", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
	} else if params.Type == 3 {
		_, err := commonDB.DB.UserGroupMsgLogs(params.UserID, params.OperationID, params.OperationID)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "ResetSystemMsgList failed", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": 0,
		"errMsg":  "",
	})
}
