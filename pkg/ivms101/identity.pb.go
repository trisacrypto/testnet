// Specification Copyright (c) 2020 Joint Working Group on interVASP Messaging Standards
// https://intervasp.org/
// https://intervasp.org/wp-content/uploads/2020/05/IVMS101-interVASP-data-model-standard-issue-1-FINAL.pdf

// Protocol Buffer Specification Copyright (c) 2020 CipherTrace, Inc. https://ciphertrace.com

// Licensed under MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// NOTE ON THE SPECIFICATION MAPPING
// This protocol buffers specification has applied the Protocol Buffers style guide
// https://developers.google.com/protocol-buffers/docs/style to the ISVM101
// specification to be consistent with other Protocol Buffers specifications and to
// avoid common pitfalls when generating language specific classes.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: ivms101/identity.proto

package ivms101

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Originator struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the account holder who allows the VA transfer from that account or,
	// where there is no account, the natural or legal person that places the order with
	// the originating VASP to perform the VA transfer.
	// One or More
	OriginatorPersons []*Person `protobuf:"bytes,1,rep,name=originator_persons,json=originatorPersons,proto3" json:"originator_persons,omitempty"`
	// Definition: Identifier of an account that is used to process the transaction.
	// The value for this element is case-sensitive.
	// Datatype: “Max100Text”
	// Zero or More
	AccountNumbers []string `protobuf:"bytes,2,rep,name=account_numbers,json=accountNumbers,proto3" json:"account_numbers,omitempty"`
}

func (x *Originator) Reset() {
	*x = Originator{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Originator) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Originator) ProtoMessage() {}

func (x *Originator) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Originator.ProtoReflect.Descriptor instead.
func (*Originator) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{0}
}

func (x *Originator) GetOriginatorPersons() []*Person {
	if x != nil {
		return x.OriginatorPersons
	}
	return nil
}

func (x *Originator) GetAccountNumbers() []string {
	if x != nil {
		return x.AccountNumbers
	}
	return nil
}

type Beneficiary struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the natural or legal person or legal arrangement who is identified
	// by the originator as the receiver of the requested VA transfer.
	// One or More
	BeneficiaryPersons []*Person `protobuf:"bytes,1,rep,name=beneficiary_persons,json=beneficiaryPersons,proto3" json:"beneficiary_persons,omitempty"`
	// Definition: Identifier of an account that is used to process the transaction.
	// The value for this element is case-sensitive.
	// Datatype: “Max100Text”
	// Zero or More
	AccountNumbers []string `protobuf:"bytes,2,rep,name=account_numbers,json=accountNumbers,proto3" json:"account_numbers,omitempty"`
}

func (x *Beneficiary) Reset() {
	*x = Beneficiary{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Beneficiary) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Beneficiary) ProtoMessage() {}

func (x *Beneficiary) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Beneficiary.ProtoReflect.Descriptor instead.
func (*Beneficiary) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{1}
}

func (x *Beneficiary) GetBeneficiaryPersons() []*Person {
	if x != nil {
		return x.BeneficiaryPersons
	}
	return nil
}

func (x *Beneficiary) GetAccountNumbers() []string {
	if x != nil {
		return x.AccountNumbers
	}
	return nil
}

type OriginatingVasp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: refers to the VASP which initiates the VA transfer, and transfers
	// the VA upon receiving the request for a VA transfer on behalf of the originator.
	// Optional
	OriginatingVasp *Person `protobuf:"bytes,1,opt,name=originating_vasp,json=originatingVasp,proto3" json:"originating_vasp,omitempty"`
}

func (x *OriginatingVasp) Reset() {
	*x = OriginatingVasp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OriginatingVasp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OriginatingVasp) ProtoMessage() {}

func (x *OriginatingVasp) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OriginatingVasp.ProtoReflect.Descriptor instead.
func (*OriginatingVasp) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{2}
}

func (x *OriginatingVasp) GetOriginatingVasp() *Person {
	if x != nil {
		return x.OriginatingVasp
	}
	return nil
}

type BeneficiaryVasp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the VASP which receives the transfer of a virtual asset from the
	// originating VASP directly or through an intermediary VASP and makes the funds
	// available to the beneficiary.
	// Optional
	BeneficiaryVasp *Person `protobuf:"bytes,1,opt,name=beneficiary_vasp,json=beneficiaryVasp,proto3" json:"beneficiary_vasp,omitempty"`
}

func (x *BeneficiaryVasp) Reset() {
	*x = BeneficiaryVasp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BeneficiaryVasp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BeneficiaryVasp) ProtoMessage() {}

func (x *BeneficiaryVasp) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BeneficiaryVasp.ProtoReflect.Descriptor instead.
func (*BeneficiaryVasp) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{3}
}

func (x *BeneficiaryVasp) GetBeneficiaryVasp() *Person {
	if x != nil {
		return x.BeneficiaryVasp
	}
	return nil
}

type IntermediaryVasp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the VASP in a serial chain that receives and retransmits a VA
	// transfer on behalf of the originating VASP and the beneficiary VASP, or another
	// intermediary VASP.
	// Required
	IntermediaryVasp *Person `protobuf:"bytes,1,opt,name=intermediary_vasp,json=intermediaryVasp,proto3" json:"intermediary_vasp,omitempty"`
	// Definition: the sequence in a serial chain at which the corresponding
	// intermediary VASP participates in the transfer.
	// Constraints: totalDigits: 18; fractionDigits: 0
	// Required
	Sequence uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
}

func (x *IntermediaryVasp) Reset() {
	*x = IntermediaryVasp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IntermediaryVasp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IntermediaryVasp) ProtoMessage() {}

func (x *IntermediaryVasp) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IntermediaryVasp.ProtoReflect.Descriptor instead.
func (*IntermediaryVasp) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{4}
}

func (x *IntermediaryVasp) GetIntermediaryVasp() *Person {
	if x != nil {
		return x.IntermediaryVasp
	}
	return nil
}

func (x *IntermediaryVasp) GetSequence() uint64 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

type TransferPath struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the intermediary VASP(s) participating in a serial chain that
	// receive and retransmit a VA transfer on behalf of the originating VASP and the
	// beneficiary VASP, or another intermediary VASP, together with their
	// corresponding sequence number.
	// Zero or More
	TransferPath []*IntermediaryVasp `protobuf:"bytes,1,rep,name=transfer_path,json=transferPath,proto3" json:"transfer_path,omitempty"`
}

func (x *TransferPath) Reset() {
	*x = TransferPath{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransferPath) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferPath) ProtoMessage() {}

func (x *TransferPath) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransferPath.ProtoReflect.Descriptor instead.
func (*TransferPath) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{5}
}

func (x *TransferPath) GetTransferPath() []*IntermediaryVasp {
	if x != nil {
		return x.TransferPath
	}
	return nil
}

type PayloadMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Definition: the method used to map from a national system of writing to Latin script.
	// Zero or More
	TransliterationMethod []TransliterationMethodCode `protobuf:"varint,1,rep,packed,name=transliteration_method,json=transliterationMethod,proto3,enum=ivms101.TransliterationMethodCode" json:"transliteration_method,omitempty"`
}

func (x *PayloadMetadata) Reset() {
	*x = PayloadMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PayloadMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PayloadMetadata) ProtoMessage() {}

func (x *PayloadMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PayloadMetadata.ProtoReflect.Descriptor instead.
func (*PayloadMetadata) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{6}
}

func (x *PayloadMetadata) GetTransliterationMethod() []TransliterationMethodCode {
	if x != nil {
		return x.TransliterationMethod
	}
	return nil
}

type IdentityPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Originator      *Originator      `protobuf:"bytes,1,opt,name=originator,proto3" json:"originator,omitempty"`
	Beneficiary     *Beneficiary     `protobuf:"bytes,2,opt,name=beneficiary,proto3" json:"beneficiary,omitempty"`
	OriginatingVasp *OriginatingVasp `protobuf:"bytes,3,opt,name=originating_vasp,json=originatingVasp,proto3" json:"originating_vasp,omitempty"`
	BeneficiaryVasp *BeneficiaryVasp `protobuf:"bytes,4,opt,name=beneficiary_vasp,json=beneficiaryVasp,proto3" json:"beneficiary_vasp,omitempty"`
	TransferPath    *TransferPath    `protobuf:"bytes,5,opt,name=transfer_path,json=transferPath,proto3" json:"transfer_path,omitempty"`
	PayloadMetadata *PayloadMetadata `protobuf:"bytes,6,opt,name=payload_metadata,json=payloadMetadata,proto3" json:"payload_metadata,omitempty"`
}

func (x *IdentityPayload) Reset() {
	*x = IdentityPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ivms101_identity_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IdentityPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdentityPayload) ProtoMessage() {}

func (x *IdentityPayload) ProtoReflect() protoreflect.Message {
	mi := &file_ivms101_identity_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdentityPayload.ProtoReflect.Descriptor instead.
func (*IdentityPayload) Descriptor() ([]byte, []int) {
	return file_ivms101_identity_proto_rawDescGZIP(), []int{7}
}

func (x *IdentityPayload) GetOriginator() *Originator {
	if x != nil {
		return x.Originator
	}
	return nil
}

func (x *IdentityPayload) GetBeneficiary() *Beneficiary {
	if x != nil {
		return x.Beneficiary
	}
	return nil
}

func (x *IdentityPayload) GetOriginatingVasp() *OriginatingVasp {
	if x != nil {
		return x.OriginatingVasp
	}
	return nil
}

func (x *IdentityPayload) GetBeneficiaryVasp() *BeneficiaryVasp {
	if x != nil {
		return x.BeneficiaryVasp
	}
	return nil
}

func (x *IdentityPayload) GetTransferPath() *TransferPath {
	if x != nil {
		return x.TransferPath
	}
	return nil
}

func (x *IdentityPayload) GetPayloadMetadata() *PayloadMetadata {
	if x != nil {
		return x.PayloadMetadata
	}
	return nil
}

var File_ivms101_identity_proto protoreflect.FileDescriptor

var file_ivms101_identity_proto_rawDesc = []byte{
	0x0a, 0x16, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30,
	0x31, 0x1a, 0x15, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2f, 0x69, 0x76, 0x6d, 0x73, 0x31,
	0x30, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x12, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30,
	0x31, 0x2f, 0x65, 0x6e, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x75, 0x0a, 0x0a,
	0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x3e, 0x0a, 0x12, 0x6f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31,
	0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52, 0x11, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61,
	0x74, 0x6f, 0x72, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x73, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x0e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x22, 0x78, 0x0a, 0x0b, 0x42, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61,
	0x72, 0x79, 0x12, 0x40, 0x0a, 0x13, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72,
	0x79, 0x5f, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e,
	0x52, 0x12, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x50, 0x65, 0x72,
	0x73, 0x6f, 0x6e, 0x73, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f,
	0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x22, 0x4d, 0x0a,
	0x0f, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x73, 0x70,
	0x12, 0x3a, 0x0a, 0x10, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f,
	0x76, 0x61, 0x73, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x69, 0x76, 0x6d,
	0x73, 0x31, 0x30, 0x31, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52, 0x0f, 0x6f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x73, 0x70, 0x22, 0x4d, 0x0a, 0x0f,
	0x42, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73, 0x70, 0x12,
	0x3a, 0x0a, 0x10, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x5f, 0x76,
	0x61, 0x73, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x69, 0x76, 0x6d, 0x73,
	0x31, 0x30, 0x31, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52, 0x0f, 0x62, 0x65, 0x6e, 0x65,
	0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73, 0x70, 0x22, 0x6c, 0x0a, 0x10, 0x49,
	0x6e, 0x74, 0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73, 0x70, 0x12,
	0x3c, 0x0a, 0x11, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x72, 0x79, 0x5f,
	0x76, 0x61, 0x73, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x69, 0x76, 0x6d,
	0x73, 0x31, 0x30, 0x31, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52, 0x10, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73, 0x70, 0x12, 0x1a, 0x0a,
	0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x4e, 0x0a, 0x0c, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x65, 0x72, 0x50, 0x61, 0x74, 0x68, 0x12, 0x3e, 0x0a, 0x0d, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x65, 0x72, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x19, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73, 0x70, 0x52, 0x0c, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x65, 0x72, 0x50, 0x61, 0x74, 0x68, 0x22, 0x6c, 0x0a, 0x0f, 0x50, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x59, 0x0a, 0x16,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x69, 0x74, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x69,
	0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x69, 0x74, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x43, 0x6f, 0x64, 0x65,
	0x52, 0x15, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x69, 0x74, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x22, 0x89, 0x03, 0x0a, 0x0f, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x33, 0x0a, 0x0a, 0x6f,
	0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e,
	0x61, 0x74, 0x6f, 0x72, 0x52, 0x0a, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x6f, 0x72,
	0x12, 0x36, 0x0a, 0x0b, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e,
	0x42, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x52, 0x0b, 0x62, 0x65, 0x6e,
	0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x12, 0x43, 0x0a, 0x10, 0x6f, 0x72, 0x69, 0x67,
	0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61, 0x73, 0x70, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x18, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30, 0x31, 0x2e, 0x4f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x73, 0x70, 0x52, 0x0f, 0x6f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x73, 0x70, 0x12, 0x43, 0x0a,
	0x10, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x5f, 0x76, 0x61, 0x73,
	0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30,
	0x31, 0x2e, 0x42, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61, 0x73,
	0x70, 0x52, 0x0f, 0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x72, 0x79, 0x56, 0x61,
	0x73, 0x70, 0x12, 0x3a, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x5f, 0x70,
	0x61, 0x74, 0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x69, 0x76, 0x6d, 0x73,
	0x31, 0x30, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x50, 0x61, 0x74, 0x68,
	0x52, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x50, 0x61, 0x74, 0x68, 0x12, 0x43,
	0x0a, 0x10, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x69, 0x76, 0x6d, 0x73, 0x31,
	0x30, 0x31, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x52, 0x0f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x74, 0x72, 0x69, 0x73, 0x61, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x69, 0x76, 0x6d, 0x73, 0x31, 0x30,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ivms101_identity_proto_rawDescOnce sync.Once
	file_ivms101_identity_proto_rawDescData = file_ivms101_identity_proto_rawDesc
)

func file_ivms101_identity_proto_rawDescGZIP() []byte {
	file_ivms101_identity_proto_rawDescOnce.Do(func() {
		file_ivms101_identity_proto_rawDescData = protoimpl.X.CompressGZIP(file_ivms101_identity_proto_rawDescData)
	})
	return file_ivms101_identity_proto_rawDescData
}

var file_ivms101_identity_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_ivms101_identity_proto_goTypes = []interface{}{
	(*Originator)(nil),             // 0: ivms101.Originator
	(*Beneficiary)(nil),            // 1: ivms101.Beneficiary
	(*OriginatingVasp)(nil),        // 2: ivms101.OriginatingVasp
	(*BeneficiaryVasp)(nil),        // 3: ivms101.BeneficiaryVasp
	(*IntermediaryVasp)(nil),       // 4: ivms101.IntermediaryVasp
	(*TransferPath)(nil),           // 5: ivms101.TransferPath
	(*PayloadMetadata)(nil),        // 6: ivms101.PayloadMetadata
	(*IdentityPayload)(nil),        // 7: ivms101.IdentityPayload
	(*Person)(nil),                 // 8: ivms101.Person
	(TransliterationMethodCode)(0), // 9: ivms101.TransliterationMethodCode
}
var file_ivms101_identity_proto_depIdxs = []int32{
	8,  // 0: ivms101.Originator.originator_persons:type_name -> ivms101.Person
	8,  // 1: ivms101.Beneficiary.beneficiary_persons:type_name -> ivms101.Person
	8,  // 2: ivms101.OriginatingVasp.originating_vasp:type_name -> ivms101.Person
	8,  // 3: ivms101.BeneficiaryVasp.beneficiary_vasp:type_name -> ivms101.Person
	8,  // 4: ivms101.IntermediaryVasp.intermediary_vasp:type_name -> ivms101.Person
	4,  // 5: ivms101.TransferPath.transfer_path:type_name -> ivms101.IntermediaryVasp
	9,  // 6: ivms101.PayloadMetadata.transliteration_method:type_name -> ivms101.TransliterationMethodCode
	0,  // 7: ivms101.IdentityPayload.originator:type_name -> ivms101.Originator
	1,  // 8: ivms101.IdentityPayload.beneficiary:type_name -> ivms101.Beneficiary
	2,  // 9: ivms101.IdentityPayload.originating_vasp:type_name -> ivms101.OriginatingVasp
	3,  // 10: ivms101.IdentityPayload.beneficiary_vasp:type_name -> ivms101.BeneficiaryVasp
	5,  // 11: ivms101.IdentityPayload.transfer_path:type_name -> ivms101.TransferPath
	6,  // 12: ivms101.IdentityPayload.payload_metadata:type_name -> ivms101.PayloadMetadata
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_ivms101_identity_proto_init() }
func file_ivms101_identity_proto_init() {
	if File_ivms101_identity_proto != nil {
		return
	}
	file_ivms101_ivms101_proto_init()
	file_ivms101_enum_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ivms101_identity_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Originator); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Beneficiary); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OriginatingVasp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BeneficiaryVasp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IntermediaryVasp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransferPath); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PayloadMetadata); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ivms101_identity_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IdentityPayload); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ivms101_identity_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ivms101_identity_proto_goTypes,
		DependencyIndexes: file_ivms101_identity_proto_depIdxs,
		MessageInfos:      file_ivms101_identity_proto_msgTypes,
	}.Build()
	File_ivms101_identity_proto = out.File
	file_ivms101_identity_proto_rawDesc = nil
	file_ivms101_identity_proto_goTypes = nil
	file_ivms101_identity_proto_depIdxs = nil
}