// Code generated by protoc-gen-go. DO NOT EDIT.
// source: conversation/conversation.proto

/*
Package conversation is a generated protocol buffer package.

It is generated from these files:
	conversation/conversation.proto

It has these top-level messages:
	CommonResp
	Conversation
	ModifyConversationFieldReq
	ModifyConversationFieldResp
*/
package conversation

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CommonResp struct {
	ErrCode int32  `protobuf:"varint,1,opt,name=errCode" json:"errCode,omitempty"`
	ErrMsg  string `protobuf:"bytes,2,opt,name=errMsg" json:"errMsg,omitempty"`
}

func (m *CommonResp) Reset()                    { *m = CommonResp{} }
func (m *CommonResp) String() string            { return proto.CompactTextString(m) }
func (*CommonResp) ProtoMessage()               {}
func (*CommonResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CommonResp) GetErrCode() int32 {
	if m != nil {
		return m.ErrCode
	}
	return 0
}

func (m *CommonResp) GetErrMsg() string {
	if m != nil {
		return m.ErrMsg
	}
	return ""
}

type Conversation struct {
	OwnerUserID           string `protobuf:"bytes,1,opt,name=ownerUserID" json:"ownerUserID,omitempty"`
	ConversationID        string `protobuf:"bytes,2,opt,name=conversationID" json:"conversationID,omitempty"`
	RecvMsgOpt            int32  `protobuf:"varint,3,opt,name=recvMsgOpt" json:"recvMsgOpt,omitempty"`
	ConversationType      int32  `protobuf:"varint,4,opt,name=conversationType" json:"conversationType,omitempty"`
	UserID                string `protobuf:"bytes,5,opt,name=userID" json:"userID,omitempty"`
	GroupID               string `protobuf:"bytes,6,opt,name=groupID" json:"groupID,omitempty"`
	UnreadCount           int32  `protobuf:"varint,7,opt,name=unreadCount" json:"unreadCount,omitempty"`
	DraftTextTime         int64  `protobuf:"varint,8,opt,name=draftTextTime" json:"draftTextTime,omitempty"`
	IsPinned              bool   `protobuf:"varint,9,opt,name=isPinned" json:"isPinned,omitempty"`
	AttachedInfo          string `protobuf:"bytes,10,opt,name=attachedInfo" json:"attachedInfo,omitempty"`
	IsPrivateChat         bool   `protobuf:"varint,11,opt,name=isPrivateChat" json:"isPrivateChat,omitempty"`
	GroupAtType           int32  `protobuf:"varint,12,opt,name=groupAtType" json:"groupAtType,omitempty"`
	IsNotInGroup          bool   `protobuf:"varint,13,opt,name=isNotInGroup" json:"isNotInGroup,omitempty"`
	Ex                    string `protobuf:"bytes,14,opt,name=ex" json:"ex,omitempty"`
	UpdateUnreadCountTime int64  `protobuf:"varint,15,opt,name=updateUnreadCountTime" json:"updateUnreadCountTime,omitempty"`
}

func (m *Conversation) Reset()                    { *m = Conversation{} }
func (m *Conversation) String() string            { return proto.CompactTextString(m) }
func (*Conversation) ProtoMessage()               {}
func (*Conversation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Conversation) GetOwnerUserID() string {
	if m != nil {
		return m.OwnerUserID
	}
	return ""
}

func (m *Conversation) GetConversationID() string {
	if m != nil {
		return m.ConversationID
	}
	return ""
}

func (m *Conversation) GetRecvMsgOpt() int32 {
	if m != nil {
		return m.RecvMsgOpt
	}
	return 0
}

func (m *Conversation) GetConversationType() int32 {
	if m != nil {
		return m.ConversationType
	}
	return 0
}

func (m *Conversation) GetUserID() string {
	if m != nil {
		return m.UserID
	}
	return ""
}

func (m *Conversation) GetGroupID() string {
	if m != nil {
		return m.GroupID
	}
	return ""
}

func (m *Conversation) GetUnreadCount() int32 {
	if m != nil {
		return m.UnreadCount
	}
	return 0
}

func (m *Conversation) GetDraftTextTime() int64 {
	if m != nil {
		return m.DraftTextTime
	}
	return 0
}

func (m *Conversation) GetIsPinned() bool {
	if m != nil {
		return m.IsPinned
	}
	return false
}

func (m *Conversation) GetAttachedInfo() string {
	if m != nil {
		return m.AttachedInfo
	}
	return ""
}

func (m *Conversation) GetIsPrivateChat() bool {
	if m != nil {
		return m.IsPrivateChat
	}
	return false
}

func (m *Conversation) GetGroupAtType() int32 {
	if m != nil {
		return m.GroupAtType
	}
	return 0
}

func (m *Conversation) GetIsNotInGroup() bool {
	if m != nil {
		return m.IsNotInGroup
	}
	return false
}

func (m *Conversation) GetEx() string {
	if m != nil {
		return m.Ex
	}
	return ""
}

func (m *Conversation) GetUpdateUnreadCountTime() int64 {
	if m != nil {
		return m.UpdateUnreadCountTime
	}
	return 0
}

type ModifyConversationFieldReq struct {
	Conversation *Conversation `protobuf:"bytes,1,opt,name=conversation" json:"conversation,omitempty"`
	FieldType    int32         `protobuf:"varint,2,opt,name=fieldType" json:"fieldType,omitempty"`
	UserIDList   []string      `protobuf:"bytes,3,rep,name=userIDList" json:"userIDList,omitempty"`
	OperationID  string        `protobuf:"bytes,4,opt,name=operationID" json:"operationID,omitempty"`
}

func (m *ModifyConversationFieldReq) Reset()                    { *m = ModifyConversationFieldReq{} }
func (m *ModifyConversationFieldReq) String() string            { return proto.CompactTextString(m) }
func (*ModifyConversationFieldReq) ProtoMessage()               {}
func (*ModifyConversationFieldReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ModifyConversationFieldReq) GetConversation() *Conversation {
	if m != nil {
		return m.Conversation
	}
	return nil
}

func (m *ModifyConversationFieldReq) GetFieldType() int32 {
	if m != nil {
		return m.FieldType
	}
	return 0
}

func (m *ModifyConversationFieldReq) GetUserIDList() []string {
	if m != nil {
		return m.UserIDList
	}
	return nil
}

func (m *ModifyConversationFieldReq) GetOperationID() string {
	if m != nil {
		return m.OperationID
	}
	return ""
}

type ModifyConversationFieldResp struct {
	CommonResp *CommonResp `protobuf:"bytes,1,opt,name=commonResp" json:"commonResp,omitempty"`
}

func (m *ModifyConversationFieldResp) Reset()                    { *m = ModifyConversationFieldResp{} }
func (m *ModifyConversationFieldResp) String() string            { return proto.CompactTextString(m) }
func (*ModifyConversationFieldResp) ProtoMessage()               {}
func (*ModifyConversationFieldResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ModifyConversationFieldResp) GetCommonResp() *CommonResp {
	if m != nil {
		return m.CommonResp
	}
	return nil
}

func init() {
	proto.RegisterType((*CommonResp)(nil), "conversation.CommonResp")
	proto.RegisterType((*Conversation)(nil), "conversation.Conversation")
	proto.RegisterType((*ModifyConversationFieldReq)(nil), "conversation.ModifyConversationFieldReq")
	proto.RegisterType((*ModifyConversationFieldResp)(nil), "conversation.ModifyConversationFieldResp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Conversation service

type ConversationClient interface {
	ModifyConversationField(ctx context.Context, in *ModifyConversationFieldReq, opts ...grpc.CallOption) (*ModifyConversationFieldResp, error)
}

type conversationClient struct {
	cc *grpc.ClientConn
}

func NewConversationClient(cc *grpc.ClientConn) ConversationClient {
	return &conversationClient{cc}
}

func (c *conversationClient) ModifyConversationField(ctx context.Context, in *ModifyConversationFieldReq, opts ...grpc.CallOption) (*ModifyConversationFieldResp, error) {
	out := new(ModifyConversationFieldResp)
	err := grpc.Invoke(ctx, "/conversation.conversation/ModifyConversationField", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Conversation service

type ConversationServer interface {
	ModifyConversationField(context.Context, *ModifyConversationFieldReq) (*ModifyConversationFieldResp, error)
}

func RegisterConversationServer(s *grpc.Server, srv ConversationServer) {
	s.RegisterService(&_Conversation_serviceDesc, srv)
}

func _Conversation_ModifyConversationField_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModifyConversationFieldReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConversationServer).ModifyConversationField(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/conversation.conversation/ModifyConversationField",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConversationServer).ModifyConversationField(ctx, req.(*ModifyConversationFieldReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Conversation_serviceDesc = grpc.ServiceDesc{
	ServiceName: "conversation.conversation",
	HandlerType: (*ConversationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ModifyConversationField",
			Handler:    _Conversation_ModifyConversationField_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "conversation/conversation.proto",
}

func init() { proto.RegisterFile("conversation/conversation.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 505 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0x51, 0x8b, 0x13, 0x31,
	0x14, 0x85, 0x99, 0x6e, 0xb7, 0xdb, 0xde, 0x76, 0xab, 0x04, 0xd4, 0x50, 0x45, 0x4b, 0x11, 0x19,
	0x15, 0xb7, 0xb0, 0xfa, 0x20, 0x08, 0x0b, 0xda, 0xa2, 0x0c, 0x58, 0x77, 0x09, 0x5d, 0x04, 0x5f,
	0x64, 0xec, 0xdc, 0x76, 0x83, 0x36, 0x89, 0x49, 0x5a, 0xbb, 0x2f, 0xfe, 0x06, 0x7f, 0x90, 0x3f,
	0x4e, 0x92, 0x69, 0xb7, 0xc9, 0x6a, 0xc1, 0xc7, 0xfb, 0xe5, 0xce, 0xc9, 0x39, 0xd3, 0xd3, 0x81,
	0x07, 0x13, 0x29, 0x96, 0xa8, 0x4d, 0x6e, 0xb9, 0x14, 0xfd, 0x70, 0x38, 0x52, 0x5a, 0x5a, 0x49,
	0x5a, 0x21, 0xeb, 0x9d, 0x00, 0x0c, 0xe4, 0x7c, 0x2e, 0x05, 0x43, 0xa3, 0x08, 0x85, 0x03, 0xd4,
	0x7a, 0x20, 0x0b, 0xa4, 0x49, 0x37, 0x49, 0xf7, 0xd9, 0x66, 0x24, 0xb7, 0xa1, 0x86, 0x5a, 0x8f,
	0xcc, 0x8c, 0x56, 0xba, 0x49, 0xda, 0x60, 0xeb, 0xa9, 0xf7, 0xab, 0x0a, 0xad, 0x41, 0x20, 0x48,
	0xba, 0xd0, 0x94, 0x3f, 0x04, 0xea, 0x73, 0x83, 0x3a, 0x1b, 0x7a, 0x99, 0x06, 0x0b, 0x11, 0x79,
	0x04, 0xed, 0xd0, 0x42, 0x36, 0x5c, 0x4b, 0x5e, 0xa3, 0xe4, 0x3e, 0x80, 0xc6, 0xc9, 0x72, 0x64,
	0x66, 0xa7, 0xca, 0xd2, 0x3d, 0xef, 0x27, 0x20, 0xe4, 0x09, 0xdc, 0x0c, 0x9f, 0x18, 0x5f, 0x2a,
	0xa4, 0x55, 0xbf, 0xf5, 0x17, 0x77, 0xf6, 0x17, 0xa5, 0xa1, 0xfd, 0xd2, 0x7e, 0x39, 0xb9, 0xc0,
	0x33, 0x2d, 0x17, 0x2a, 0x1b, 0xd2, 0x9a, 0x3f, 0xd8, 0x8c, 0x2e, 0xc7, 0x42, 0x68, 0xcc, 0x8b,
	0x81, 0x5c, 0x08, 0x4b, 0x0f, 0xbc, 0x70, 0x88, 0xc8, 0x43, 0x38, 0x2c, 0x74, 0x3e, 0xb5, 0x63,
	0x5c, 0xd9, 0x31, 0x9f, 0x23, 0xad, 0x77, 0x93, 0x74, 0x8f, 0xc5, 0x90, 0x74, 0xa0, 0xce, 0xcd,
	0x19, 0x17, 0x02, 0x0b, 0xda, 0xe8, 0x26, 0x69, 0x9d, 0x5d, 0xcd, 0xa4, 0x07, 0xad, 0xdc, 0xda,
	0x7c, 0x72, 0x81, 0x45, 0x26, 0xa6, 0x92, 0x82, 0xb7, 0x10, 0x31, 0x77, 0x0b, 0x37, 0x67, 0x9a,
	0x2f, 0x73, 0x8b, 0x83, 0x8b, 0xdc, 0xd2, 0xa6, 0x17, 0x89, 0xa1, 0x73, 0xeb, 0x8d, 0xbf, 0xb6,
	0xfe, 0x35, 0xb4, 0x4a, 0xb7, 0x01, 0x72, 0x77, 0x71, 0xf3, 0x41, 0xda, 0x4c, 0xbc, 0x73, 0x94,
	0x1e, 0x7a, 0x99, 0x88, 0x91, 0x36, 0x54, 0x70, 0x45, 0xdb, 0xde, 0x45, 0x05, 0x57, 0xe4, 0x05,
	0xdc, 0x5a, 0xa8, 0x22, 0xb7, 0x78, 0xbe, 0x8d, 0xed, 0x93, 0xde, 0xf0, 0x49, 0xff, 0x7d, 0xd8,
	0xfb, 0x9d, 0x40, 0x67, 0x24, 0x0b, 0x3e, 0xbd, 0x0c, 0x8b, 0xf1, 0x96, 0xe3, 0xb7, 0x82, 0xe1,
	0x77, 0x72, 0x02, 0x51, 0x03, 0x7d, 0x43, 0x9a, 0xc7, 0x9d, 0xa3, 0xa8, 0xaa, 0xe1, 0x93, 0x2c,
	0xda, 0x27, 0xf7, 0xa0, 0x31, 0x75, 0x5a, 0x3e, 0x68, 0xc5, 0x07, 0xdd, 0x02, 0x57, 0x9a, 0xf2,
	0xa7, 0x7d, 0xcf, 0x8d, 0x2b, 0xcd, 0x5e, 0xda, 0x60, 0x01, 0xf1, 0xf5, 0x54, 0xa8, 0x37, 0xcd,
	0xab, 0xae, 0xeb, 0xb9, 0x45, 0xbd, 0x8f, 0x70, 0x77, 0xa7, 0x7b, 0xa3, 0xc8, 0x4b, 0x80, 0xc9,
	0xd5, 0x1f, 0x66, 0x6d, 0x9e, 0x5e, 0x37, 0xbf, 0x39, 0x67, 0xc1, 0xee, 0xf1, 0xcf, 0x38, 0x38,
	0x11, 0x70, 0x67, 0xc7, 0x45, 0x24, 0x8d, 0x05, 0x77, 0xbf, 0xcd, 0xce, 0xe3, 0xff, 0xdc, 0x34,
	0xea, 0xcd, 0xb3, 0x4f, 0x4f, 0x4f, 0x15, 0x8a, 0xcf, 0xd9, 0xa8, 0xaf, 0xbe, 0xce, 0xfa, 0xfe,
	0x6b, 0x10, 0x7d, 0x20, 0x5e, 0x85, 0xc3, 0x97, 0x9a, 0x5f, 0x78, 0xfe, 0x27, 0x00, 0x00, 0xff,
	0xff, 0x58, 0x04, 0xda, 0x8a, 0x51, 0x04, 0x00, 0x00,
}
