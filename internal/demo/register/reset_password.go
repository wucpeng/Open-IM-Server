package register

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type resetPasswordRequest struct {
	VerificationCode string `json:"verificationCode" binding:"required"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNumber"`
	NewPassword      string `json:"newPassword" binding:"required"`
	AreaCode         string `json:"areaCode"`
	OperationID      string `json:"operationID"`
}

func ResetPassword(c *gin.Context) {
	var (
		req resetPasswordRequest
	)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	if req.Email != "" {
		account = req.Email
	} else {
		account = req.PhoneNumber
	}

	if (config.Config.Demo.UseSuperCode && req.VerificationCode != config.Config.Demo.SuperCode) || !config.Config.Demo.UseSuperCode {
		accountKey := req.AreaCode + account + "_" + constant.VerificationCodeForResetSuffix
		v, err := db.DB.GetAccountCode(accountKey)
		if err != nil || v != req.VerificationCode {
			log.NewError(req.OperationID, "password Verification code error", account, req.VerificationCode, v)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code error!"})
			return
		}
	}
	user, err := im_mysql_model.GetRegister(account, req.AreaCode, "")
	if err != nil || user.Account == "" {
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "get register error", err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "user not register!"})
		return
	}
	if err := im_mysql_model.ResetPassword(account, req.NewPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ResetPasswordFailed, "errMsg": "reset password failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "reset password success"})
}
