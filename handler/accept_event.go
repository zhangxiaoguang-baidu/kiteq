package handler

import (
	"errors"
	log "github.com/blackbeans/log4go"
	"github.com/blackbeans/turbo/pipe"
	"kiteq/protocol"
	"kiteq/store"
	"os"
	"time"
)

//--------------------如下为具体的处理Handler
type AcceptHandler struct {
	pipe.BaseForwardHandler
	topics     []string
	kiteserver string
}

func NewAcceptHandler(name string) *AcceptHandler {
	ahandler := &AcceptHandler{}
	ahandler.BaseForwardHandler = pipe.NewBaseForwardHandler(name, ahandler)
	hn, _ := os.Hostname()
	ahandler.kiteserver = hn
	return ahandler
}

func (self *AcceptHandler) TypeAssert(event pipe.IEvent) bool {
	_, ok := self.cast(event)
	return ok
}

func (self *AcceptHandler) cast(event pipe.IEvent) (val *acceptEvent, ok bool) {
	val, ok = event.(*acceptEvent)
	return
}

var INVALID_MSG_TYPE_ERROR = errors.New("INVALID MSG TYPE !")

func (self *AcceptHandler) Process(ctx *pipe.DefaultPipelineContext, event pipe.IEvent) error {
	// log.Debug("AcceptHandler|Process|%s|%t\n", self.GetName(), event)

	ae, ok := self.cast(event)
	if !ok {
		return pipe.ERROR_INVALID_EVENT_TYPE
	}
	//这里处理一下ae,做一下校验
	var msg *store.MessageEntity
	switch ae.msgType {
	case protocol.CMD_DELIVER_ACK:
		//收到投递结果直接attach响应
		ae.remoteClient.Attach(ae.opaque, ae.msg)
		return nil
	case protocol.CMD_HEARTBEAT:
		hb := ae.msg.(*protocol.HeartBeat)
		event = pipe.NewHeartbeatEvent(ae.remoteClient, ae.opaque, hb.GetVersion())
		ctx.SendForward(event)
		return nil

	case protocol.CMD_BYTES_MESSAGE:
		msg = store.NewMessageEntity(protocol.NewQMessage(ae.msg.(*protocol.BytesMessage)))
	case protocol.CMD_STRING_MESSAGE:
		msg = store.NewMessageEntity(protocol.NewQMessage(ae.msg.(*protocol.StringMessage)))
	default:
		//这只是一个bug不支持的数据类型能给你
		log.Warn("AcceptHandler|Process|%s|%t\n", INVALID_MSG_TYPE_ERROR, ae.msg)
	}

	if nil != msg {
		msg.PublishTime = time.Now().Unix()
		msg.KiteServer = self.kiteserver
		deliver := newPersistentEvent(msg, ae.remoteClient, ae.opaque)
		ctx.SendForward(deliver)
		return nil
	}
	return INVALID_MSG_TYPE_ERROR
}
