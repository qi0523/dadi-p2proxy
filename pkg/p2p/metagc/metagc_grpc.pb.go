// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: pkg/p2p/metagc/metagc.proto

package metagc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MetaGcServiceClient is the client API for MetaGcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetaGcServiceClient interface {
	GcMetadata(ctx context.Context, in *MetaGcRequest, opts ...grpc.CallOption) (*MetaGcResponse, error)
}

type metaGcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMetaGcServiceClient(cc grpc.ClientConnInterface) MetaGcServiceClient {
	return &metaGcServiceClient{cc}
}

func (c *metaGcServiceClient) GcMetadata(ctx context.Context, in *MetaGcRequest, opts ...grpc.CallOption) (*MetaGcResponse, error) {
	out := new(MetaGcResponse)
	err := c.cc.Invoke(ctx, "/metagc.MetaGcService/GcMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetaGcServiceServer is the server API for MetaGcService service.
// All implementations must embed UnimplementedMetaGcServiceServer
// for forward compatibility
type MetaGcServiceServer interface {
	GcMetadata(context.Context, *MetaGcRequest) (*MetaGcResponse, error)
	mustEmbedUnimplementedMetaGcServiceServer()
}

// UnimplementedMetaGcServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMetaGcServiceServer struct {
}

func (UnimplementedMetaGcServiceServer) GcMetadata(context.Context, *MetaGcRequest) (*MetaGcResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GcMetadata not implemented")
}
func (UnimplementedMetaGcServiceServer) mustEmbedUnimplementedMetaGcServiceServer() {}

// UnsafeMetaGcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetaGcServiceServer will
// result in compilation errors.
type UnsafeMetaGcServiceServer interface {
	mustEmbedUnimplementedMetaGcServiceServer()
}

func RegisterMetaGcServiceServer(s grpc.ServiceRegistrar, srv MetaGcServiceServer) {
	s.RegisterService(&MetaGcService_ServiceDesc, srv)
}

func _MetaGcService_GcMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetaGcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetaGcServiceServer).GcMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/metagc.MetaGcService/GcMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetaGcServiceServer).GcMetadata(ctx, req.(*MetaGcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MetaGcService_ServiceDesc is the grpc.ServiceDesc for MetaGcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetaGcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "metagc.MetaGcService",
	HandlerType: (*MetaGcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GcMetadata",
			Handler:    _MetaGcService_GcMetadata_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/p2p/metagc/metagc.proto",
}