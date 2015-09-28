#include <nan.h>
#include "number.h"
#include "../create_string.h"

using namespace v8;

namespace SassTypes
{
  Number::Number(Sass_Value* v) : SassValueWrapper(v) {}

  Sass_Value* Number::construct(const std::vector<Local<v8::Value>> raw_val) {
    double value = 0;
    char const* unit = "";

    if (raw_val.size() >= 1) {
      if (!raw_val[0]->IsNumber()) {
        throw std::invalid_argument("First argument should be a number.");
      }

      value = raw_val[0]->ToNumber()->Value();

      if (raw_val.size() >= 2) {
        if (!raw_val[1]->IsString()) {
          throw std::invalid_argument("Second argument should be a string.");
        }

        unit = create_string(raw_val[1]);
      }
    }

    return sass_make_number(value, unit);
  }

  void Number::initPrototype(Handle<ObjectTemplate> proto) {
    proto->Set(NanNew("getValue"), NanNew<FunctionTemplate>(GetValue)->GetFunction());
    proto->Set(NanNew("getUnit"), NanNew<FunctionTemplate>(GetUnit)->GetFunction());
    proto->Set(NanNew("setValue"), NanNew<FunctionTemplate>(SetValue)->GetFunction());
    proto->Set(NanNew("setUnit"), NanNew<FunctionTemplate>(SetUnit)->GetFunction());
  }

  NAN_METHOD(Number::GetValue) {
    NanScope();
    NanReturnValue(NanNew(sass_number_get_value(unwrap(args.This())->value)));
  }

  NAN_METHOD(Number::GetUnit) {
    NanScope();
    NanReturnValue(NanNew(sass_number_get_unit(unwrap(args.This())->value)));
  }

  NAN_METHOD(Number::SetValue) {
    NanScope();

    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied value should be a number"));
    }

    sass_number_set_value(unwrap(args.This())->value, args[0]->ToNumber()->Value());
    NanReturnUndefined();
  }

  NAN_METHOD(Number::SetUnit) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsString()) {
      return NanThrowError(NanNew("Supplied value should be a string"));
    }

    sass_number_set_unit(unwrap(args.This())->value, create_string(args[0]));
    NanReturnUndefined();
  }
}
