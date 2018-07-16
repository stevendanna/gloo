// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: spec.proto

/*
Package rest is a generated protocol buffer package.

It is generated from these files:
	spec.proto

It has these top-level messages:
	RouteExtension
	Parameters
	TransformationSpec
*/
package rest

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import google_protobuf1 "github.com/gogo/protobuf/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// The REST Route Extension contains two components:
// * parameters for calling REST functions
// * Response Transformation
type RouteExtension struct {
	// If specified, these parameters will be used as inputs for REST templates for
	// the destination function for the route
	// (if the route destination is a functional destination that has a REST transformation)
	Parameters *Parameters `protobuf:"bytes,1,opt,name=parameters" json:"parameters,omitempty"`
	// If specified, responses on this route will be transformed according to the template(s) provided
	// in the transformation spec here
	ResponseTransformation *TransformationSpec `protobuf:"bytes,2,opt,name=response_transformation,json=responseTransformation" json:"response_transformation,omitempty"`
	// If specified, paremeters for the response transformation will be extracted from these sources
	ResponseParams *Parameters `protobuf:"bytes,3,opt,name=response_params,json=responseParams" json:"response_params,omitempty"`
}

func (m *RouteExtension) Reset()                    { *m = RouteExtension{} }
func (m *RouteExtension) String() string            { return proto.CompactTextString(m) }
func (*RouteExtension) ProtoMessage()               {}
func (*RouteExtension) Descriptor() ([]byte, []int) { return fileDescriptorSpec, []int{0} }

func (m *RouteExtension) GetParameters() *Parameters {
	if m != nil {
		return m.Parameters
	}
	return nil
}

func (m *RouteExtension) GetResponseTransformation() *TransformationSpec {
	if m != nil {
		return m.ResponseTransformation
	}
	return nil
}

func (m *RouteExtension) GetResponseParams() *Parameters {
	if m != nil {
		return m.ResponseParams
	}
	return nil
}

// Parameters define a set of parameters for REST Transformations
// Parameters can be extracted from HTTP Headers and Request Path
// Parameters can also be extracted from the HTTP Body, provided that it is
// valid JSON-encoded
// Gloo will search for parameters by their name in strings, enclosed in single
// curly braces, and attempt to match them to the variables in REST Function Templates
// for example:
//   # route
//   match: {...}
//   destination: {...}
//   extensions:
//     parameters:
//         headers:
//           x-user-id: { userId }
//   ---
//   # function
//   name: myfunc
//   spec:
//     body: |
//     {
//       "id": {{ userId }}
//     }
type Parameters struct {
	// headers that will be used to extract data for processing output templates
	// Gloo will search for parameters by their name in header value strings, enclosed in single
	// curly braces
	// Example:
	//   extensions:
	//     parameters:
	//         headers:
	//           x-user-id: { userId }
	Headers map[string]string `protobuf:"bytes,1,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// part of the (or the entire) path that will be used extract data for processing output templates
	// Gloo will search for parameters by their name in header value strings, enclosed in single
	// curly braces
	// Example:
	//   extensions:
	//     parameters:
	//         path: /users/{ userId }
	Path *google_protobuf1.StringValue `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
}

func (m *Parameters) Reset()                    { *m = Parameters{} }
func (m *Parameters) String() string            { return proto.CompactTextString(m) }
func (*Parameters) ProtoMessage()               {}
func (*Parameters) Descriptor() ([]byte, []int) { return fileDescriptorSpec, []int{1} }

func (m *Parameters) GetHeaders() map[string]string {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *Parameters) GetPath() *google_protobuf1.StringValue {
	if m != nil {
		return m.Path
	}
	return nil
}

// TransformationSpec can act as part of a Route Extension (as a Response Transformation), or as
// a FunctionSpec (as a Request Transformation).
// Use TransformationSpec as the Function Spec for REST Services (where `Upstream.ServiceInfo.Type == "REST"`)
// TransformationSpec contains a set of templates that will be used to modify the Path, Headers, and Body
// Parameters for the tempalte come from the following sources:
// path: HTTP Request path (if present)
// method: HTTP Request method (if present)
// parameters specified in the RouteExtension.Parameters (or, in the case of ResponseTransformation, RouteExtension.ResponseParams)
// Parameters can also be extracted from the Request / Response Body provided that they are JSON
// To do so, specify the field using JSONPath syntax
// any field from the request body, assuming it's json (http://goessner.net/articles/JsonPath/index.html#e2)
type TransformationSpec struct {
	// a Jinja-style Template string for the outbound request path. Only useful for request transformation
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// a map of keys to Jinja-style Template strings HTTP Headers. Useful for request and response transformations
	Headers map[string]string `protobuf:"bytes,2,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// a Jinja-style Template string for the outbound HTTP Body. Useful for request and response transformations
	// If this is nil, the body will be passed through unmodified. If set to an empty string, the body will be removed
	// from the HTTP message.
	Body *google_protobuf1.StringValue `protobuf:"bytes,3,opt,name=body" json:"body,omitempty"`
}

func (m *TransformationSpec) Reset()                    { *m = TransformationSpec{} }
func (m *TransformationSpec) String() string            { return proto.CompactTextString(m) }
func (*TransformationSpec) ProtoMessage()               {}
func (*TransformationSpec) Descriptor() ([]byte, []int) { return fileDescriptorSpec, []int{2} }

func (m *TransformationSpec) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *TransformationSpec) GetHeaders() map[string]string {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *TransformationSpec) GetBody() *google_protobuf1.StringValue {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*RouteExtension)(nil), "gloo.api.rest.v1.RouteExtension")
	proto.RegisterType((*Parameters)(nil), "gloo.api.rest.v1.Parameters")
	proto.RegisterType((*TransformationSpec)(nil), "gloo.api.rest.v1.TransformationSpec")
}
func (this *RouteExtension) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RouteExtension)
	if !ok {
		that2, ok := that.(RouteExtension)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Parameters.Equal(that1.Parameters) {
		return false
	}
	if !this.ResponseTransformation.Equal(that1.ResponseTransformation) {
		return false
	}
	if !this.ResponseParams.Equal(that1.ResponseParams) {
		return false
	}
	return true
}
func (this *Parameters) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Parameters)
	if !ok {
		that2, ok := that.(Parameters)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.Headers) != len(that1.Headers) {
		return false
	}
	for i := range this.Headers {
		if this.Headers[i] != that1.Headers[i] {
			return false
		}
	}
	if !this.Path.Equal(that1.Path) {
		return false
	}
	return true
}
func (this *TransformationSpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*TransformationSpec)
	if !ok {
		that2, ok := that.(TransformationSpec)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Path != that1.Path {
		return false
	}
	if len(this.Headers) != len(that1.Headers) {
		return false
	}
	for i := range this.Headers {
		if this.Headers[i] != that1.Headers[i] {
			return false
		}
	}
	if !this.Body.Equal(that1.Body) {
		return false
	}
	return true
}

func init() { proto.RegisterFile("spec.proto", fileDescriptorSpec) }

var fileDescriptorSpec = []byte{
	// 392 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x91, 0xcf, 0x6a, 0x2a, 0x31,
	0x14, 0xc6, 0x19, 0xbd, 0x7f, 0x30, 0x5e, 0xbc, 0x12, 0xe4, 0x5e, 0x11, 0x91, 0x8b, 0xdc, 0x85,
	0x5d, 0x34, 0xa9, 0x76, 0x53, 0xa4, 0xab, 0x16, 0xa1, 0xd0, 0x4d, 0x19, 0x4b, 0x17, 0x85, 0x52,
	0x32, 0x1a, 0xe3, 0xe0, 0x98, 0x13, 0x92, 0x8c, 0xad, 0x6f, 0xd4, 0xb7, 0xe8, 0xbb, 0xb4, 0xaf,
	0xd0, 0x07, 0x28, 0x93, 0x71, 0xfc, 0x53, 0xa9, 0x08, 0xdd, 0x9d, 0xcc, 0xf9, 0xce, 0x77, 0xce,
	0xef, 0x1b, 0x84, 0x8c, 0xe2, 0x03, 0xa2, 0x34, 0x58, 0xc0, 0x65, 0x11, 0x01, 0x10, 0xa6, 0x42,
	0xa2, 0xb9, 0xb1, 0x64, 0xd6, 0xae, 0x55, 0x04, 0x08, 0x70, 0x4d, 0x9a, 0x54, 0xa9, 0xae, 0xd6,
	0x10, 0x00, 0x22, 0xe2, 0xd4, 0xbd, 0x82, 0x78, 0x44, 0x1f, 0x34, 0x53, 0x8a, 0x6b, 0x93, 0xf6,
	0x9b, 0x6f, 0x1e, 0x2a, 0xf9, 0x10, 0x5b, 0xde, 0x7b, 0xb4, 0x5c, 0x9a, 0x10, 0x24, 0x3e, 0x45,
	0x48, 0x31, 0xcd, 0xa6, 0xdc, 0x72, 0x6d, 0xaa, 0xde, 0x3f, 0xaf, 0x55, 0xec, 0xd4, 0xc9, 0xc7,
	0x7d, 0xe4, 0x6a, 0xa9, 0xf1, 0xd7, 0xf4, 0xf8, 0x0e, 0xfd, 0xd5, 0xdc, 0x28, 0x90, 0x86, 0xdf,
	0x5b, 0xcd, 0xa4, 0x19, 0x81, 0x9e, 0x32, 0x1b, 0x82, 0xac, 0xe6, 0x9c, 0xd5, 0xff, 0x6d, 0xab,
	0xeb, 0x0d, 0x5d, 0x5f, 0xf1, 0x81, 0xff, 0x27, 0x33, 0xd9, 0xec, 0xe1, 0x1e, 0xfa, 0xbd, 0xb4,
	0x77, 0x5b, 0x4d, 0x35, 0xbf, 0xc7, 0x85, 0xa5, 0x6c, 0xc8, 0x7d, 0x33, 0xcd, 0x67, 0x0f, 0xa1,
	0x55, 0x1b, 0x9f, 0xa3, 0x9f, 0x63, 0xce, 0x86, 0x29, 0x6f, 0xbe, 0x55, 0xec, 0x1c, 0xec, 0x72,
	0x23, 0x17, 0xa9, 0xb6, 0x27, 0xad, 0x9e, 0xfb, 0xd9, 0x24, 0x3e, 0x42, 0xdf, 0x14, 0xb3, 0xe3,
	0x05, 0x66, 0x9d, 0xa4, 0xc9, 0x93, 0x2c, 0x79, 0xd2, 0xb7, 0x3a, 0x94, 0xe2, 0x86, 0x45, 0x31,
	0xf7, 0x9d, 0xb2, 0xd6, 0x45, 0xbf, 0xd6, 0xad, 0x70, 0x19, 0xe5, 0x27, 0x7c, 0xee, 0x22, 0x2f,
	0xf8, 0x49, 0x89, 0x2b, 0xe8, 0xfb, 0x2c, 0x19, 0x70, 0xa6, 0x05, 0x3f, 0x7d, 0x74, 0x73, 0x27,
	0x5e, 0xf3, 0xd5, 0x43, 0x78, 0x3b, 0x37, 0x8c, 0x17, 0x47, 0xa4, 0x1e, 0xae, 0xc6, 0x97, 0x2b,
	0xba, 0x9c, 0xa3, 0x6b, 0xef, 0xf3, 0x0b, 0x3e, 0xa7, 0x0c, 0x60, 0x38, 0x5f, 0xa5, 0xbe, 0x8b,
	0x32, 0x51, 0x7e, 0x85, 0xf2, 0x8c, 0x3c, 0xbd, 0x34, 0xbc, 0xdb, 0x96, 0x08, 0xed, 0x38, 0x0e,
	0xc8, 0x00, 0xa6, 0xd4, 0x40, 0x04, 0x87, 0x21, 0xd0, 0x84, 0x80, 0xaa, 0x89, 0xa0, 0x2a, 0x8a,
	0x45, 0x28, 0x0d, 0x4d, 0x48, 0x82, 0x1f, 0xee, 0x8e, 0xe3, 0xf7, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x4b, 0xe7, 0xb0, 0x17, 0x2b, 0x03, 0x00, 0x00,
}