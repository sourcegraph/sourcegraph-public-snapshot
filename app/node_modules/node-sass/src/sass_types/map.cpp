#include <nan.h>
#include "map.h"

using namespace v8;

namespace SassTypes
{
  Map::Map(Sass_Value* v) : SassValueWrapper(v) {}

  Sass_Value* Map::construct(const std::vector<Local<v8::Value>> raw_val) {
    size_t length = 0;

    if (raw_val.size() >= 1) {
      if (!raw_val[0]->IsNumber()) {
        throw std::invalid_argument("First argument should be an integer.");
      }

      length = raw_val[0]->ToInt32()->Value();
    }

    return sass_make_map(length);
  }

  void Map::initPrototype(Handle<ObjectTemplate> proto) {
    proto->Set(NanNew("getLength"), NanNew<FunctionTemplate>(GetLength)->GetFunction());
    proto->Set(NanNew("getKey"), NanNew<FunctionTemplate>(GetKey)->GetFunction());
    proto->Set(NanNew("setKey"), NanNew<FunctionTemplate>(SetKey)->GetFunction());
    proto->Set(NanNew("getValue"), NanNew<FunctionTemplate>(GetValue)->GetFunction());
    proto->Set(NanNew("setValue"), NanNew<FunctionTemplate>(SetValue)->GetFunction());
  }

  NAN_METHOD(Map::GetValue) {
    NanScope();

    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied index should be an integer"));
    }

    Sass_Value* map = unwrap(args.This())->value;
    size_t index = args[0]->ToInt32()->Value();


    if (index >= sass_map_get_length(map)) {
      return NanThrowError(NanNew("Out of bound index"));
    }

    NanReturnValue(NanNew(Factory::create(sass_map_get_value(map, args[0]->ToInt32()->Value()))->get_js_object()));
  }

  NAN_METHOD(Map::SetValue) {
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
    sass_map_set_value(unwrap(args.This())->value, args[0]->ToInt32()->Value(), sass_value->get_sass_value());
    NanReturnUndefined();
  }

  NAN_METHOD(Map::GetKey) {
    NanScope();

    if (args.Length() != 1) {
      return NanThrowError(NanNew("Expected just one argument"));
    }

    if (!args[0]->IsNumber()) {
      return NanThrowError(NanNew("Supplied index should be an integer"));
    }

    Sass_Value* map = unwrap(args.This())->value;
    size_t index = args[0]->ToInt32()->Value();


    if (index >= sass_map_get_length(map)) {
      return NanThrowError(NanNew("Out of bound index"));
    }

    NanReturnValue(Factory::create(sass_map_get_key(map, args[0]->ToInt32()->Value()))->get_js_object());
  }

  NAN_METHOD(Map::SetKey) {
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
    sass_map_set_key(unwrap(args.This())->value, args[0]->ToInt32()->Value(), sass_value->get_sass_value());
    NanReturnUndefined();
  }

  NAN_METHOD(Map::GetLength) {
    NanScope();
    NanReturnValue(NanNew<v8::Number>(sass_map_get_length(unwrap(args.This())->value)));
  }
}
