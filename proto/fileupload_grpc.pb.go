// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: fileupload.proto

package proto

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

const (
	Uploader_UploadFile_FullMethodName = "/fileupload.Uploader/UploadFile"
)

// UploaderClient is the client API for Uploader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UploaderClient interface {
	UploadFile(ctx context.Context, opts ...grpc.CallOption) (Uploader_UploadFileClient, error)
}

type uploaderClient struct {
	cc grpc.ClientConnInterface
}

func NewUploaderClient(cc grpc.ClientConnInterface) UploaderClient {
	return &uploaderClient{cc}
}

func (c *uploaderClient) UploadFile(ctx context.Context, opts ...grpc.CallOption) (Uploader_UploadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &Uploader_ServiceDesc.Streams[0], Uploader_UploadFile_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &uploaderUploadFileClient{stream}
	return x, nil
}

type Uploader_UploadFileClient interface {
	Send(*UploadRequest) error
	CloseAndRecv() (*UploadResponse, error)
	grpc.ClientStream
}

type uploaderUploadFileClient struct {
	grpc.ClientStream
}

func (x *uploaderUploadFileClient) Send(m *UploadRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *uploaderUploadFileClient) CloseAndRecv() (*UploadResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(UploadResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UploaderServer is the server API for Uploader service.
// All implementations must embed UnimplementedUploaderServer
// for forward compatibility
type UploaderServer interface {
	UploadFile(Uploader_UploadFileServer) error
	mustEmbedUnimplementedUploaderServer()
}

// UnimplementedUploaderServer must be embedded to have forward compatible implementations.
type UnimplementedUploaderServer struct {
}

func (UnimplementedUploaderServer) UploadFile(Uploader_UploadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadFile not implemented")
}
func (UnimplementedUploaderServer) mustEmbedUnimplementedUploaderServer() {}

// UnsafeUploaderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UploaderServer will
// result in compilation errors.
type UnsafeUploaderServer interface {
	mustEmbedUnimplementedUploaderServer()
}

func RegisterUploaderServer(s grpc.ServiceRegistrar, srv UploaderServer) {
	s.RegisterService(&Uploader_ServiceDesc, srv)
}

func _Uploader_UploadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(UploaderServer).UploadFile(&uploaderUploadFileServer{stream})
}

type Uploader_UploadFileServer interface {
	SendAndClose(*UploadResponse) error
	Recv() (*UploadRequest, error)
	grpc.ServerStream
}

type uploaderUploadFileServer struct {
	grpc.ServerStream
}

func (x *uploaderUploadFileServer) SendAndClose(m *UploadResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *uploaderUploadFileServer) Recv() (*UploadRequest, error) {
	m := new(UploadRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Uploader_ServiceDesc is the grpc.ServiceDesc for Uploader service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Uploader_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fileupload.Uploader",
	HandlerType: (*UploaderServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadFile",
			Handler:       _Uploader_UploadFile_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "fileupload.proto",
}
