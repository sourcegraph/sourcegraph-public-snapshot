#include <nan.h>
#include "list.h"

using namespace v8;

namespace SassTypes
{
  List::List(Sass_Value* v) : SassValueWrapper(v) {}

  Sass_Value* List::construct(const std::vector<Local<v8::Value>> raw_val) {
    size_t length = 0;
    bool comma = true;

    if (raw_val.size() >= 1) {
      if (!raw_val[0]->IsNumber()) {
        throw std::invalid_argument("First argument should be an integer.");
      }

      length = raw_val[0]->ToInt32()->Value();

      if (raw_val.size() >= 2) {
        if (!raw_val[1]->IsBoolean()) {
          throw std::invalid_argument("Second argument should be a boolean.");
        }

        comma = raw_val[1]->ToBoolean()->Value();
      }
    }

    return sass_make_list(length, comma ? SASS_COMMA : SASS_SPACE);
  }

  void List::initPrototype(Handle<ObjectTemplate> proto) {
    proto->Set(NanNew("getLength"), NanNew<FunctionTemplate>(GetLength)->GetFunction());
    proto->Set(NanNew("getSeparator"), NanNew<FunctionTemplate>(GetSeparator)->GetFunction());
    proto->Set(NanNew("setSeparator"), NanNew<FunctionTemplate>(SetSeparator)->GetFunction());
    proto->Set(NanNew("getValue"), NanNew<FunctionTemplate>(GetValue)->GetFunction());
    proto->Set(NanNew("setValue"), NanNew<FunctionTemplate>(SetValue)->GetFunction());
  }

  NAN_METHOD(List::GetValue) {
    NanScope();

    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied index should be an integer"));
    }

    Sass_Value* list = unwrap(args.This())->value;
    size_t index = args[0]->ToInt32()->Value();


    if (index >= sass_list_get_length(list)) {
      return NanThrowError(NanNew("Out of bound index"));
    }

    NanReturnValue(Factory::create(sass_list_get_value(list, args[0]->ToInt32()->Value()))->get_js_object());
  }

  NAN_METHOD(List::SetValue) {
    if (args.Length() != 2) {
      return NanThrowError(NanNew("Expected two arguments"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied index should be an integer"));
    }

    if (!args[1]->IsObject()) {
      return NanThrowError(NanNew("Supplied value should be a SassValue object"));
    }

    Value* sass_value = Factory::unwrap(args[1]);
    sass_list_set_value(unwrap(args.This())->value, args[0]->ToInt32()->Value(), sass_value->get_sass_value());
    NanReturnUndefined();
  }

  NAN_METHOD(List::GetSeparator) {
    NanScope();
    NanReturnValue(NanNew(sass_list_get_separator(unwrap(args.This())->value) == SASS_COMMA));
  }

  NAN_METHOD(List::SetSeparator) {
    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsBoolean()) {
      return NanThrowError(NanNew("Supplied value should be a boolean"));
    }

    sass_list_set_separator(unwrap(args.This())->value, args[0]->ToBoolean()->Value() ? SASS_COMMA : SASS_SPACE);
    NanReturnUndefined();
  }

  NAN_METHOD(List::GetLength) {
    NanScope();
    NanReturnValue(NanNew<v8::Number>(sass_list_get_length(unwrap(args.This())->value)));
  }
}
