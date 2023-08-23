package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/kafka"
	promePkg "Open_IM/pkg/common/prometheus"
	"Open_IM/pkg/statistics"
	"fmt"
	"sync"
)

const OnlineTopicBusy = 1
const OnlineTopicVacancy = 0
const Msg = 2
const ConsumerMsgs = 3
const AggregationMessages = 4
const MongoMessages = 5
const ChannelNum = 100

var (
	persistentCH          PersistentConsumerHandler   // ws2mschat save mysql
	historyCH             OnlineHistoryRedisConsumerHandler
	historyMongoCH        OnlineHistoryMongoConsumerHandler
	producer              *kafka.Producer       // push Ms2pschat
	producerToMongo       *kafka.Producer       // mongo MsgToMongo
	cmdCh                 chan Cmd2Value
	onlineTopicStatus     int
	w                     *sync.Mutex
	singleMsgSuccessCount uint64
	singleMsgFailedCount  uint64

	singleMsgSuccessCountMutex sync.Mutex
)

func Init() {
	cmdCh = make(chan Cmd2Value, 10000)
	w = new(sync.Mutex)
	if config.Config.Prometheus.Enable {
		initPrometheus()
	}
	persistentCH.Init()
	historyCH.Init(cmdCh) //
	historyMongoCH.Init()
	onlineTopicStatus = OnlineTopicVacancy

	statistics.NewStatistics(&singleMsgSuccessCount, config.Config.ModuleName.MsgTransferName, fmt.Sprintf("%d second singleMsgCount insert to mongo", constant.StatisticsTimeInterval), constant.StatisticsTimeInterval)
	producer = kafka.NewKafkaProducer(config.Config.Kafka.Ms2pschat.Addr, config.Config.Kafka.Ms2pschat.Topic)
	producerToMongo = kafka.NewKafkaProducer(config.Config.Kafka.MsgToMongo.Addr, config.Config.Kafka.MsgToMongo.Topic)
}
func Run(promethuesPort int) {
	//register mysqlConsumerHandler to
	if config.Config.ChatPersistenceMysql {
		go persistentCH.persistentConsumerGroup.RegisterHandleAndConsumer(&persistentCH)
	} else {
		fmt.Println("not start mysql consumer")
	}
	go historyCH.historyConsumerGroup.RegisterHandleAndConsumer(&historyCH)
	go historyMongoCH.historyConsumerGroup.RegisterHandleAndConsumer(&historyMongoCH)
	go func() {
		err := promePkg.StartPromeSrv(promethuesPort)
		if err != nil {
			panic(err)
		}
	}()
}
func SetOnlineTopicStatus(status int) {
	w.Lock()
	defer w.Unlock()
	onlineTopicStatus = status
}
func GetOnlineTopicStatus() int {
	w.Lock()
	defer w.Unlock()
	return onlineTopicStatus
}
