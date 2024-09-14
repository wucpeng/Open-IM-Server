package clientInit

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbRelay "Open_IM/pkg/proto/relay"
	"Open_IM/pkg/utils"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetClientInitConfig(c *gin.Context) {
	var req api.SetClientInitConfigReq
	var resp api.SetClientInitConfigResp
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)
	err, _ := token_verify.ParseTokenGetUserID(c.Request.Header.Get("token"), req.OperationID)
	if err != nil {
		errMsg := "ParseTokenGetUserID failed " + err.Error() + c.Request.Header.Get("token")
		log.NewError(req.OperationID, errMsg, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	m := make(map[string]interface{})
	if req.DiscoverPageURL != nil {
		m["discover_page_url"] = *req.DiscoverPageURL
	}
	if len(m) > 0 {
		err := imdb.SetClientInitConfig(m)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetClientInitConfig(c *gin.Context) {
	var req api.GetClientInitConfigReq
	var resp api.GetClientInitConfigResp
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)
	err, _ := token_verify.ParseTokenGetUserID(c.Request.Header.Get("token"), req.OperationID)
	if err != nil {
		errMsg := "ParseTokenGetUserID failed " + err.Error() + c.Request.Header.Get("token")
		log.NewError(req.OperationID, errMsg, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	config, err := imdb.GetClientInitConfig()
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
			return
		}
	}
	resp.Data.DiscoverPageURL = config.DiscoverPageURL
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp ", resp)
	c.JSON(http.StatusOK, resp)

}

func GetClientPlatformIds(c *gin.Context) {
	var req api.GetClientPlatformIdsReq
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "UserID: ", req.UserID)
	ids := make([]int32, 0)
	grpcCons := getcdv3.GetDefaultGatewayConn4Unique(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), req.OperationID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "UserID: ", req.UserID, len(grpcCons))
	for _, v := range grpcCons {
		msgClient := pbRelay.NewRelayClient(v)
		reply, err := msgClient.GetUserOnlinePlatformIds(context.Background(), &pbRelay.OnlineUserReq{OperationID: req.OperationID, UserID: req.UserID})
		if err != nil {
			log.NewError("GetUserOnlinePlatformIds push data to client rpc err", req.OperationID, "err", err)
			continue
		}
		if reply != nil && reply.PlatformIDs != nil {
			log.NewInfo(req.OperationID, "online send result", reply.PlatformIDs)
			ids = append(ids, reply.PlatformIDs...)
		}
	}
	var resp api.GetClientPlatformIdsResp
	resp.Data.PlatformIDs = ids
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp ", resp)
	c.JSON(http.StatusOK, resp)
}
