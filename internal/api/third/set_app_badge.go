package apiThird

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

//func TokenCheck() *gin.HandlerFunc {
//	return func(c *gin.Context) {
//		//var ok bool
//		//var errInfo, opUserID string
//		//ok, opUserID, errInfo = token_verify.GetUserIDFromToken(c.Request.Header.Get("token"), "xxxxxxxxxxxx")
//		//if !ok {
//		//	errMsg := req.OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + c.Request.Header.Get("token")
//		//	log.NewError(req.OperationID, errMsg)
//		//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
//		//	return
//		//}
//		token := c.Request.Header.Get("token")
//		log.NewError(token, utils.GetSelfFuncName())
//		return
//	}
//}

func TokenCheck(c *gin.Context) {
	OperationID := utils.OperationIDGenerator()
	token := c.Request.Header.Get("token")
	log.NewError(OperationID, utils.GetSelfFuncName(), token)
	ok, opUserID, errInfo := token_verify.GetUserIDFromToken(token, OperationID)
	if !ok {
		errMsg := OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + token
		log.NewError(OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		c.Abort()
		return
	}
	c.Set("opUserID", opUserID)
	c.Next()
}

func SetAppBadge(c *gin.Context) {
	var (
		req  api.SetAppBadgeReq
		resp api.SetAppBadgeResp
	)
	if err := c.Bind(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req)

	var ok bool
	var errInfo, opUserID string
	ok, opUserID, errInfo = token_verify.GetUserIDFromToken(c.Request.Header.Get("token"), req.OperationID)
	if !ok {
		errMsg := req.OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + c.Request.Header.Get("token")
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	if !token_verify.CheckAccess(opUserID, req.FromUserID) {
		log.NewError(req.OperationID, "CheckAccess false ", opUserID, req.FromUserID)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "no permission"})
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req, opUserID)
	err := db.DB.SetUserBadgeUnreadCountSum(req.FromUserID, int(req.AppUnreadCount))
	if err != nil {
		errMsg := req.OperationID + " " + "SetUserBadgeUnreadCountSum failed " + err.Error() + " token:" + c.Request.Header.Get("token")
		log.NewError(req.OperationID, errMsg)
		resp.ErrCode = 500
		resp.ErrMsg = errMsg
		c.JSON(http.StatusInternalServerError, resp)
		return
	}
	c.JSON(http.StatusOK, resp)
	return
}
