package kernel

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/aws/jsii-runtime-go/internal/api"
	"github.com/aws/jsii-runtime-go/internal/kernel/process"
	"github.com/aws/jsii-runtime-go/internal/objectstore"
	"github.com/aws/jsii-runtime-go/internal/typeregistry"
)

var (
	clientInstance      *Client
	clientInstanceMutex sync.Mutex
	clientOnce          sync.Once
	types               *typeregistry.TypeRegistry = typeregistry.New()
)

// The Client struct owns the jsii child process and its io interfaces. It also
// owns a map (objects) that tracks all object references by ID. This is used
// to call methods and access properties on objects passed by the runtime
// process by reference.
type Client struct {
	process *process.Process
	objects *objectstore.ObjectStore

	// Supports the idempotency of the Load method.
	loaded map[LoadProps]LoadResponse
}

// GetClient returns a singleton Client instance, initializing one the first
// time it is called.
func GetClient() *Client {
	clientOnce.Do(func() {
		// Locking early to be safe with a concurrent Close execution
		clientInstanceMutex.Lock()
		defer clientInstanceMutex.Unlock()

		client, err := newClient()
		if err != nil {
			panic(err)
		}

		clientInstance = client
	})

	return clientInstance
}

// CloseClient finalizes the runtime process, signalling the end of the
// execution to the jsii kernel process, and waiting for graceful termination.
//
// If a jsii Client is used *after* CloseClient was called, a new jsii kernel
// process will be initialized, and CloseClient should be called again to
// correctly finalize that, too.
func CloseClient() {
	// Locking early to be safe with a concurrent getClient execution
	clientInstanceMutex.Lock()
	defer clientInstanceMutex.Unlock()

	// Reset the "once" so a new Client would get initialized next time around
	clientOnce = sync.Once{}

	if clientInstance != nil {
		// Close the Client & reset it
		clientInstance.close()
		clientInstance = nil
	}
}

// newClient initializes a client, making it ready for business.
func newClient() (*Client, error) {
	if process, err := process.NewProcess(fmt.Sprintf("^%v", version)); err != nil {
		return nil, err
	} else {
		result := &Client{
			process: process,
			objects: objectstore.New(),
			loaded:  make(map[LoadProps]LoadResponse),
		}

		// Register a finalizer to call Close()
		runtime.SetFinalizer(result, func(c *Client) {
			c.close()
		})

		return result, nil
	}
}

func (c *Client) Types() *typeregistry.TypeRegistry {
	return types
}

func (c *Client) RegisterInstance(instance reflect.Value, objectRef api.ObjectRef) error {
	return c.objects.Register(instance, objectRef)
}

func (c *Client) request(req kernelRequester, res kernelResponder) error {
	return c.process.Request(req, res)
}

func (c *Client) FindObjectRef(obj reflect.Value) (ref api.ObjectRef, found bool) {
	ref = api.ObjectRef{}
	found = false

	switch obj.Kind() {
	case reflect.Struct:
		// Structs can be checked only if they are addressable, meaning
		// they are obtained from fields of an addressable struct.
		if !obj.CanAddr() {
			return
		}
		obj = obj.Addr()
		fallthrough
	case reflect.Interface, reflect.Ptr:
		if ref.InstanceID, found = c.objects.InstanceID(obj); found {
			ref.Interfaces = c.objects.Interfaces(ref.InstanceID)
		}
		return
	default:
		// Other types cannot possibly be object references!
		return
	}
}

func (c *Client) GetObject(objref api.ObjectRef) interface{} {
	if obj, ok := c.objects.GetObject(objref.InstanceID); ok {
		return obj.Interface()
	}
	panic(fmt.Errorf("no object found for ObjectRef %v", objref))
}

func (c *Client) close() {
	c.process.Close()

	// We no longer need a finalizer to run
	runtime.SetFinalizer(c, nil)
}
