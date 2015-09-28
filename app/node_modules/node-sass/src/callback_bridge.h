#ifndef CALLBACK_BRIDGE_H
#define CALLBACK_BRIDGE_H

#include <vector>
#include <nan.h>
#include <condition_variable>
#include <algorithm>

#define COMMA ,

using namespace v8;

template <typename T, typename L = void*>
class CallbackBridge {
  public:
    CallbackBridge(NanCallback*, bool);
    virtual ~CallbackBridge();

    // Executes the callback
    T operator()(std::vector<void*>);

  protected:
    // We will expose a bridge object to the JS callback that wraps this instance so we don't loose context.
    // This is the V8 constructor for such objects.
    static Handle<Function> get_wrapper_constructor();
    static void async_gone(uv_handle_t *handle);
    static NAN_METHOD(New);
    static NAN_METHOD(ReturnCallback);
    static Persistent<Function> wrapper_constructor;
    Persistent<Object> wrapper;

    // The callback that will get called in the main thread after the worker thread used for the sass
    // compilation step makes a call to uv_async_send()
    static void dispatched_async_uv_callback(uv_async_t*);

    // The V8 values sent to our ReturnCallback must be read on the main thread not the sass worker thread.
    // This gives a chance to specialized subclasses to transform those values into whatever makes sense to
    // sass before we resume the worker thread.
    virtual T post_process_return_value(Handle<Value>) const =0;


    virtual std::vector<Handle<Value>> pre_process_args(std::vector<L>) const =0;

    NanCallback* callback;
    bool is_sync;

    std::mutex cv_mutex;
    std::condition_variable condition_variable;
    uv_async_t *async;
    std::vector<L> argv;
    bool has_returned;
    T return_value;
};

template <typename T, typename L>
Persistent<Function> CallbackBridge<T, L>::wrapper_constructor;

template <typename T, typename L>
CallbackBridge<T, L>::CallbackBridge(NanCallback* callback, bool is_sync) : callback(callback), is_sync(is_sync) {
  // This assumes the main thread will be the one instantiating the bridge
  if (!is_sync) {
    this->async = new uv_async_t;
    this->async->data = (void*) this;
    uv_async_init(uv_default_loop(), this->async, (uv_async_cb) dispatched_async_uv_callback);
  }

  NanAssignPersistent(wrapper, NanNew(CallbackBridge<T, L>::get_wrapper_constructor())->NewInstance());
  NanSetInternalFieldPointer(NanNew(wrapper), 0, this);
}

template <typename T, typename L>
CallbackBridge<T, L>::~CallbackBridge() {
  delete this->callback;
  NanDisposePersistent(this->wrapper);

  if (!is_sync) {
    uv_close((uv_handle_t*)this->async, &async_gone);
  }
}

template <typename T, typename L>
T CallbackBridge<T, L>::operator()(std::vector<void*> argv) {
  // argv.push_back(wrapper);

  if (this->is_sync) {
    std::vector<Handle<Value>> argv_v8 = pre_process_args(argv);
    argv_v8.push_back(NanNew(wrapper));

    return this->post_process_return_value(
      NanNew<Value>(this->callback->Call(argv_v8.size(), &argv_v8[0]))
    );
  }

  this->argv = argv;

  std::unique_lock<std::mutex> lock(this->cv_mutex);
  this->has_returned = false;
  uv_async_send(this->async);
  this->condition_variable.wait(lock, [this] { return this->has_returned; });

  return this->return_value;
}

template <typename T, typename L>
void CallbackBridge<T, L>::dispatched_async_uv_callback(uv_async_t *req) {
  CallbackBridge* bridge = static_cast<CallbackBridge*>(req->data);

  NanScope();
  TryCatch try_catch;

  std::vector<Handle<Value>> argv_v8 = bridge->pre_process_args(bridge->argv);
  argv_v8.push_back(NanNew(bridge->wrapper));

  NanNew<Value>(bridge->callback->Call(argv_v8.size(), &argv_v8[0]));

  if (try_catch.HasCaught()) {
    node::FatalException(try_catch);
  }
}

template <typename T, typename L>
NAN_METHOD(CallbackBridge<T COMMA L>::ReturnCallback) {
  NanScope();

  CallbackBridge<T, L>* bridge = static_cast<CallbackBridge<T, L>*>(NanGetInternalFieldPointer(args.This(), 0));
  TryCatch try_catch;

  bridge->return_value = bridge->post_process_return_value(args[0]);

  {
    std::lock_guard<std::mutex> lock(bridge->cv_mutex);
    bridge->has_returned = true;
  }

  bridge->condition_variable.notify_all();

  if (try_catch.HasCaught()) {
    node::FatalException(try_catch);
  }

  NanReturnUndefined();
}

template <typename T, typename L>
Handle<Function> CallbackBridge<T, L>::get_wrapper_constructor() {
  if (wrapper_constructor.IsEmpty()) {
    Local<FunctionTemplate> tpl = NanNew<FunctionTemplate>(New);
    tpl->SetClassName(NanNew("CallbackBridge"));
    tpl->InstanceTemplate()->SetInternalFieldCount(1);
    tpl->PrototypeTemplate()->Set(
      NanNew("success"),
      NanNew<FunctionTemplate>(ReturnCallback)->GetFunction()
    );

    NanAssignPersistent(wrapper_constructor, tpl->GetFunction());
  }

  return NanNew(wrapper_constructor);
}

template <typename T, typename L>
NAN_METHOD(CallbackBridge<T COMMA L>::New) {
  NanScope();
  NanReturnValue(args.This());
}

template <typename T, typename L>
void CallbackBridge<T, L>::async_gone(uv_handle_t *handle) {
  delete (uv_async_t *)handle;
}

#endif
