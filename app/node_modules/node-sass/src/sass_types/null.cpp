#include <nan.h>
#include "null.h"

using namespace v8;

namespace SassTypes
{
  Persistent<Function> Null::constructor;
  bool Null::constructor_locked = false;

  Null::Null() {}

  Null& Null::get_singleton() {
    static Null singleton_instance;
    return singleton_instance;
  }

  Handle<Function> Null::get_constructor() {
    if (constructor.IsEmpty()) {
      Local<FunctionTemplate> tpl = NanNew<FunctionTemplate>(New);

      tpl->SetClassName(NanNew("SassNull"));
      tpl->InstanceTemplate()->SetInternalFieldCount(1);

      NanAssignPersistent(constructor, tpl->GetFunction());

      NanAssignPersistent(get_singleton().js_object, NanNew(constructor)->NewInstance());
      NanSetInternalFieldPointer(NanNew(get_singleton().js_object), 0, &get_singleton());
      NanNew(constructor)->Set(NanNew("NULL"), NanNew(get_singleton().js_object));

      constructor_locked = true;
    }

    return NanNew(constructor);
  }

  Sass_Value* Null::get_sass_value() {
    return sass_make_null();
  }

  Local<Object> Null::get_js_object() {
    return NanNew(this->js_object);
  }

  NAN_METHOD(Null::New) {
    NanScope();

    if (args.IsConstructCall()) {
      if (constructor_locked) {
        return NanThrowError(NanNew("Cannot instantiate SassNull"));
      }
    }
    else {
      NanReturnValue(NanNew(get_singleton().get_js_object()));
    }

    NanReturnUndefined();
  }
}
