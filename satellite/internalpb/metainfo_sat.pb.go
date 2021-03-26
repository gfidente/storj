// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: metainfo_sat.proto

package internalpb

import (
	fmt "fmt"
	math "math"
	time "time"

	proto "github.com/gogo/protobuf/proto"

	pb "storj.io/common/pb"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type StreamID struct {
	Bucket               []byte                   `protobuf:"bytes,1,opt,name=bucket,proto3" json:"bucket,omitempty"`
	EncryptedObjectKey   []byte                   `protobuf:"bytes,2,opt,name=encrypted_object_key,json=encryptedObjectKey,proto3" json:"encrypted_object_key,omitempty"`
	Version              int32                    `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
	EncryptionParameters *pb.EncryptionParameters `protobuf:"bytes,12,opt,name=encryption_parameters,json=encryptionParameters,proto3" json:"encryption_parameters,omitempty"`
	CreationDate         time.Time                `protobuf:"bytes,5,opt,name=creation_date,json=creationDate,proto3,stdtime" json:"creation_date"`
	ExpirationDate       time.Time                `protobuf:"bytes,6,opt,name=expiration_date,json=expirationDate,proto3,stdtime" json:"expiration_date"`
	MultipartObject      bool                     `protobuf:"varint,11,opt,name=multipart_object,json=multipartObject,proto3" json:"multipart_object,omitempty"`
	SatelliteSignature   []byte                   `protobuf:"bytes,9,opt,name=satellite_signature,json=satelliteSignature,proto3" json:"satellite_signature,omitempty"`
	StreamId             []byte                   `protobuf:"bytes,10,opt,name=stream_id,json=streamId,proto3" json:"stream_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *StreamID) Reset()         { *m = StreamID{} }
func (m *StreamID) String() string { return proto.CompactTextString(m) }
func (*StreamID) ProtoMessage()    {}
func (*StreamID) Descriptor() ([]byte, []int) {
	return fileDescriptor_47c60bd892d94aaf, []int{0}
}
func (m *StreamID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StreamID.Unmarshal(m, b)
}
func (m *StreamID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StreamID.Marshal(b, m, deterministic)
}
func (m *StreamID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StreamID.Merge(m, src)
}
func (m *StreamID) XXX_Size() int {
	return xxx_messageInfo_StreamID.Size(m)
}
func (m *StreamID) XXX_DiscardUnknown() {
	xxx_messageInfo_StreamID.DiscardUnknown(m)
}

var xxx_messageInfo_StreamID proto.InternalMessageInfo

func (m *StreamID) GetBucket() []byte {
	if m != nil {
		return m.Bucket
	}
	return nil
}

func (m *StreamID) GetEncryptedObjectKey() []byte {
	if m != nil {
		return m.EncryptedObjectKey
	}
	return nil
}

func (m *StreamID) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *StreamID) GetEncryptionParameters() *pb.EncryptionParameters {
	if m != nil {
		return m.EncryptionParameters
	}
	return nil
}

func (m *StreamID) GetCreationDate() time.Time {
	if m != nil {
		return m.CreationDate
	}
	return time.Time{}
}

func (m *StreamID) GetExpirationDate() time.Time {
	if m != nil {
		return m.ExpirationDate
	}
	return time.Time{}
}

func (m *StreamID) GetMultipartObject() bool {
	if m != nil {
		return m.MultipartObject
	}
	return false
}

func (m *StreamID) GetSatelliteSignature() []byte {
	if m != nil {
		return m.SatelliteSignature
	}
	return nil
}

func (m *StreamID) GetStreamId() []byte {
	if m != nil {
		return m.StreamId
	}
	return nil
}

type SegmentID struct {
	StreamId             *StreamID                 `protobuf:"bytes,1,opt,name=stream_id,json=streamId,proto3" json:"stream_id,omitempty"`
	PartNumber           int32                     `protobuf:"varint,2,opt,name=part_number,json=partNumber,proto3" json:"part_number,omitempty"`
	Index                int32                     `protobuf:"varint,3,opt,name=index,proto3" json:"index,omitempty"`
	RootPieceId          PieceID                   `protobuf:"bytes,5,opt,name=root_piece_id,json=rootPieceId,proto3,customtype=PieceID" json:"root_piece_id"`
	OriginalOrderLimits  []*pb.AddressedOrderLimit `protobuf:"bytes,6,rep,name=original_order_limits,json=originalOrderLimits,proto3" json:"original_order_limits,omitempty"`
	CreationDate         time.Time                 `protobuf:"bytes,7,opt,name=creation_date,json=creationDate,proto3,stdtime" json:"creation_date"`
	SatelliteSignature   []byte                    `protobuf:"bytes,8,opt,name=satellite_signature,json=satelliteSignature,proto3" json:"satellite_signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *SegmentID) Reset()         { *m = SegmentID{} }
func (m *SegmentID) String() string { return proto.CompactTextString(m) }
func (*SegmentID) ProtoMessage()    {}
func (*SegmentID) Descriptor() ([]byte, []int) {
	return fileDescriptor_47c60bd892d94aaf, []int{1}
}
func (m *SegmentID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SegmentID.Unmarshal(m, b)
}
func (m *SegmentID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SegmentID.Marshal(b, m, deterministic)
}
func (m *SegmentID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SegmentID.Merge(m, src)
}
func (m *SegmentID) XXX_Size() int {
	return xxx_messageInfo_SegmentID.Size(m)
}
func (m *SegmentID) XXX_DiscardUnknown() {
	xxx_messageInfo_SegmentID.DiscardUnknown(m)
}

var xxx_messageInfo_SegmentID proto.InternalMessageInfo

func (m *SegmentID) GetStreamId() *StreamID {
	if m != nil {
		return m.StreamId
	}
	return nil
}

func (m *SegmentID) GetPartNumber() int32 {
	if m != nil {
		return m.PartNumber
	}
	return 0
}

func (m *SegmentID) GetIndex() int32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *SegmentID) GetOriginalOrderLimits() []*pb.AddressedOrderLimit {
	if m != nil {
		return m.OriginalOrderLimits
	}
	return nil
}

func (m *SegmentID) GetCreationDate() time.Time {
	if m != nil {
		return m.CreationDate
	}
	return time.Time{}
}

func (m *SegmentID) GetSatelliteSignature() []byte {
	if m != nil {
		return m.SatelliteSignature
	}
	return nil
}

func init() {
	proto.RegisterType((*StreamID)(nil), "satellite.metainfo.StreamID")
	proto.RegisterType((*SegmentID)(nil), "satellite.metainfo.SegmentID")
}

func init() { proto.RegisterFile("metainfo_sat.proto", fileDescriptor_47c60bd892d94aaf) }

var fileDescriptor_47c60bd892d94aaf = []byte{
	// 530 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0x4d, 0x6f, 0xd3, 0x40,
	0x10, 0xc5, 0x44, 0x49, 0x93, 0x4d, 0xda, 0x54, 0xdb, 0x14, 0x59, 0x01, 0x14, 0xab, 0x08, 0x29,
	0x5c, 0x6c, 0xd4, 0x9e, 0x38, 0x12, 0x85, 0x43, 0xc4, 0x47, 0x8b, 0x03, 0x17, 0x2e, 0xd6, 0x3a,
	0x3b, 0xb5, 0xb6, 0xb5, 0x77, 0xad, 0xdd, 0x09, 0x6a, 0x8e, 0xfc, 0x03, 0x7e, 0x16, 0x3f, 0x01,
	0x71, 0x28, 0x7f, 0x05, 0x79, 0xfd, 0x91, 0x48, 0xb4, 0x07, 0xb8, 0xed, 0xbc, 0x79, 0xfb, 0x76,
	0xe7, 0xbd, 0x21, 0x34, 0x03, 0x64, 0x42, 0x5e, 0xaa, 0xc8, 0x30, 0xf4, 0x73, 0xad, 0x50, 0x51,
	0x6a, 0x18, 0x42, 0x9a, 0x0a, 0x04, 0xbf, 0xee, 0x8e, 0x0f, 0x41, 0xae, 0xf4, 0x26, 0x47, 0xa1,
	0x64, 0xc9, 0x1a, 0x93, 0x44, 0x25, 0xaa, 0x3a, 0x4f, 0x12, 0xa5, 0x92, 0x14, 0x02, 0x5b, 0xc5,
	0xeb, 0xcb, 0x00, 0x45, 0x06, 0x06, 0x59, 0x96, 0x57, 0x84, 0x83, 0x5a, 0xa8, 0xac, 0x4f, 0x7e,
	0xb6, 0x48, 0x77, 0x89, 0x1a, 0x58, 0xb6, 0x98, 0xd3, 0x47, 0xa4, 0x13, 0xaf, 0x57, 0xd7, 0x80,
	0xae, 0xe3, 0x39, 0xd3, 0x41, 0x58, 0x55, 0xf4, 0x25, 0x19, 0x55, 0xaf, 0x02, 0x8f, 0x54, 0x7c,
	0x05, 0x2b, 0x8c, 0xae, 0x61, 0xe3, 0x3e, 0xb4, 0x2c, 0xda, 0xf4, 0xce, 0x6d, 0xeb, 0x2d, 0x6c,
	0xa8, 0x4b, 0xf6, 0xbe, 0x82, 0x36, 0x42, 0x49, 0xb7, 0xe5, 0x39, 0xd3, 0x76, 0x58, 0x97, 0xf4,
	0x33, 0x39, 0xde, 0x4e, 0x10, 0xe5, 0x4c, 0xb3, 0x0c, 0x10, 0xb4, 0x71, 0x07, 0x9e, 0x33, 0xed,
	0x9f, 0x7a, 0xfe, 0xce, 0x7c, 0x6f, 0x9a, 0xe3, 0x45, 0xc3, 0x0b, 0x47, 0x70, 0x07, 0x4a, 0x17,
	0x64, 0x7f, 0xa5, 0x81, 0x59, 0x51, 0xce, 0x10, 0xdc, 0xb6, 0x95, 0x1b, 0xfb, 0xa5, 0x21, 0x7e,
	0x6d, 0x88, 0xff, 0xa9, 0x36, 0x64, 0xd6, 0xfd, 0x71, 0x3b, 0x79, 0xf0, 0xfd, 0xf7, 0xc4, 0x09,
	0x07, 0xf5, 0xd5, 0x39, 0x43, 0xa0, 0xef, 0xc9, 0x10, 0x6e, 0x72, 0xa1, 0x77, 0xc4, 0x3a, 0xff,
	0x20, 0x76, 0xb0, 0xbd, 0x6c, 0xe5, 0x5e, 0x90, 0xc3, 0x6c, 0x9d, 0xa2, 0xc8, 0x99, 0xc6, 0xca,
	0x3c, 0xb7, 0xef, 0x39, 0xd3, 0x6e, 0x38, 0x6c, 0xf0, 0xd2, 0x38, 0x1a, 0x90, 0xa3, 0x26, 0xf1,
	0xc8, 0x88, 0x44, 0x32, 0x5c, 0x6b, 0x70, 0x7b, 0xa5, 0xcd, 0x4d, 0x6b, 0x59, 0x77, 0xe8, 0x63,
	0xd2, 0x33, 0x36, 0xbc, 0x48, 0x70, 0x97, 0x58, 0x5a, 0xb7, 0x04, 0x16, 0xfc, 0xe4, 0x5b, 0x8b,
	0xf4, 0x96, 0x90, 0x64, 0x20, 0x71, 0x31, 0xa7, 0xaf, 0x76, 0xa9, 0x8e, 0x9d, 0xe7, 0x89, 0xff,
	0xf7, 0x7e, 0xf9, 0xf5, 0x32, 0x6c, 0x85, 0xe8, 0x84, 0xf4, 0xed, 0xe7, 0xe5, 0x3a, 0x8b, 0x41,
	0xdb, 0xd4, 0xdb, 0x21, 0x29, 0xa0, 0x0f, 0x16, 0xa1, 0x23, 0xd2, 0x16, 0x92, 0xc3, 0x4d, 0x95,
	0x75, 0x59, 0xd0, 0x33, 0xb2, 0xaf, 0x95, 0xc2, 0x28, 0x17, 0xb0, 0x82, 0xe2, 0xd5, 0x22, 0x92,
	0xc1, 0x6c, 0x58, 0x38, 0xf5, 0xeb, 0x76, 0xb2, 0x77, 0x51, 0xe0, 0x8b, 0x79, 0xd8, 0x2f, 0x58,
	0x65, 0xc1, 0xe9, 0x47, 0x72, 0xac, 0xb4, 0x48, 0x84, 0x64, 0x69, 0xa4, 0x34, 0x07, 0x1d, 0xa5,
	0x22, 0x13, 0x68, 0xdc, 0x8e, 0xd7, 0x9a, 0xf6, 0x4f, 0x9f, 0x6e, 0x3f, 0xfa, 0x9a, 0x73, 0x0d,
	0xc6, 0x00, 0x3f, 0x2f, 0x68, 0xef, 0x0a, 0x56, 0x78, 0x54, 0xdf, 0xdd, 0x62, 0x77, 0xac, 0xc6,
	0xde, 0x7f, 0xaf, 0xc6, 0x3d, 0x01, 0x75, 0xef, 0x0b, 0x68, 0xf6, 0xfc, 0xcb, 0x33, 0x83, 0x4a,
	0x5f, 0xf9, 0x42, 0x05, 0xf6, 0x10, 0x34, 0xa4, 0x40, 0x48, 0x04, 0x2d, 0x59, 0x9a, 0xc7, 0x71,
	0xc7, 0xfe, 0xe1, 0xec, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x02, 0x47, 0xce, 0x91, 0x05, 0x04,
	0x00, 0x00,
}
