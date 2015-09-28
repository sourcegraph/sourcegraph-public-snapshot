#include <nan.h>
#include "factory.h"
#include "value.h"
#include "number.h"
#include "string.h"
#include "color.h"
#include "boolean.h"
#include "list.h"
#include "map.h"
#include "null.h"
#include "error.h"

using namespace v8;

namespace SassTypes
{
  Value* Factory::create(Sass_Value* v) {
    switch (sass_value_get_tag(v)) {
    case SASS_NUMBER:
      return new Number(v);

    case SASS_STRING:
      return new String(v);

    case SASS_COLOR:
      return new Color(v);

    case SASS_BOOLEAN:
      return &Boolean::get_singleton(sass_boolean_get_value(v));

    case SASS_LIST:
      return new List(v);

    case SASS_MAP:
      return new Map(v);

    case SASS_NULL:
      return &Null::get_singleton();

    case SASS_ERROR:
      return new Error(v);

    default:
      throw std::invalid_argument("Unknown type encountered.");
    }
  }

  void Factory::initExports(Handle<Object> exports) {
    Local<Object> types = NanNew<Object>();
    exports->Set(NanNew("types"), types);

    types->Set(NanNew("Number"), Number::get_constructor());
    types->Set(NanNew("String"), String::get_constructor());
    types->Set(NanNew("Color"), Color::get_constructor());
    types->Set(NanNew("Boolean"), Boolean::get_constructor());
    types->Set(NanNew("List"), List::get_constructor());
    types->Set(NanNew("Map"), Map::get_constructor());
    types->Set(NanNew("Null"), Null::get_constructor());
    types->Set(NanNew("Error"), Error::get_constructor());
  }

  Value* Factory::unwrap(Handle<v8::Value> obj) {
    // Todo: non-SassValue objects could easily fall under that condition, need to be more specific.
    if (!obj->IsObject() || obj->ToObject()->InternalFieldCount() != 1) {
      throw std::invalid_argument("A SassValue object was expected.");
    }

    return static_cast<Value*>(NanGetInternalFieldPointer(obj->ToObject(), 0));
  }
}
