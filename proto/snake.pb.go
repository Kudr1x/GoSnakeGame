// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v3.19.6
// source: snake.proto

package snake

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Direction int32

const (
	Direction_UP    Direction = 0
	Direction_DOWN  Direction = 1
	Direction_LEFT  Direction = 2
	Direction_RIGHT Direction = 3
)

// Enum value maps for Direction.
var (
	Direction_name = map[int32]string{
		0: "UP",
		1: "DOWN",
		2: "LEFT",
		3: "RIGHT",
	}
	Direction_value = map[string]int32{
		"UP":    0,
		"DOWN":  1,
		"LEFT":  2,
		"RIGHT": 3,
	}
)

func (x Direction) Enum() *Direction {
	p := new(Direction)
	*p = x
	return p
}

func (x Direction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Direction) Descriptor() protoreflect.EnumDescriptor {
	return file_snake_proto_enumTypes[0].Descriptor()
}

func (Direction) Type() protoreflect.EnumType {
	return &file_snake_proto_enumTypes[0]
}

func (x Direction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Direction.Descriptor instead.
func (Direction) EnumDescriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{0}
}

type JoinRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PlayerName    string                 `protobuf:"bytes,1,opt,name=player_name,json=playerName,proto3" json:"player_name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *JoinRequest) Reset() {
	*x = JoinRequest{}
	mi := &file_snake_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JoinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinRequest) ProtoMessage() {}

func (x *JoinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinRequest.ProtoReflect.Descriptor instead.
func (*JoinRequest) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{0}
}

func (x *JoinRequest) GetPlayerName() string {
	if x != nil {
		return x.PlayerName
	}
	return ""
}

type DirectionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PlayerName    string                 `protobuf:"bytes,1,opt,name=player_name,json=playerName,proto3" json:"player_name,omitempty"`
	Direction     Direction              `protobuf:"varint,2,opt,name=direction,proto3,enum=snake.Direction" json:"direction,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DirectionRequest) Reset() {
	*x = DirectionRequest{}
	mi := &file_snake_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DirectionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DirectionRequest) ProtoMessage() {}

func (x *DirectionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DirectionRequest.ProtoReflect.Descriptor instead.
func (*DirectionRequest) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{1}
}

func (x *DirectionRequest) GetPlayerName() string {
	if x != nil {
		return x.PlayerName
	}
	return ""
}

func (x *DirectionRequest) GetDirection() Direction {
	if x != nil {
		return x.Direction
	}
	return Direction_UP
}

type GameState struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Players       []*Player              `protobuf:"bytes,1,rep,name=players,proto3" json:"players,omitempty"`
	Food          []*Point               `protobuf:"bytes,2,rep,name=food,proto3" json:"food,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GameState) Reset() {
	*x = GameState{}
	mi := &file_snake_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GameState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameState) ProtoMessage() {}

func (x *GameState) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameState.ProtoReflect.Descriptor instead.
func (*GameState) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{2}
}

func (x *GameState) GetPlayers() []*Player {
	if x != nil {
		return x.Players
	}
	return nil
}

func (x *GameState) GetFood() []*Point {
	if x != nil {
		return x.Food
	}
	return nil
}

type Player struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Body          []*Point               `protobuf:"bytes,2,rep,name=body,proto3" json:"body,omitempty"`
	Alive         bool                   `protobuf:"varint,3,opt,name=alive,proto3" json:"alive,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Player) Reset() {
	*x = Player{}
	mi := &file_snake_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Player) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Player) ProtoMessage() {}

func (x *Player) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Player.ProtoReflect.Descriptor instead.
func (*Player) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{3}
}

func (x *Player) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Player) GetBody() []*Point {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Player) GetAlive() bool {
	if x != nil {
		return x.Alive
	}
	return false
}

type Point struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	X             int32                  `protobuf:"varint,1,opt,name=x,proto3" json:"x,omitempty"`
	Y             int32                  `protobuf:"varint,2,opt,name=y,proto3" json:"y,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Point) Reset() {
	*x = Point{}
	mi := &file_snake_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Point) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Point) ProtoMessage() {}

func (x *Point) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Point.ProtoReflect.Descriptor instead.
func (*Point) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{4}
}

func (x *Point) GetX() int32 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *Point) GetY() int32 {
	if x != nil {
		return x.Y
	}
	return 0
}

type Empty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_snake_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_snake_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_snake_proto_rawDescGZIP(), []int{5}
}

var File_snake_proto protoreflect.FileDescriptor

var file_snake_proto_rawDesc = string([]byte{
	0x0a, 0x0b, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x73,
	0x6e, 0x61, 0x6b, 0x65, 0x22, 0x2e, 0x0a, 0x0b, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x63, 0x0a, 0x10, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70,
	0x6c, 0x61, 0x79, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x09, 0x64, 0x69, 0x72,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x73,
	0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09,
	0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x56, 0x0a, 0x09, 0x47, 0x61, 0x6d,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x27, 0x0a, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e,
	0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x52, 0x07, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x12,
	0x20, 0x0a, 0x04, 0x66, 0x6f, 0x6f, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x04, 0x66, 0x6f, 0x6f,
	0x64, 0x22, 0x54, 0x0a, 0x06, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x20, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x04, 0x62, 0x6f, 0x64,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x05, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x22, 0x23, 0x0a, 0x05, 0x50, 0x6f, 0x69, 0x6e, 0x74,
	0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x78, 0x12, 0x0c,
	0x0a, 0x01, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x79, 0x22, 0x07, 0x0a, 0x05,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x2a, 0x32, 0x0a, 0x09, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x06, 0x0a, 0x02, 0x55, 0x50, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x4f,
	0x57, 0x4e, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x45, 0x46, 0x54, 0x10, 0x02, 0x12, 0x09,
	0x0a, 0x05, 0x52, 0x49, 0x47, 0x48, 0x54, 0x10, 0x03, 0x32, 0x77, 0x0a, 0x09, 0x53, 0x6e, 0x61,
	0x6b, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x32, 0x0a, 0x08, 0x4a, 0x6f, 0x69, 0x6e, 0x47, 0x61,
	0x6d, 0x65, 0x12, 0x12, 0x2e, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x47,
	0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x30, 0x01, 0x12, 0x36, 0x0a, 0x0d, 0x53, 0x65,
	0x6e, 0x64, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x2e, 0x73, 0x6e,
	0x61, 0x6b, 0x65, 0x2e, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x73, 0x6e, 0x61, 0x6b, 0x65, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_snake_proto_rawDescOnce sync.Once
	file_snake_proto_rawDescData []byte
)

func file_snake_proto_rawDescGZIP() []byte {
	file_snake_proto_rawDescOnce.Do(func() {
		file_snake_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_snake_proto_rawDesc), len(file_snake_proto_rawDesc)))
	})
	return file_snake_proto_rawDescData
}

var file_snake_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_snake_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_snake_proto_goTypes = []any{
	(Direction)(0),           // 0: snake.Direction
	(*JoinRequest)(nil),      // 1: snake.JoinRequest
	(*DirectionRequest)(nil), // 2: snake.DirectionRequest
	(*GameState)(nil),        // 3: snake.GameState
	(*Player)(nil),           // 4: snake.Player
	(*Point)(nil),            // 5: snake.Point
	(*Empty)(nil),            // 6: snake.Empty
}
var file_snake_proto_depIdxs = []int32{
	0, // 0: snake.DirectionRequest.direction:type_name -> snake.Direction
	4, // 1: snake.GameState.players:type_name -> snake.Player
	5, // 2: snake.GameState.food:type_name -> snake.Point
	5, // 3: snake.Player.body:type_name -> snake.Point
	1, // 4: snake.SnakeGame.JoinGame:input_type -> snake.JoinRequest
	2, // 5: snake.SnakeGame.SendDirection:input_type -> snake.DirectionRequest
	3, // 6: snake.SnakeGame.JoinGame:output_type -> snake.GameState
	6, // 7: snake.SnakeGame.SendDirection:output_type -> snake.Empty
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_snake_proto_init() }
func file_snake_proto_init() {
	if File_snake_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_snake_proto_rawDesc), len(file_snake_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_snake_proto_goTypes,
		DependencyIndexes: file_snake_proto_depIdxs,
		EnumInfos:         file_snake_proto_enumTypes,
		MessageInfos:      file_snake_proto_msgTypes,
	}.Build()
	File_snake_proto = out.File
	file_snake_proto_goTypes = nil
	file_snake_proto_depIdxs = nil
}
