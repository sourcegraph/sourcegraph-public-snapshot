package protoprint

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

func (p *Printer) printMessageLiteralCompact(msg protoreflect.Message, res *protoregistry.Types, pkg, scope string) string {
	var buf bytes.Buffer
	p.printMessageLiteralToBuffer(&buf, msg, res, pkg, scope, 0, -1)
	return buf.String()
}

func (p *Printer) printMessageLiteral(msg protoreflect.Message, res *protoregistry.Types, pkg, scope string, threshold, indent int) string {
	var buf bytes.Buffer
	p.printMessageLiteralToBuffer(&buf, msg, res, pkg, scope, threshold, indent)
	return buf.String()
}

var (
	anyTypeName = (&anypb.Any{}).ProtoReflect().Descriptor().FullName()
)

const (
	anyTypeUrlTag = 1
	anyValueTag   = 2
)

func (p *Printer) printMessageLiteralToBuffer(buf *bytes.Buffer, msg protoreflect.Message, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	if p.maybePrintAnyMessageToBuffer(buf, msg, res, pkg, scope, threshold, indent) {
		return
	}

	buf.WriteRune('{')
	if indent >= 0 {
		indent++
	}

	type fieldVal struct {
		fld protoreflect.FieldDescriptor
		val protoreflect.Value
	}
	var fields []fieldVal
	msg.Range(func(fld protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		fields = append(fields, fieldVal{fld: fld, val: val})
		return true
	})
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].fld.Number() < fields[j].fld.Number()
	})

	for i, fldVal := range fields {
		fld, val := fldVal.fld, fldVal.val
		if i > 0 {
			buf.WriteRune(',')
		}
		p.maybeNewline(buf, indent)
		if fld.IsExtension() {
			buf.WriteRune('[')
			buf.WriteString(p.qualifyExtensionLiteralName(pkg, scope, string(fld.FullName())))
			buf.WriteRune(']')
		} else {
			buf.WriteString(string(fld.Name()))
		}
		buf.WriteString(": ")
		switch {
		case fld.IsList():
			p.printArrayLiteralToBufferMaybeCompact(buf, fld, val.List(), res, pkg, scope, threshold, indent)
		case fld.IsMap():
			p.printMapLiteralToBufferMaybeCompact(buf, fld, val.Map(), res, pkg, scope, threshold, indent)
		case fld.Kind() == protoreflect.MessageKind || fld.Kind() == protoreflect.GroupKind:
			p.printMessageLiteralToBufferMaybeCompact(buf, val.Message(), res, pkg, scope, threshold, indent)
		default:
			p.printValueLiteralToBuffer(buf, fld, val.Interface())
		}
	}

	if indent >= 0 {
		indent--
	}
	p.maybeNewline(buf, indent)
	buf.WriteRune('}')
}

func (p *Printer) printMessageLiteralToBufferMaybeCompact(buf *bytes.Buffer, msg protoreflect.Message, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	if indent >= 0 {
		// first see if the message is compact enough to fit on one line
		str := p.printMessageLiteralCompact(msg, res, pkg, scope)
		fieldCount := strings.Count(str, ",")
		nestedCount := strings.Count(str, "{") - 1
		if fieldCount <= 1 && nestedCount == 0 {
			// can't expand
			buf.WriteString(str)
			return
		}
		if len(str) <= threshold {
			// no need to expand
			buf.WriteString(str)
			return
		}
	}
	p.printMessageLiteralToBuffer(buf, msg, res, pkg, scope, threshold, indent)
}

func (p *Printer) maybePrintAnyMessageToBuffer(buf *bytes.Buffer, msg protoreflect.Message, res *protoregistry.Types, pkg, scope string, threshold, indent int) bool {
	md := msg.Descriptor()
	if md.FullName() != anyTypeName {
		return false
	}
	typeUrlFld := md.Fields().ByNumber(anyTypeUrlTag)
	if typeUrlFld == nil || typeUrlFld.Kind() != protoreflect.StringKind || typeUrlFld.Cardinality() == protoreflect.Repeated {
		return false
	}
	valueFld := md.Fields().ByNumber(anyValueTag)
	if valueFld == nil || valueFld.Kind() != protoreflect.BytesKind || valueFld.Cardinality() == protoreflect.Repeated {
		return false
	}
	typeUrl := msg.Get(typeUrlFld).String()
	if typeUrl == "" {
		return false
	}
	mt, err := res.FindMessageByURL(typeUrl)
	if err != nil {
		return false
	}
	valueMsg := mt.New()
	valueBytes := msg.Get(valueFld).Bytes()
	if err := (proto.UnmarshalOptions{Resolver: res}).Unmarshal(valueBytes, valueMsg.Interface()); err != nil {
		return false
	}

	buf.WriteRune('{')
	if indent >= 0 {
		indent++
	}
	p.maybeNewline(buf, indent)

	buf.WriteRune('[')
	buf.WriteString(typeUrl)
	buf.WriteString("]: ")
	p.printMessageLiteralToBufferMaybeCompact(buf, valueMsg, res, pkg, scope, threshold, indent)

	if indent >= 0 {
		indent--
	}
	p.maybeNewline(buf, indent)
	buf.WriteRune('}')

	return true
}

func (p *Printer) printValueLiteralToBuffer(buf *bytes.Buffer, fld protoreflect.FieldDescriptor, value interface{}) {
	switch val := value.(type) {
	case protoreflect.EnumNumber:
		ev := fld.Enum().Values().ByNumber(val)
		if ev == nil {
			_, _ = fmt.Fprintf(buf, "%v", value)
		} else {
			buf.WriteString(string(ev.Name()))
		}
	case string:
		buf.WriteString(quotedString(val))
	case []byte:
		buf.WriteString(quotedBytes(string(val)))
	case int32, uint32, int64, uint64:
		_, _ = fmt.Fprintf(buf, "%d", val)
	case float32, float64:
		_, _ = fmt.Fprintf(buf, "%f", val)
	default:
		_, _ = fmt.Fprintf(buf, "%v", val)
	}
}

func (p *Printer) maybeNewline(buf *bytes.Buffer, indent int) {
	if indent < 0 {
		// compact form
		buf.WriteRune(' ')
		return
	}
	buf.WriteRune('\n')
	p.indent(buf, indent)
}

func (p *Printer) printArrayLiteralToBufferMaybeCompact(buf *bytes.Buffer, fld protoreflect.FieldDescriptor, val protoreflect.List, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	if indent >= 0 {
		// first see if the array is compact enough to fit on one line
		str := p.printArrayLiteralCompact(fld, val, res, pkg, scope)
		elementCount := strings.Count(str, ",")
		nestedCount := strings.Count(str, "{") - 1
		if elementCount <= 1 && nestedCount == 0 {
			// can't expand
			buf.WriteString(str)
			return
		}
		if len(str) <= threshold {
			// no need to expand
			buf.WriteString(str)
			return
		}
	}
	p.printArrayLiteralToBuffer(buf, fld, val, res, pkg, scope, threshold, indent)
}

func (p *Printer) printArrayLiteralCompact(fld protoreflect.FieldDescriptor, val protoreflect.List, res *protoregistry.Types, pkg, scope string) string {
	var buf bytes.Buffer
	p.printArrayLiteralToBuffer(&buf, fld, val, res, pkg, scope, 0, -1)
	return buf.String()
}

func (p *Printer) printArrayLiteralToBuffer(buf *bytes.Buffer, fld protoreflect.FieldDescriptor, val protoreflect.List, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	buf.WriteRune('[')
	if indent >= 0 {
		indent++
	}

	for i := 0; i < val.Len(); i++ {
		if i > 0 {
			buf.WriteRune(',')
		}
		p.maybeNewline(buf, indent)
		if fld.Kind() == protoreflect.MessageKind || fld.Kind() == protoreflect.GroupKind {
			p.printMessageLiteralToBufferMaybeCompact(buf, val.Get(i).Message(), res, pkg, scope, threshold, indent)
		} else {
			p.printValueLiteralToBuffer(buf, fld, val.Get(i).Interface())
		}
	}

	if indent >= 0 {
		indent--
	}
	p.maybeNewline(buf, indent)
	buf.WriteRune(']')
}

func (p *Printer) printMapLiteralToBufferMaybeCompact(buf *bytes.Buffer, fld protoreflect.FieldDescriptor, val protoreflect.Map, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	if indent >= 0 {
		// first see if the map is compact enough to fit on one line
		str := p.printMapLiteralCompact(fld, val, res, pkg, scope)
		if len(str) <= threshold {
			buf.WriteString(str)
			return
		}
	}
	p.printMapLiteralToBuffer(buf, fld, val, res, pkg, scope, threshold, indent)
}

func (p *Printer) printMapLiteralCompact(fld protoreflect.FieldDescriptor, val protoreflect.Map, res *protoregistry.Types, pkg, scope string) string {
	var buf bytes.Buffer
	p.printMapLiteralToBuffer(&buf, fld, val, res, pkg, scope, 0, -1)
	return buf.String()
}

func (p *Printer) printMapLiteralToBuffer(buf *bytes.Buffer, fld protoreflect.FieldDescriptor, val protoreflect.Map, res *protoregistry.Types, pkg, scope string, threshold, indent int) {
	keys := sortKeys(val)
	l := &mapAsList{
		m:      val,
		entry:  dynamicpb.NewMessageType(fld.Message()),
		keyFld: fld.MapKey(),
		valFld: fld.MapValue(),
		keys:   keys,
	}
	p.printArrayLiteralToBuffer(buf, fld, l, res, pkg, scope, threshold, indent)
}

type mapAsList struct {
	m              protoreflect.Map
	entry          protoreflect.MessageType
	keyFld, valFld protoreflect.FieldDescriptor
	keys           []protoreflect.MapKey
}

func (m *mapAsList) Len() int {
	return len(m.keys)
}

func (m *mapAsList) Get(i int) protoreflect.Value {
	msg := m.entry.New()
	key := m.keys[i]
	msg.Set(m.keyFld, key.Value())
	val := m.m.Get(key)
	msg.Set(m.valFld, val)
	return protoreflect.ValueOfMessage(msg)
}

func (m *mapAsList) Set(_i int, _ protoreflect.Value) {
	panic("Set is not implemented")
}

func (m *mapAsList) Append(_ protoreflect.Value) {
	panic("Append is not implemented")
}

func (m *mapAsList) AppendMutable() protoreflect.Value {
	panic("AppendMutable is not implemented")
}

func (m *mapAsList) Truncate(_ int) {
	panic("Truncate is not implemented")
}

func (m *mapAsList) NewElement() protoreflect.Value {
	panic("NewElement is not implemented")
}

func (m *mapAsList) IsValid() bool {
	return true
}
