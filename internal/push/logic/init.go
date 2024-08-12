package logic

import (
	pusher "Open_IM/internal/push"
	fcm "Open_IM/internal/push/fcm"
	"Open_IM/internal/push/getui"
	jpush "Open_IM/internal/push/jpush"
	"Open_IM/internal/push/mobpush"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	promePkg "Open_IM/pkg/common/prometheus"
	"Open_IM/pkg/statistics"
	"fmt"
)

var (
	rpcServer RPCServer
	pushCh    PushConsumerHandler
	//producer      *kafka.Producer
	offlinePusher pusher.OfflinePusher
	successCount  uint64
)

func Init(rpcPort int) {
	rpcServer.Init(rpcPort)
	pushCh.Init()

}
func init() {
	//producer = kafka.NewKafkaProducer(config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.Ws2mschat.Topic)
	statistics.NewStatistics(&successCount, config.Config.ModuleName.PushName, fmt.Sprintf("%d second push to msg_gateway count", constant.StatisticsTimeInterval), constant.StatisticsTimeInterval)
	if config.Config.Push.Getui.Enable {
		offlinePusher = getui.GetuiClient
	}
	if config.Config.Push.Jpns.Enable {
		offlinePusher = jpush.JPushClient
	}

	if config.Config.Push.Fcm.Enable {
		offlinePusher = fcm.NewFcm()
	}

	if config.Config.Push.Mob.Enable {
		offlinePusher = mobpush.MobPushClient
	}
}

func initPrometheus() {
	promePkg.NewMsgOfflinePushSuccessCounter()
	promePkg.NewMsgOfflinePushFailedCounter()
}

func Run(promethuesPort int) {
	go rpcServer.run()
	go pushCh.pushConsumerGroup.RegisterHandleAndConsumer(&pushCh)
	go func() {
		err := promePkg.StartPromeSrv(promethuesPort)
		if err != nil {
			panic(err)
		}
	}()
}
