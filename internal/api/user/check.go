package user

import (
	"Open_IM/pkg/common/constant"
	commonDB "Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"net/http"
	"github.com/gin-gonic/gin"
)

type CheckTokenReq struct {
	OperationID string `json:"operationID"  binding:"required"`
	Token       string `json:"token"  binding:"required"`
}

func CheckToken(c *gin.Context) {
	params := CheckTokenReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	claims, err := token_verify.GetClaimFromToken(params.Token)
	if err != nil {
		log.NewError(params.OperationID, "GetClaimFromToken err", err.Error(), params.Token)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": err.Error()})
		return
	}

	claimses := []*token_verify.Claims{}
	m, err := commonDB.DB.GetTokenMapByUidPid(claims.UID, claims.Platform)
	if err == nil && m != nil {
		for k, v := range m {
			cl, err := token_verify.GetClaimFromToken(k)
			log.NewInfo(params.OperationID, "climsa", v)
			if err == nil {
				claimses = append(claimses, cl)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errCode":       0,
		"errMsg":        "",
		"data1": claims,
		"data2": m,
		"platformId": constant.PlatformNameToID(claims.Platform),
		"claimses": claimses,
	})
}
