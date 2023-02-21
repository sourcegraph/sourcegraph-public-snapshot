package protorand

import (
	"fmt"
	"math/rand"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func Generate[T proto.Message](rand *rand.Rand, size int) reflect.Value {
	var zero T
	mds := zero.ProtoReflect().Descriptor()
	dm := newDynamicProtoRand(mds, rand, size)

	out := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(proto.Message)
	proto.Merge(out, dm)
	return reflect.ValueOf(out)
}

// newDynamicProtoRand creates dynamicpb with assiging random value to proto.
func newDynamicProtoRand(mds protoreflect.MessageDescriptor, rand *rand.Rand, size int) *dynamicpb.Message {
	println(mds.Name())
	getRandValue := func(fd protoreflect.FieldDescriptor) protoreflect.Value {
		switch fd.Kind() {
		case protoreflect.Int32Kind,
			protoreflect.Sint32Kind,
			protoreflect.Sfixed32Kind:
			return protoreflect.ValueOfInt32(rand.Int31())
		case
			protoreflect.Int64Kind,
			protoreflect.Sint64Kind,
			protoreflect.Sfixed64Kind:
			return protoreflect.ValueOfInt64(rand.Int63())
		case protoreflect.Uint32Kind,
			protoreflect.Fixed32Kind:
			return protoreflect.ValueOfUint32(rand.Uint32())
		case protoreflect.Uint64Kind,
			protoreflect.Fixed64Kind:
			return protoreflect.ValueOfUint64(rand.Uint64())
		case protoreflect.FloatKind:
			return protoreflect.ValueOfFloat32(rand.Float32())
		case protoreflect.DoubleKind:
			return protoreflect.ValueOfFloat64(rand.Float64())
		case protoreflect.BoolKind:
			return protoreflect.ValueOfBool(rand.Int()%2 == 0)
		case protoreflect.EnumKind:
			values := fd.Enum().Values()
			selected := values.Get(rand.Intn(values.Len() - 1))
			return protoreflect.ValueOfEnum(selected.Number())
		case protoreflect.BytesKind:
			buf := make([]byte, size)
			rand.Read(buf)
			return protoreflect.ValueOfBytes(buf)
		case protoreflect.StringKind:
			buf := make([]rune, size)
			for i := 0; i < size; i++ {
				buf[i] = rand.Int31n(0x10ffff)
			}
			return protoreflect.ValueOfString(string(buf))
		case protoreflect.MessageKind:
			// process recursively
			rm := newDynamicProtoRand(fd.Message(), rand, size)
			return protoreflect.ValueOfMessage(rm)
		default:
			panic(fmt.Sprintf("unexpected type: %v", fd.Kind()))
		}
	}

	// decide which fields in each OneOf will be populated in advance
	populatedOneOfField := map[protoreflect.Name]protoreflect.FieldNumber{}
	oneOfs := mds.Oneofs()
	for i := 0; i < oneOfs.Len(); i++ {
		oneOf := oneOfs.Get(i)
		populatedOneOfField[oneOf.Name()] = chooseOneOfFieldRandomly(oneOf, rand).Number()
	}

	dm := dynamicpb.NewMessage(mds)
	fds := mds.Fields()
	for k := 0; k < fds.Len(); k++ {
		fd := fds.Get(k)

		// If a field is in OneOf, check if the field should be populated
		if oneOf := fd.ContainingOneof(); oneOf != nil {
			populatedFieldNum := populatedOneOfField[oneOf.Name()]
			if populatedFieldNum != fd.Number() {
				continue
			}
		}

		if fd.IsList() {
			list := dm.Mutable(fd).List()
			// TODO: decide the number of elements randomly
			value := getRandValue(fd)
			list.Append(value)
			dm.Set(fd, protoreflect.ValueOfList(list))
			continue
		}
		if fd.IsMap() {
			mp := dm.Mutable(fd).Map()
			// TODO: make the number of elements randomly
			key := getRandValue(fd.MapKey())
			value := getRandValue(fd.MapValue())
			mp.Set(protoreflect.MapKey(key), protoreflect.Value(value))
			dm.Set(fd, protoreflect.ValueOfMap(mp))
			continue
		}

		value := getRandValue(fd)
		dm.Set(fd, value)
	}

	return dm
}

func chooseOneOfFieldRandomly(oneOf protoreflect.OneofDescriptor, rand *rand.Rand) protoreflect.FieldDescriptor {
	index := rand.Intn(oneOf.Fields().Len() - 1)
	return oneOf.Fields().Get(index)
}
