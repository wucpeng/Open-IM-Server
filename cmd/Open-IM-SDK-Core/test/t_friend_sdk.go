package test

import (
	"encoding/json"
	"fmt"
	"open_im_sdk/internal/login"
	"open_im_sdk/pkg/server_api_params"

	//"gorm.io/gorm/callbacks"
	X "log"
	"open_im_sdk/open_im_sdk"
	"open_im_sdk/pkg/log"
	"open_im_sdk/pkg/sdk_params_callback"
	"open_im_sdk/pkg/utils"
	"open_im_sdk/sdk_struct"
	"os"
	"runtime"
	"time"
)

var loggerf *X.Logger

func init() {
	loggerf = X.New(os.Stdout, "", X.Llongfile|X.Ltime|X.Ldate)
}

type TestSendImg struct {
}

func (TestSendImg) OnSuccess(data string) {
	fmt.Println("testSendImg, OnSuccess, output: ", data)
}

func (TestSendImg) OnError(code int, msg string) {
	fmt.Println("testSendImg, OnError, ", code, msg)
}

func (TestSendImg) OnProgress(progress int) {
	fmt.Println("progress: ", progress)
}

func TestLog(v ...interface{}) {
	//X.SetFlags(X.Lshortfile | X.LstdFlags)
	loggerf.Println(v)
	a, b, c, d := runtime.Caller(1)
	X.Println(a, b, c, d)
}

var Friend_uid = "13911112222"

func SetTestFriendID(friendUserID string) {
	Friend_uid = friendUserID
}

///////////////////////////////////////////////////////////

type testGetFriendApplicationList struct {
	baseCallback
}

func DoTestGetFriendApplicationList() {
	var test testGetFriendApplicationList
	test.OperationID = utils.OperationIDGenerator()
	log.Info(test.OperationID, utils.GetSelfFuncName(), "input ")
	open_im_sdk.GetRecvFriendApplicationList(test, test.OperationID)

}

//////////////////////////////////////////////////////////`
type testSetSelfInfo struct {
	baseCallback
}

func DoTestSetSelfInfo() {
	var test testSetSelfInfo
	test.OperationID = utils.OperationIDGenerator()
	userInfo := server_api_params.UserInfo{}
	userInfo.Nickname = "new 4444444444444 Gordon001"
	jsonString := utils.StructToJsonStringDefault(userInfo)
	fmt.Println("SetSelfInfo, input: ")
	open_im_sdk.SetSelfInfo(test, test.OperationID, jsonString)
}

/////////////////////////////////////////////////////////
type testGetUsersInfo struct {
	baseCallback
}

func DoTestGetUsersInfo() {
	var test testGetUsersInfo
	test.OperationID = utils.OperationIDGenerator()
	userIDList := []string{"openIM2222"}
	list := utils.StructToJsonStringDefault(userIDList)
	fmt.Println("testGetUsersInfo, input: ", list)
	open_im_sdk.GetUsersInfo(test, test.OperationID, list)
}

/////////////////////////////////////////////////////////
type testGetFriendsInfo struct {
	uid []string //`json:"uidList"`
}

func (testGetFriendsInfo) OnSuccess(data string) {
	fmt.Println("DoTestGetDesignatedFriendsInfo, OnSuccess, output: ", data)
}

func (testGetFriendsInfo) OnError(code int32, msg string) {
	fmt.Println("DoTestGetDesignatedFriendsInfo, OnError, ", code, msg)
}

func DoTestGetDesignatedFriendsInfo() {
	var test testGetFriendsInfo
	test.uid = append(test.uid, Friend_uid)

	jsontest, _ := json.Marshal(test.uid)
	fmt.Println("testGetFriendsInfo, input: ", string(jsontest))
	open_im_sdk.GetDesignatedFriendsInfo(test, "xxxxxxxxxxx", string(jsontest))
}

///////////////////////////////////////////////////////
//
type testAddToBlackList struct {
	OperationID string
}

func (t testAddToBlackList) OnSuccess(string) {
	log.Info(t.OperationID, "testAddToBlackList, OnSuccess")
}

func (t testAddToBlackList) OnError(code int32, msg string) {
	log.Info(t.OperationID, "testAddToBlackList, OnError, ", code, msg)
}

func DoTestAddToBlackList() {
	var test testAddToBlackList
	test.OperationID = utils.OperationIDGenerator()
	fmt.Println("AddToBlackList, input: ", Friend_uid)
	open_im_sdk.AddBlack(test, test.OperationID, Friend_uid)
}

///////////////////////////////////////////////////////
type testDeleteFromBlackList struct {
	baseCallback
}

func DoTestDeleteFromBlackList() {
	var test testDeleteFromBlackList
	test.OperationID = utils.OperationIDGenerator()
	fmt.Println("DeleteFromBlackList, input: ")
	open_im_sdk.RemoveBlack(test, test.OperationID, Friend_uid)
}

//////////////////////////////////////////////////////
type testGetBlackList struct {
	OperationID string
}

func (t testGetBlackList) OnSuccess(data string) {
	log.Info(t.OperationID, "testGetBlackList, OnSuccess, output: ", data)
}
func (t testGetBlackList) OnError(code int32, msg string) {
	log.Info(t.OperationID, "testGetBlackList, OnError, ", code, msg)
}
func DoTestGetBlackList() {
	var test testGetBlackList
	test.OperationID = utils.OperationIDGenerator()
	open_im_sdk.GetBlackList(test, test.OperationID)
}

//////////////////////////////////////////////////////

type testCheckFriend struct {
	OperationID string
}

func (t testCheckFriend) OnSuccess(data string) {
	log.Info(t.OperationID, "testCheckFriend, OnSuccess, output: ", data)
}
func (t testCheckFriend) OnError(code int32, msg string) {
	log.Info(t.OperationID, "testCheckFriend, OnError, ", code, msg)
}
func DoTestCheckFriend() {
	var test testCheckFriend
	test.OperationID = utils.OperationIDGenerator()
	userIDList := []string{"openIM100"}
	list := utils.StructToJsonString(userIDList)
	fmt.Println("CheckFriend, input: ", list)
	open_im_sdk.CheckFriend(test, test.OperationID, list)
}

///////////////////////////////////////////////////////////
type testSetFriendRemark struct {
	baseCallback
}

func DotestSetFriendRemark() {
	var test testSetFriendRemark
	test.OperationID = utils.OperationIDGenerator()

	var param sdk_params_callback.SetFriendRemarkParams
	param.ToUserID = Friend_uid
	param.Remark = "ttttttttttttttttt "
	jsontest := utils.StructToJsonString(param)
	log.Info(test.OperationID, utils.GetSelfFuncName(), "input ", jsontest)
	open_im_sdk.SetFriendRemark(test, test.OperationID, jsontest)
}

/////////////////////
////////////////////////////////////////////////////////

type testDeleteFriend struct {
	baseCallback
}

func DotestDeleteFriend() {
	var test testDeleteFriend
	test.OperationID = utils.OperationIDGenerator()

	log.Info(test.OperationID, utils.GetSelfFuncName(), "input ")
	open_im_sdk.DeleteFriend(test, test.OperationID, Friend_uid)
}

///////////////////////////////////////////////////////
/////////////////////////////////////////////////////////
type testaddFriend struct {
	OperationID string
}

func (t testaddFriend) OnSuccess(data string) {
	log.Info(t.OperationID, "testaddFriend, OnSuccess", data)
}
func (t testaddFriend) OnError(code int32, msg string) {
	log.Info(t.OperationID, "testaddFriend, OnError", code, msg)
}

func DoTestAddFriend() {
	var test testaddFriend
	test.OperationID = utils.OperationIDGenerator()
	params := sdk_params_callback.AddFriendParams{
		ToUserID: Friend_uid,
		ReqMsg:   "777777777777777777777777",
	}
	jsontestaddFriend := utils.StructToJsonString(params)
	log.Info(test.OperationID, "addFriend input:", jsontestaddFriend)
	open_im_sdk.AddFriend(test, test.OperationID, jsontestaddFriend)
}

////////////////////////////////////////////////////////////////////

/////////////////////////////////////////////////////////
type testGetSendFriendApplicationList struct {
	OperationID string
}

func (t testGetSendFriendApplicationList) OnSuccess(data string) {
	log.Info(t.OperationID, "testGetSendFriendApplicationList, OnSuccess", data)
}
func (t testGetSendFriendApplicationList) OnError(code int32, msg string) {
	log.Info(t.OperationID, "testGetSendFriendApplicationList, OnError", code, msg)
}

func DoTestGetSendFriendApplicationList() {
	var test testGetSendFriendApplicationList
	test.OperationID = utils.OperationIDGenerator()
	log.Info(test.OperationID, "GetSendFriendApplicationList input:")
	open_im_sdk.GetSendFriendApplicationList(test, test.OperationID)
}

////////////////////////////////////////////////////////////////////

type testGetFriendList struct {
	baseCallback
}

func DotestGetFriendList() {
	var test testGetFriendList
	test.OperationID = utils.OperationIDGenerator()
	log.Info(test.OperationID, utils.GetSelfFuncName(), "input ")
	open_im_sdk.GetFriendList(test, test.OperationID)
}

type testSearchFriends struct {
	baseCallback
}

func DotestSearchFriends() {
	var test testSearchFriends
	test.OperationID = utils.OperationIDGenerator()
	test.callName = "SearchFriends"
	var params sdk_params_callback.SearchFriendsParam
	params.KeywordList = []string{"G"}
	params.IsSearchUserID = true
	params.IsSearchNickname = true
	params.IsSearchRemark = true
	log.Info(test.OperationID, utils.GetSelfFuncName(), "input ", params)
	open_im_sdk.SearchFriends(test, test.OperationID, utils.StructToJsonString(params))
}

/////////////////////////////////////////////////////////////////////

type testAcceptFriendApplication struct {
	baseCallback
}

func DoTestAcceptFriendApplication() {
	var test testAcceptFriendApplication
	test.OperationID = utils.OperationIDGenerator()
	var param sdk_params_callback.ProcessFriendApplicationParams
	param.HandleMsg = "ok ok "
	param.ToUserID = Friend_uid
	input := utils.StructToJsonString(param)
	open_im_sdk.AcceptFriendApplication(test, test.OperationID, input)
}

type testRefuseFriendApplication struct {
	baseCallback
}

func DoTestRefuseFriendApplication() {
	var test testRefuseFriendApplication
	test.OperationID = utils.OperationIDGenerator()
	var param sdk_params_callback.ProcessFriendApplicationParams
	param.HandleMsg = "nonono"
	param.ToUserID = Friend_uid
	input := utils.StructToJsonString(param)
	open_im_sdk.RefuseFriendApplication(test, test.OperationID, input)
}

/*
type testRefuseFriendApplication struct {
	ui2AcceptFriend
}

func (testRefuseFriendApplication) OnSuccess(info string) {
	fmt.Println("RefuseFriendApplication OnSuccess", info)
}
func (testRefuseFriendApplication) OnError(code int, msg string) {
	fmt.Println("RefuseFriendApplication, OnError, ", code, msg)
}
*/
/*
func DoTestRefuseFriendApplication() {

	var test testRefuseFriendApplication
	test.UID = Friend_uid

	js, _ := json.Marshal(test)
	RefuseFriendApplication(test, string(js))
	fmt.Println("RefuseFriendApplication, input: ", string(js))
}


*/

/////////////////////////////////////////////////////////////////////

//type testRefuseFriendApplication struct {
//	open_im_sdk.ui2AcceptFriend
//}
//
//func (testRefuseFriendApplication) OnSuccess(info string) {
//	fmt.Println("testRefuseFriendApplication OnSuccess", info)
//}
//func (testRefuseFriendApplication) OnError(code int32, msg string) {
//	fmt.Println("testRefuseFriendApplication, OnError, ", code, msg)
//}
//func DoTestRefuseFriendApplication() {
//	var testRefuseFriendApplication testRefuseFriendApplication
//	testRefuseFriendApplication.ui2AcceptFriend = Friend_uid
//
//	jsontestfusetFriendappclicatrion, _ := json.Marshal(testRefuseFriendApplication.UID)
//	open_im_sdk.RefuseFriendApplication(testRefuseFriendApplication, string(jsontestfusetFriendappclicatrion), "")
//	fmt.Println("RefuseFriendApplication, input: ", string(jsontestfusetFriendappclicatrion))
//}

////////////////////////////////////////////////////////////////////

type BaseSuccessFailed struct {
	successData string
	errCode     int
	errMsg      string
	funcName    string
	time        time.Time
}

func (b *BaseSuccessFailed) OnError(errCode int32, errMsg string) {
	b.errCode = -1
	b.errMsg = errMsg
	log.Error("login failed", errCode, errMsg)

}

func (b *BaseSuccessFailed) OnSuccess(data string) {
	b.errCode = 1
	b.successData = data
	log.Info("login success", data, time.Since(b.time))
}

func InOutlllogin(uid, tk string) {
	var callback BaseSuccessFailed
	callback.time = time.Now()
	callback.funcName = utils.GetSelfFuncName()
	operationID := utils.OperationIDGenerator()
	open_im_sdk.Login(&callback, operationID, uid, tk)
	for {
		if callback.errCode == 1 {
			return
		} else if callback.errCode == -1 {
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(100 * time.Millisecond)
			log.Info(operationID, "waiting login ")
		}
	}
}

func InOutLogou() {
	var callback BaseSuccessFailed
	callback.funcName = utils.GetSelfFuncName()
	opretaionID := utils.OperationIDGenerator()
	open_im_sdk.Logout(&callback, opretaionID)
}

func InOutDoTest(uid, tk, ws, api string) {
	var cf sdk_struct.IMConfig
	cf.ApiAddr = api
	cf.Platform = 1
	cf.WsAddr = ws
	cf.DataDir = "./"
	cf.LogLevel = 6
	cf.ObjectStorage = "minio"

	var s string
	b, _ := json.Marshal(cf)
	s = string(b)
	fmt.Println(s)
	var testinit testInitLister

	operationID := utils.OperationIDGenerator()
	if !open_im_sdk.InitSDK(&testinit, operationID, s) {
		log.Error("", "InitSDK failed")
		return
	}

	var testConversation conversationCallBack
	open_im_sdk.SetConversationListener(testConversation)

	var testUser userCallback
	open_im_sdk.SetUserListener(testUser)

	var msgCallBack MsgListenerCallBak
	open_im_sdk.SetAdvancedMsgListener(&msgCallBack)

	var batchMsg BatchMsg
	open_im_sdk.SetBatchMsgListener(&batchMsg)

	var friendListener testFriendListener
	open_im_sdk.SetFriendListener(friendListener)

	var groupListener testGroupListener
	open_im_sdk.SetGroupListener(groupListener)
	var signalingListener testSignalingListener
	open_im_sdk.SetSignalingListener(&signalingListener)

	var organizationListener testOrganizationListener
	open_im_sdk.SetOrganizationListener(organizationListener)

	var workMomentsListener testWorkMomentsListener
	open_im_sdk.SetWorkMomentsListener(workMomentsListener)

	InOutlllogin(uid, tk)

	log.Warn("", "InOutDoTest fin")
}

func lllogin(uid, tk string) bool {
	var callback BaseSuccessFailed
	callback.funcName = utils.GetSelfFuncName()
	operationID := utils.OperationIDGenerator()
	open_im_sdk.Login(&callback, uid, operationID, tk)

	for true {
		if callback.errCode == 1 {
			fmt.Println("success code 1")
			return true
		} else if callback.errCode == -1 {
			fmt.Println("failed code -1")
			return false
		} else {
			fmt.Println("code sleep")
			time.Sleep(1 * time.Second)
			continue
		}
	}
	return true
}

func ReliabilityInitAndLogin(index int, uid, tk, ws, api string) {
	coreMgrLock.Lock()
	defer coreMgrLock.Unlock()

	var cf sdk_struct.IMConfig
	cf.ApiAddr = api
	cf.WsAddr = ws
	cf.Platform = 1
	cf.DataDir = "./"
	cf.LogLevel = 3
	log.Info("", "DoReliabilityTest", uid, tk, ws, api)

	operationID := utils.OperationIDGenerator()
	var testinit testInitLister
	lg := new(login.LoginMgr)
	log.Info(operationID, "new login ", lg)

	sdk_struct.SvrConf = cf
	allLoginMgr[index].mgr = lg
	lg.InitSDK(sdk_struct.SvrConf, &testinit, operationID)

	log.Info(operationID, "InitSDK ", sdk_struct.SvrConf)

	var testConversation conversationCallBack
	lg.SetConversationListener(testConversation)

	var testUser userCallback
	lg.SetUserListener(testUser)

	var msgCallBack MsgListenerCallBak
	lg.SetAdvancedMsgListener(&msgCallBack)

	var friendListener testFriendListener
	lg.SetFriendListener(friendListener)

	var groupListener testGroupListener
	lg.SetGroupListener(groupListener)

	var organizationListener testOrganizationListener
	lg.SetOrganizationListener(organizationListener)

	var callback BaseSuccessFailed
	callback.funcName = utils.GetSelfFuncName()
	lg.Login(&callback, uid, tk, operationID)

	for {
		if callback.errCode == 1 {
			log.Info(operationID, "login ok ", uid)
			return
		}
		log.Warn(operationID, "waiting login...", uid)
		time.Sleep(100 * time.Millisecond)
	}

}

func DoTest(uid, tk, ws, api string) {
	var cf sdk_struct.IMConfig
	cf.ApiAddr = api // "http://120.24.45.199:10000"
	//	cf.IpWsAddr = "wss://open-im.rentsoft.cn/wss"
	cf.WsAddr = ws //"ws://120.24.45.199:17778"
	cf.Platform = 1
	cf.DataDir = "./"

	var s string
	b, _ := json.Marshal(cf)
	s = string(b)
	fmt.Println(s)
	var testinit testInitLister
	operationID := utils.OperationIDGenerator()
	if !open_im_sdk.InitSDK(&testinit, operationID, s) {
		log.Error("", "InitSDK failed")
		return
	}

	var testConversation conversationCallBack
	open_im_sdk.SetConversationListener(testConversation)

	var testUser userCallback
	open_im_sdk.SetUserListener(testUser)

	//var msgCallBack MsgListenerCallBak
	//open_im_sdk.AddAdvancedMsgListener(msgCallBack)

	var friendListener testFriendListener
	open_im_sdk.SetFriendListener(friendListener)

	var groupListener testGroupListener
	open_im_sdk.SetGroupListener(groupListener)

	var signalingListener testSignalingListener
	open_im_sdk.SetSignalingListener(&signalingListener)
	time.Sleep(1 * time.Second)

	for !lllogin(uid, tk) {
		fmt.Println("lllogin, failed, login...")
		time.Sleep(1 * time.Second)
	}

}

////////////////////////////////////////////////////////////////////

type TestSendMsgCallBack struct {
	msg         string
	OperationID string
	sendID      string
	recvID      string
	msgID       string
}

func (t *TestSendMsgCallBack) OnError(errCode int32, errMsg string) {
	log.Error(t.OperationID, "test_openim: send msg failed: ", errCode, errMsg, t.msgID, t.msg)
	SendMsgMapLock.Lock()
	defer SendMsgMapLock.Unlock()
	SendFailedAllMsg[t.msgID] = t.sendID + t.recvID

}

func (t *TestSendMsgCallBack) OnSuccess(data string) {
	log.Info(t.OperationID, "test_openim: send msg success: |", t.msgID, t.msg, data)
	SendMsgMapLock.Lock()
	defer SendMsgMapLock.Unlock()
	SendSuccAllMsg[t.msgID] = t.sendID + t.recvID
}

func (t *TestSendMsgCallBack) OnProgress(progress int) {
	//	fmt.Printf("msg_send , onProgress %d\n", progress)
}

type TestSendMsgCallBackPress struct {
	msg         string
	OperationID string
	sendID      string
	recvID      string
	msgID       string
}

func (t *TestSendMsgCallBackPress) OnError(errCode int32, errMsg string) {
	log.Warn(t.OperationID, "TestSendMsgCallBackPress: send msg failed: ", errCode, errMsg, t.msgID, t.msg)
}

func (t *TestSendMsgCallBackPress) OnSuccess(data string) {
	log.Info(t.OperationID, "TestSendMsgCallBackPress: send msg success: |", t.msgID, t.msg)
}

func (t *TestSendMsgCallBackPress) OnProgress(progress int) {
	//	fmt.Printf("msg_send , onProgress %d\n", progress)
}

type BaseSuccessFailedTest struct {
	successData string
	errCode     int
	errMsg      string
	funcName    string
}

func (b *BaseSuccessFailedTest) OnError(errCode int32, errMsg string) {
	b.errCode = -1
	b.errMsg = errMsg
	fmt.Println("22onError ", b.funcName, errCode, errMsg)
}

func (b *BaseSuccessFailedTest) OnSuccess(data string) {
	b.errCode = 1
	b.successData = data
	fmt.Println("22OnSuccess: ", b.funcName, data)
}

func InOutDoTestSendMsg(sendId, receiverID string) {
	m := "test:" + sendId + ":" + receiverID + ":"
	//s := CreateTextMessage(m)
	var testSendMsg TestSendMsgCallBack
	//	testSendMsg.msg = SendMessage(&testSendMsg, s, receiverID, "", false)
	fmt.Println("func send ", m, testSendMsg.msg)
	fmt.Println("test to recv : ", receiverID)
}

//func DoTestGetAllConversationList() {
//	var test TestGetAllConversationListCallBack
//	open_im_sdk.GetAllConversationList(test)
//}

type userCallback struct {
}

func (userCallback) OnSelfInfoUpdated(callbackData string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackData)
}

////////////////////////////////////////////////////////////////////
type testInitLister struct {
}

func (t *testInitLister) OnUserTokenExpired() {
	log.Info("", utils.GetSelfFuncName())
}
func (t *testInitLister) OnConnecting() {
	log.Info("", utils.GetSelfFuncName())
}

func (t *testInitLister) OnConnectSuccess() {
	log.Info("", utils.GetSelfFuncName())
}

func (t *testInitLister) OnConnectFailed(ErrCode int32, ErrMsg string) {
	log.Info("", utils.GetSelfFuncName(), ErrCode, ErrMsg)
}

func (t *testInitLister) OnKickedOffline() {
	log.Info("", utils.GetSelfFuncName())
}

func (t *testInitLister) OnSelfInfoUpdated(info string) {
	log.Info("", utils.GetSelfFuncName())
}

func (t *testInitLister) OnSuccess() {
	log.Info("", utils.GetSelfFuncName())
}

func (t *testInitLister) OnError(code int32, msg string) {
	log.Info("", utils.GetSelfFuncName(), code, msg)
}

type testLogin struct {
}

func (testLogin) OnSuccess(string) {
	fmt.Println("testLogin OnSuccess")
}

func (testLogin) OnError(code int32, msg string) {
	fmt.Println("testLogin, OnError", code, msg)
}

type testFriendListener struct {
	x int
}

func (testFriendListener) OnFriendApplicationAdded(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}
func (testFriendListener) OnFriendApplicationDeleted(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnFriendApplicationAccepted(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnFriendApplicationRejected(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnFriendAdded(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnFriendDeleted(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnBlackAdded(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}
func (testFriendListener) OnBlackDeleted(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnFriendInfoChanged(callbackInfo string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), callbackInfo)
}

func (testFriendListener) OnSuccess() {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName())
}

func (testFriendListener) OnError(code int32, msg string) {
	log.Info(utils.OperationIDGenerator(), utils.GetSelfFuncName(), code, msg)
}
