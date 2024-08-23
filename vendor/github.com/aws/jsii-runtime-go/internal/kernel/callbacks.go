package kernel

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/jsii-runtime-go/internal/api"
)

type callback struct {
	CallbackID string          `json:"cbid"`
	Cookie     string          `json:"cookie"`
	Invoke     *invokeCallback `json:"invoke"`
	Get        *getCallback    `json:"get"`
	Set        *setCallback    `json:"set"`
}

func (c *callback) handle(result kernelResponder) error {
	var (
		retval reflect.Value
		err    error
	)
	if c.Invoke != nil {
		retval, err = c.Invoke.handle(c.Cookie)
	} else if c.Get != nil {
		retval, err = c.Get.handle(c.Cookie)
	} else if c.Set != nil {
		retval, err = c.Set.handle(c.Cookie)
	} else {
		return fmt.Errorf("invalid callback object: %v", c)
	}

	if err != nil {
		return err
	}

	type callbackResult struct {
		CallbackID string      `json:"cbid"`
		Result     interface{} `json:"result,omitempty"`
		Error      string      `json:"err,omitempty"`
	}
	type completeRequest struct {
		kernelRequester
		callbackResult `json:"complete"`
	}

	client := GetClient()
	request := completeRequest{}
	request.CallbackID = c.CallbackID
	request.Result = client.CastPtrToRef(retval)
	if err != nil {
		request.Error = err.Error()
	}
	return client.request(request, result)
}

type invokeCallback struct {
	Method    string        `json:"method"`
	Arguments []interface{} `json:"args"`
	ObjRef    api.ObjectRef `json:"objref"`
}

func (i *invokeCallback) handle(cookie string) (retval reflect.Value, err error) {
	client := GetClient()

	receiver := reflect.ValueOf(client.GetObject(i.ObjRef))
	method := receiver.MethodByName(cookie)

	return client.invoke(method, i.Arguments)
}

type getCallback struct {
	Property string        `json:"property"`
	ObjRef   api.ObjectRef `json:"objref"`
}

func (g *getCallback) handle(cookie string) (retval reflect.Value, err error) {
	client := GetClient()

	receiver := reflect.ValueOf(client.GetObject(g.ObjRef))

	if strings.HasPrefix(cookie, ".") {
		// Ready to catch an error if the access panics...
		defer func() {
			if r := recover(); r != nil {
				if err == nil {
					var ok bool
					if err, ok = r.(error); !ok {
						err = fmt.Errorf("%v", r)
					}
				} else {
					// This is not expected - so we panic!
					panic(r)
				}
			}
		}()

		// Need to access the underlying struct...
		receiver = receiver.Elem()
		retval = receiver.FieldByName(cookie[1:])

		if retval.IsZero() {
			// Omit zero-values if a json tag instructs so...
			field, _ := receiver.Type().FieldByName(cookie[1:])
			if tag := field.Tag.Get("json"); tag != "" {
				for _, attr := range strings.Split(tag, ",")[1:] {
					if attr == "omitempty" {
						retval = reflect.ValueOf(nil)
						break
					}
				}
			}
		}

		return
	} else {
		method := receiver.MethodByName(cookie)
		return client.invoke(method, nil)
	}
}

type setCallback struct {
	Property string        `json:"property"`
	Value    interface{}   `json:"value"`
	ObjRef   api.ObjectRef `json:"objref"`
}

func (s *setCallback) handle(cookie string) (retval reflect.Value, err error) {
	client := GetClient()

	receiver := reflect.ValueOf(client.GetObject(s.ObjRef))
	if strings.HasPrefix(cookie, ".") {
		// Ready to catch an error if the access panics...
		defer func() {
			if r := recover(); r != nil {
				if err == nil {
					var ok bool
					if err, ok = r.(error); !ok {
						err = fmt.Errorf("%v", r)
					}
				} else {
					// This is not expected - so we panic!
					panic(r)
				}
			}
		}()

		// Need to access the underlying struct...
		receiver = receiver.Elem()
		field := receiver.FieldByName(cookie[1:])
		meta, _ := receiver.Type().FieldByName(cookie[1:])

		field.Set(convert(reflect.ValueOf(s.Value), meta.Type))
		// Both retval & err are set to zero values here...
		return
	} else {
		method := receiver.MethodByName(fmt.Sprintf("Set%v", cookie))
		return client.invoke(method, []interface{}{s.Value})
	}
}

func convert(value reflect.Value, typ reflect.Type) reflect.Value {
retry:
	vt := value.Type()

	if vt.AssignableTo(typ) {
		return value
	}
	if value.CanConvert(typ) {
		return value.Convert(typ)
	}

	if typ.Kind() == reflect.Ptr {
		switch value.Kind() {
		case reflect.String:
			str := value.String()
			value = reflect.ValueOf(&str)
		case reflect.Bool:
			bool := value.Bool()
			value = reflect.ValueOf(&bool)
		case reflect.Int:
			int := value.Int()
			value = reflect.ValueOf(&int)
		case reflect.Float64:
			float := value.Float()
			value = reflect.ValueOf(&float)
		default:
			iface := value.Interface()
			value = reflect.ValueOf(&iface)
		}
		goto retry
	}

	// Unsure what to do... let default behavior happen...
	return value
}

func (c *Client) invoke(method reflect.Value, args []interface{}) (retval reflect.Value, err error) {
	if !method.IsValid() {
		err = fmt.Errorf("invalid method")
		return
	}

	// Convert the arguments, if any...
	callArgs := make([]reflect.Value, len(args))
	methodType := method.Type()
	numIn := methodType.NumIn()
	for i, arg := range args {
		var argType reflect.Type
		if i < numIn {
			argType = methodType.In(i)
		} else if methodType.IsVariadic() {
			argType = methodType.In(i - 1)
		} else {
			err = fmt.Errorf("too many arguments received %d for %d", len(args), numIn)
			return
		}
		if argType.Kind() == reflect.Ptr {
			callArgs[i] = reflect.New(argType.Elem())
		} else {
			callArgs[i] = reflect.New(argType)
		}
		c.castAndSetToPtr(callArgs[i].Elem(), reflect.ValueOf(arg))
		if argType.Kind() != reflect.Ptr {
			// The result of `reflect.New` is always a pointer, so if the
			// argument is by-value, we have to de-reference it first.
			callArgs[i] = callArgs[i].Elem()
		}
	}

	// Ready to catch an error if the method panics...
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				var ok bool
				if err, ok = r.(error); !ok {
					err = fmt.Errorf("%v", r)
				}
			} else {
				// This is not expected - so we panic!
				panic(r)
			}
		}
	}()

	result := method.Call(callArgs)
	switch len(result) {
	case 0:
		// Nothing to do, retval is already a 0-value.
	case 1:
		retval = result[0]
	default:
		err = fmt.Errorf("too many return values: %v", result)
	}
	return
}
