package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	kfk "Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	//"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbMsg "Open_IM/pkg/proto/msg"
	//pbPush "Open_IM/pkg/proto/push"
	"Open_IM/pkg/utils"
	//"context"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"hash/crc32"
	//"strings"
	"sync"
	"time"
)

type MsgChannelValue struct {
	aggregationID string //maybe userID or super groupID
	triggerID     string
	msgList       []*pbMsg.MsgDataToMQ
	lastSeq       uint64
}
type TriggerChannelValue struct {
	triggerID string
	cmsgList  []*sarama.ConsumerMessage
}
type fcb func(cMsg *sarama.ConsumerMessage, msgKey string, sess sarama.ConsumerGroupSession)
type Cmd2Value struct {
	Cmd   int
	Value interface{}
}
type OnlineHistoryRedisConsumerHandler struct {
	historyConsumerGroup *kfk.MConsumerGroup
	chArrays             [ChannelNum]chan Cmd2Value
	msgDistributionCh    chan Cmd2Value
}

func (och *OnlineHistoryRedisConsumerHandler) Init(cmdCh chan Cmd2Value) {
	och.msgDistributionCh = make(chan Cmd2Value) //no buffer channel
	go och.MessagesDistributionHandle()
	for i := 0; i < ChannelNum; i++ {
		och.chArrays[i] = make(chan Cmd2Value, 50)
		go och.Run(i)
	}
	och.historyConsumerGroup = kfk.NewMConsumerGroup(&kfk.MConsumerGroupConfig{KafkaVersion: sarama.V2_0_0_0,
		OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, []string{config.Config.Kafka.Ws2mschat.Topic},
		config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.ConsumerGroupID.MsgToRedis)

}
func (och *OnlineHistoryRedisConsumerHandler) Run(channelID int) {
	for {
		select {
		case cmd := <-och.chArrays[channelID]:
			switch cmd.Cmd {
			case AggregationMessages:
				msgChannelValue := cmd.Value.(MsgChannelValue)
				msgList := msgChannelValue.msgList
				triggerID := msgChannelValue.triggerID
				storageMsgList := make([]*pbMsg.MsgDataToMQ, 0, 80)
				notStoragePushMsgList := make([]*pbMsg.MsgDataToMQ, 0, 80)
				//log.Info(triggerID, "msg arrived channel", channelID, msgChannelValue.aggregationID, len(msgList))
				for _, v := range msgList {
					isHistory := utils.GetSwitchFromOptions(v.MsgData.Options, constant.IsHistory)
					isSenderSync := utils.GetSwitchFromOptions(v.MsgData.Options, constant.IsSenderSync)
					//log.Info(triggerID, "OnlineHistoryRedisConsumerHandler", v.MsgData.ContentType, v.MsgData.SessionType, isHistory, isSenderSync, v.MsgData.Options)
					if isHistory {
						storageMsgList = append(storageMsgList, v)
					} else {
						if isSenderSync || msgChannelValue.aggregationID != v.MsgData.SendID {
							notStoragePushMsgList = append(notStoragePushMsgList, v)
						}
					}
				}
				if len(storageMsgList) > 0 {
					err, lastSeq := db.DB.BatchInsertChat2Cache(msgChannelValue.aggregationID, storageMsgList, triggerID) //redis append seq and modify user seq
					if err != nil {
						singleMsgFailedCount += uint64(len(storageMsgList))
						log.NewError(triggerID, "single data insert to redis err", err.Error(), storageMsgList)
					} else {
						singleMsgSuccessCountMutex.Lock()
						singleMsgSuccessCount += uint64(len(storageMsgList))
						singleMsgSuccessCountMutex.Unlock()
						och.SendMessageToMongoCH(msgChannelValue.aggregationID, triggerID, storageMsgList, lastSeq)
						for _, v := range storageMsgList {
							sendMessageToPushMQ(v, msgChannelValue.aggregationID)
						}
					}
				}
				if len(notStoragePushMsgList) > 0 {
					for _, x := range notStoragePushMsgList {
						sendMessageToPushMQ(x, msgChannelValue.aggregationID)
					}
				}
			}
		}
	}
}

func sendMessageToPushMQ(message *pbMsg.MsgDataToMQ, pushToUserID string) {
	//log.Info(message.OperationID, utils.GetSelfFuncName(), "msg ", message.String(), pushToUserID)
	mqPushMsg := pbMsg.PushMsgDataToMQ{OperationID: message.OperationID, MsgData: message.MsgData, PushToUserID: pushToUserID}
	pid, offset, err := producer.SendMessage(&mqPushMsg, mqPushMsg.PushToUserID, message.OperationID)
	if err != nil {
		log.Error(mqPushMsg.OperationID, "kafka send failed", "send data", message.String(), "pid", pid, "offset", offset, "err", err.Error())
	}
}

func (och *OnlineHistoryRedisConsumerHandler) SendMessageToMongoCH(aggregationID string, triggerID string, messages []*pbMsg.MsgDataToMQ, lastSeq uint64) {
	//log.Info(triggerID, utils.GetSelfFuncName(), "msg ", len(messages), lastSeq)
	if len(messages) > 0 {
		pid, offset, err := producerToMongo.SendMessage(&pbMsg.MsgDataToMongoByMQ{LastSeq: lastSeq, AggregationID: aggregationID, MessageList: messages, TriggerID: triggerID}, aggregationID, triggerID)
		if err != nil {
			log.Error(triggerID, "kafka send failed", "send data", len(messages), "pid", pid, "offset", offset, "err", err.Error(), "key", aggregationID)
		} else {
			//	log.NewWarn(m.OperationID, "sendMsgToKafka   client msgID ", m.MsgData.ClientMsgID)
		}
	}
}

func (och *OnlineHistoryRedisConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error { // a instance in the consumer group
	for {
		if sess == nil {
			log.NewWarn("", " sess == nil, waiting ")
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
	rwLock := new(sync.RWMutex)
	cMsg := make([]*sarama.ConsumerMessage, 0, 1000)
	t := time.NewTicker(time.Duration(100) * time.Millisecond)
	var triggerID string
	go func() {
		for {
			select {
			case <-t.C:
				if len(cMsg) > 0 {
					rwLock.Lock()
					ccMsg := make([]*sarama.ConsumerMessage, 0, 1000)
					for _, v := range cMsg {
						ccMsg = append(ccMsg, v)
					}
					cMsg = make([]*sarama.ConsumerMessage, 0, 1000)
					rwLock.Unlock()
					split := 1000
					triggerID = utils.OperationIDGenerator()
					for i := 0; i < len(ccMsg)/split; i++ {
						och.msgDistributionCh <- Cmd2Value{Cmd: ConsumerMsgs, Value: TriggerChannelValue{
							triggerID: triggerID, cmsgList: ccMsg[i*split : (i+1)*split]}}
					}
					if (len(ccMsg) % split) > 0 {
						och.msgDistributionCh <- Cmd2Value{Cmd: ConsumerMsgs, Value: TriggerChannelValue{
							triggerID: triggerID, cmsgList: ccMsg[split*(len(ccMsg)/split):]}}
					}
				}
			}
		}
	}()
	for msg := range claim.Messages() {
		rwLock.Lock()
		if len(msg.Value) != 0 {
			cMsg = append(cMsg, msg)
		}
		rwLock.Unlock()
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (och *OnlineHistoryRedisConsumerHandler) MessagesDistributionHandle() {
	for {
		aggregationMsgs := make(map[string][]*pbMsg.MsgDataToMQ, ChannelNum)
		select {
		case cmd := <-och.msgDistributionCh:
			switch cmd.Cmd {
			case ConsumerMsgs:
				triggerChannelValue := cmd.Value.(TriggerChannelValue)
				triggerID := triggerChannelValue.triggerID
				consumerMessages := triggerChannelValue.cmsgList
				for i := 0; i < len(consumerMessages); i++ {
					msgFromMQ := pbMsg.MsgDataToMQ{}
					err := proto.Unmarshal(consumerMessages[i].Value, &msgFromMQ)
					if err != nil {
						log.Error(triggerID, "msg_transfer Unmarshal msg err", "msg", string(consumerMessages[i].Value), "err", err.Error())
						return
					}
					if oldM, ok := aggregationMsgs[string(consumerMessages[i].Key)]; ok {
						oldM = append(oldM, &msgFromMQ)
						aggregationMsgs[string(consumerMessages[i].Key)] = oldM
					} else {
						m := make([]*pbMsg.MsgDataToMQ, 0, 100)
						m = append(m, &msgFromMQ)
						aggregationMsgs[string(consumerMessages[i].Key)] = m
					}
				}
				for aggregationID, v := range aggregationMsgs {
					if len(v) >= 0 {
						hashCode := getHashCode(aggregationID)
						channelID := hashCode % ChannelNum
						och.chArrays[channelID] <- Cmd2Value{Cmd: AggregationMessages, Value: MsgChannelValue{aggregationID: aggregationID, msgList: v, triggerID: triggerID}}
					}
				}
			}
		}
	}
}


func (OnlineHistoryRedisConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (OnlineHistoryRedisConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }


// String hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func getHashCode(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

