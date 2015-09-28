#include <nan.h>
#include "boolean.h"

using namespace v8;

namespace SassTypes
{
  Persistent<Function> Boolean::constructor;
  bool Boolean::constructor_locked = false;

  Boolean::Boolean(bool v) : value(v) {}

  Boolean& Boolean::get_singleton(bool v) {
    static Boolean instance_false(false), instance_true(true);
    return v ? instance_true : instance_false;
  }

  Handle<Function> Boolean::get_constructor() {
    if (constructor.IsEmpty()) {
      Local<FunctionTemplate> tpl = NanNew<FunctionTemplate>(New);

      tpl->SetClassName(NanNew("SassBoolean"));
      tpl->InstanceTemplate()->SetInternalFieldCount(1);
      tpl->PrototypeTemplate()->Set(NanNew("getValue"), NanNew<FunctionTemplate>(GetValue)->GetFunction());

      NanAssignPersistent(constructor, tpl->GetFunction());

      NanAssignPersistent(get_singleton(false).js_object, NanNew(constructor)->NewInstance());
      NanSetInternalFieldPointer(NanNew(get_singleton(false).js_object), 0, &get_singleton(false));
      NanNew(constructor)->Set(NanNew("FALSE"), NanNew(get_singleton(false).js_object));

      NanAssignPersistent(get_singleton(true).js_object, NanNew(constructor)->NewInstance());
      NanSetInternalFieldPointer(NanNew(get_singleton(true).js_object), 0, &get_singleton(true));
      NanNew(constructor)->Set(NanNew("TRUE"), NanNew(get_singleton(true).js_object));

      constructor_locked = true;
    }

    return NanNew(constructor);
  }

  Sass_Value* Boolean::get_sass_value() {
    return sass_make_boolean(value);
  }

  Local<Object> Boolean::get_js_object() {
    return NanNew(this->js_object);
  }

  NAN_METHOD(Boolean::New) {
    NanScope();

    if (args.IsConstructCall()) {
      if (constructor_locked) {
        return NanThrowError(NanNew("Cannot instantiate SassBoolean"));
      }
    }
    else {
      if (args.Length() != 1 || !args[0]->IsBoolean()) {
        return NanThrowError(NanNew("Expected one boolean argument"));
      }

      NanReturnValue(NanNew(get_singleton(args[0]->ToBoolean()->Value()).get_js_object()));
    }

    NanReturnUndefined();
  }

  NAN_METHOD(Boolean::GetValue) {
    NanScope();
    NanReturnValue(NanNew(static_cast<Boolean*>(Factory::unwrap(args.This()))->value));
  }
}
