// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pb

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/google/cel-spec/proto/checked/v1/checked"
	"reflect"
	"strings"
)

// TypeDescription is a collection of type metadata relevant to expression
// checking and evaluation.
type TypeDescription struct {
	typeName        string
	file            *FileDescription
	desc            *descpb.DescriptorProto
	fields          map[string]*FieldDescription
	fieldIndices    map[int][]*FieldDescription
	fieldProperties *proto.StructProperties
	refType         *reflect.Type
}

// FieldCount returns the number of fields declared within the type.
func (td *TypeDescription) FieldCount() int {
	// Initialize the type's internal state.
	var _, fieldIndices = td.getFieldsInfo()
	// The number of keys in the field indices map corresponds to the number
	// of fields on the proto message.
	return len(fieldIndices)
}

// FieldByName returns the FieldDescription associated with a field name.
func (td *TypeDescription) FieldByName(name string) (*FieldDescription, bool) {
	fieldMap, _ := td.getFieldsInfo()
	fd, found := fieldMap[name]
	return fd, found
}

// FieldNameAtIndex returns the field name at the specified index.
//
// For oneof field values, multiple fields may exist at the same index, so the
// appropriate oneof getter's must be invoked in order to determine which of
// oneof fields is currently set at the index.
func (td *TypeDescription) FieldNameAtIndex(index int, refObj reflect.Value) (string, bool) {
	fields := td.getFieldsAtIndex(index)
	if len(fields) == 1 {
		return fields[0].OrigName(), true
	}
	for _, fd := range fields {
		if fd.IsOneof() {
			getter := refObj.MethodByName(fd.GetterName())
			if !getter.IsValid() {
				continue
			}
			refField := getter.Call([]reflect.Value{})[0]
			if refField.IsValid() && !refField.IsNil() {
				return fd.OrigName(), true
			}
		}
	}
	return "", false
}

// Name of the type.
func (td *TypeDescription) Name() string {
	return td.typeName
}

// ReflectType returns the reflected struct type of the generated proto struct.
func (td *TypeDescription) ReflectType() reflect.Type {
	if td.refType == nil {
		refType := proto.MessageType(td.Name())
		if refType == nil {
			return nil
		}
		td.refType = &refType
	}
	return *td.refType
}

func (td *TypeDescription) getFieldsInfo() (map[string]*FieldDescription,
	map[int][]*FieldDescription) {
	if len(td.fields) == 0 {
		isProto3 := td.file.desc.GetSyntax() == "proto3"
		fieldIndexMap := make(map[string]int)
		fieldDescMap := make(map[string]*descpb.FieldDescriptorProto)
		for i, f := range td.desc.Field {
			fieldDescMap[f.GetName()] = f
			fieldIndexMap[f.GetName()] = i
		}
		fieldProps := td.getFieldProperties()
		if fieldProps != nil {
			// This is a proper message type.
			for i, prop := range fieldProps.Prop {
				if strings.HasPrefix(prop.OrigName, "XXX_") {
					// Book-keeping fields generated by protoc start with XXX_
					continue
				}
				desc := fieldDescMap[prop.OrigName]
				fd := &FieldDescription{
					desc:   desc,
					index:  i,
					prop:   prop,
					proto3: isProto3}
				td.fields[prop.OrigName] = fd
				td.fieldIndices[i] = append(td.fieldIndices[i], fd)
			}
			for _, oneofProp := range fieldProps.OneofTypes {
				desc := fieldDescMap[oneofProp.Prop.OrigName]
				fd := &FieldDescription{
					desc:      desc,
					index:     oneofProp.Field,
					prop:      oneofProp.Prop,
					oneofProp: oneofProp,
					proto3:    isProto3}
				td.fields[oneofProp.Prop.OrigName] = fd
				td.fieldIndices[oneofProp.Field] = append(td.fieldIndices[oneofProp.Field], fd)
			}
		} else {
			for fieldName, desc := range fieldDescMap {
				fd := &FieldDescription{
					desc:   desc,
					index:  int(desc.GetNumber()),
					proto3: isProto3}
				td.fields[fieldName] = fd
				index := fieldIndexMap[fieldName]
				td.fieldIndices[index] = append(td.fieldIndices[index], fd)
			}
		}
	}
	return td.fields, td.fieldIndices
}

func (td *TypeDescription) getFieldProperties() *proto.StructProperties {
	if td.fieldProperties == nil {
		refType := td.ReflectType()
		if refType == nil {
			return nil
		}
		if refType.Kind() == reflect.Ptr {
			refType = refType.Elem()
		}
		if refType.Kind() == reflect.Struct {
			td.fieldProperties = proto.GetProperties(refType)
		}
	}
	return td.fieldProperties
}

func (td *TypeDescription) getFieldsAtIndex(i int) []*FieldDescription {
	_, fieldIndicies := td.getFieldsInfo()
	return fieldIndicies[i]
}

// FieldDescription holds metadata related to fields declared within a type.
type FieldDescription struct {
	desc      *descpb.FieldDescriptorProto
	index     int
	prop      *proto.Properties
	oneofProp *proto.OneofProperties
	proto3    bool
}

// CheckedType returns the type-definition used at type-check time.
func (fd *FieldDescription) CheckedType() *checked.Type {
	if fd.IsMap() {
		td, _ := DescribeType(fd.TypeName())
		key := td.getFieldsAtIndex(0)[0]
		val := td.getFieldsAtIndex(1)[0]
		return &checked.Type{
			TypeKind: &checked.Type_MapType_{
				MapType: &checked.Type_MapType{
					KeyType:   key.typeDefToType(),
					ValueType: val.typeDefToType()}}}
	}
	if fd.IsRepeated() {
		return &checked.Type{
			TypeKind: &checked.Type_ListType_{
				ListType: &checked.Type_ListType{
					ElemType: fd.typeDefToType()}}}
	}
	return fd.typeDefToType()
}

// GetterName returns the accessor method name associated with the field
// on the proto generated struct.
func (fd *FieldDescription) GetterName() string {
	return fmt.Sprintf("Get%s", fd.prop.Name)
}

// Index returns the field index within a reflected value.
func (fd *FieldDescription) Index() int {
	return fd.index
}

// IsEnum returns true if the field type refers to an enum value.
func (fd *FieldDescription) IsEnum() bool {
	return fd.desc.GetType() == descpb.FieldDescriptorProto_TYPE_ENUM
}

// IsOneof returns true if the field is declared within a oneof block.
func (fd *FieldDescription) IsOneof() bool {
	return fd.oneofProp != nil
}

// OneofType returns the reflect.Type value of a oneof field.
//
// Oneof field values are wrapped in a struct which contains one field whose
// value is a proto.Message.
func (fd *FieldDescription) OneofType() reflect.Type {
	return fd.oneofProp.Type
}

// IsMap returns true if the field is of map type.
func (fd *FieldDescription) IsMap() bool {
	if !fd.IsRepeated() || !fd.IsMessage() {
		return false
	}
	td, err := DescribeType(fd.TypeName())
	if err != nil {
		return false
	}
	return td.desc.GetOptions().GetMapEntry()
}

// IsMessage returns true if the field is of message type.
func (fd *FieldDescription) IsMessage() bool {
	return fd.desc.GetType() == descpb.FieldDescriptorProto_TYPE_MESSAGE
}

// IsRepeated returns true if the field is a repeated value.
//
// This method will also return true for map values, so check whether the
// field is also a map.
func (fd *FieldDescription) IsRepeated() bool {
	return fd.prop.Repeated
}

// OrigName returns the snake_case name of the field as it was declared within
// the proto. This is the same name format that is expected within expressions.
func (fd *FieldDescription) OrigName() string {
	return fd.prop.OrigName
}

// Name returns the CamelCase name of the field within the proto-based struct.
func (fd *FieldDescription) Name() string {
	return fd.prop.Name
}

// SupportsPresence returns true if the field supports presence detection.
func (fd *FieldDescription) SupportsPresence() bool {
	return !fd.IsRepeated() && (fd.IsMessage() || !fd.proto3)
}

// String returns a proto-like field definition string.
func (fd *FieldDescription) String() string {
	return fmt.Sprintf("%s %s = %d `oneof=%t`",
		fd.TypeName(), fd.OrigName(), fd.Index(), fd.IsOneof())
}

// TypeName returns the type name of the field.
func (fd *FieldDescription) TypeName() string {
	return sanitizeProtoName(fd.desc.GetTypeName())
}

func (fd *FieldDescription) typeDefToType() *checked.Type {
	if fd.IsMessage() {
		if wk, found := CheckedWellKnowns[fd.TypeName()]; found {
			return wk
		}
		return checkedMessageType(fd.TypeName())
	}
	if fd.IsEnum() {
		return checkedInt
	}
	if p, found := CheckedPrimitives[fd.desc.GetType()]; found {
		return p
	}
	return CheckedPrimitives[fd.desc.GetType()]
}

func checkedMessageType(name string) *checked.Type {
	return &checked.Type{
		TypeKind: &checked.Type_MessageType{MessageType: name}}
}

func checkedPrimitive(primitive checked.Type_PrimitiveType) *checked.Type {
	return &checked.Type{
		TypeKind: &checked.Type_Primitive{Primitive: primitive}}
}

func checkedWellKnown(wellKnown checked.Type_WellKnownType) *checked.Type {
	return &checked.Type{
		TypeKind: &checked.Type_WellKnown{WellKnown: wellKnown}}
}

func checkedWrap(t *checked.Type) *checked.Type {
	return &checked.Type{
		TypeKind: &checked.Type_Wrapper{Wrapper: t.GetPrimitive()}}
}
