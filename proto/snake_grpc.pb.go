// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.19.6
// source: snake.proto

package snake

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	SnakeGame_JoinGame_FullMethodName      = "/snake.SnakeGame/JoinGame"
	SnakeGame_SendDirection_FullMethodName = "/snake.SnakeGame/SendDirection"
	SnakeGame_GetTopPlayers_FullMethodName = "/snake.SnakeGame/GetTopPlayers"
)

// SnakeGameClient is the client API for SnakeGame service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SnakeGameClient interface {
	JoinGame(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[GameState], error)
	SendDirection(ctx context.Context, in *DirectionRequest, opts ...grpc.CallOption) (*Empty, error)
	GetTopPlayers(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*TopPlayersResponse, error)
}

type snakeGameClient struct {
	cc grpc.ClientConnInterface
}

func NewSnakeGameClient(cc grpc.ClientConnInterface) SnakeGameClient {
	return &snakeGameClient{cc}
}

func (c *snakeGameClient) JoinGame(ctx context.Context, in *JoinRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[GameState], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &SnakeGame_ServiceDesc.Streams[0], SnakeGame_JoinGame_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[JoinRequest, GameState]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type SnakeGame_JoinGameClient = grpc.ServerStreamingClient[GameState]

func (c *snakeGameClient) SendDirection(ctx context.Context, in *DirectionRequest, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, SnakeGame_SendDirection_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *snakeGameClient) GetTopPlayers(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*TopPlayersResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TopPlayersResponse)
	err := c.cc.Invoke(ctx, SnakeGame_GetTopPlayers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SnakeGameServer is the server API for SnakeGame service.
// All implementations must embed UnimplementedSnakeGameServer
// for forward compatibility.
type SnakeGameServer interface {
	JoinGame(*JoinRequest, grpc.ServerStreamingServer[GameState]) error
	SendDirection(context.Context, *DirectionRequest) (*Empty, error)
	GetTopPlayers(context.Context, *Empty) (*TopPlayersResponse, error)
	mustEmbedUnimplementedSnakeGameServer()
}

// UnimplementedSnakeGameServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedSnakeGameServer struct{}

func (UnimplementedSnakeGameServer) JoinGame(*JoinRequest, grpc.ServerStreamingServer[GameState]) error {
	return status.Errorf(codes.Unimplemented, "method JoinGame not implemented")
}
func (UnimplementedSnakeGameServer) SendDirection(context.Context, *DirectionRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendDirection not implemented")
}
func (UnimplementedSnakeGameServer) GetTopPlayers(context.Context, *Empty) (*TopPlayersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTopPlayers not implemented")
}
func (UnimplementedSnakeGameServer) mustEmbedUnimplementedSnakeGameServer() {}
func (UnimplementedSnakeGameServer) testEmbeddedByValue()                   {}

// UnsafeSnakeGameServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SnakeGameServer will
// result in compilation errors.
type UnsafeSnakeGameServer interface {
	mustEmbedUnimplementedSnakeGameServer()
}

func RegisterSnakeGameServer(s grpc.ServiceRegistrar, srv SnakeGameServer) {
	// If the following call pancis, it indicates UnimplementedSnakeGameServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&SnakeGame_ServiceDesc, srv)
}

func _SnakeGame_JoinGame_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(JoinRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SnakeGameServer).JoinGame(m, &grpc.GenericServerStream[JoinRequest, GameState]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type SnakeGame_JoinGameServer = grpc.ServerStreamingServer[GameState]

func _SnakeGame_SendDirection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DirectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnakeGameServer).SendDirection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SnakeGame_SendDirection_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnakeGameServer).SendDirection(ctx, req.(*DirectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SnakeGame_GetTopPlayers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnakeGameServer).GetTopPlayers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SnakeGame_GetTopPlayers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnakeGameServer).GetTopPlayers(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// SnakeGame_ServiceDesc is the grpc.ServiceDesc for SnakeGame service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SnakeGame_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "snake.SnakeGame",
	HandlerType: (*SnakeGameServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendDirection",
			Handler:    _SnakeGame_SendDirection_Handler,
		},
		{
			MethodName: "GetTopPlayers",
			Handler:    _SnakeGame_GetTopPlayers_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "JoinGame",
			Handler:       _SnakeGame_JoinGame_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "snake.proto",
}
