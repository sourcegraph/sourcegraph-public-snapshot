// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: errorspb/testing.proto

package errorspb

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// TestError is meant for use in testing only.
type TestError struct {
}

func (m *TestError) Reset()         { *m = TestError{} }
func (m *TestError) String() string { return proto.CompactTextString(m) }
func (*TestError) ProtoMessage()    {}
func (*TestError) Descriptor() ([]byte, []int) {
	return fileDescriptor_0551f0d913d6118f, []int{0}
}
func (m *TestError) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TestError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *TestError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestError.Merge(m, src)
}
func (m *TestError) XXX_Size() int {
	return m.Size()
}
func (m *TestError) XXX_DiscardUnknown() {
	xxx_messageInfo_TestError.DiscardUnknown(m)
}

var xxx_messageInfo_TestError proto.InternalMessageInfo

func init() {
	proto.RegisterType((*TestError)(nil), "cockroach.errorspb.TestError")
}

func init() { proto.RegisterFile("errorspb/testing.proto", fileDescriptor_0551f0d913d6118f) }

var fileDescriptor_0551f0d913d6118f = []byte{
	// 118 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4b, 0x2d, 0x2a, 0xca,
	0x2f, 0x2a, 0x2e, 0x48, 0xd2, 0x2f, 0x49, 0x2d, 0x2e, 0xc9, 0xcc, 0x4b, 0xd7, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0x12, 0x4a, 0xce, 0x4f, 0xce, 0x2e, 0xca, 0x4f, 0x4c, 0xce, 0xd0, 0x83, 0xa9,
	0x50, 0xe2, 0xe6, 0xe2, 0x0c, 0x49, 0x2d, 0x2e, 0x71, 0x05, 0xf1, 0x9d, 0xb4, 0x4e, 0x3c, 0x94,
	0x63, 0x38, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x1b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c,
	0x92, 0x63, 0x9c, 0xf0, 0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2,
	0x38, 0x60, 0x1a, 0x93, 0xd8, 0xc0, 0x66, 0x1a, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x4b, 0xf3,
	0x3e, 0x92, 0x6d, 0x00, 0x00, 0x00,
}

func (m *TestError) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TestError) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TestError) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTesting(dAtA []byte, offset int, v uint64) int {
	offset -= sovTesting(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *TestError) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTesting(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTesting(x uint64) (n int) {
	return sovTesting(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TestError) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTesting
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TestError: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TestError: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTesting(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTesting
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTesting(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTesting
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTesting
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTesting
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTesting
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTesting
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTesting
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTesting        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTesting          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTesting = fmt.Errorf("proto: unexpected end of group")
)
