#ifndef SASS_TYPES_MAP_H
#define SASS_TYPES_MAP_H

#include <nan.h>
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class Map : public SassValueWrapper<Map> {
    public:
      Map(Sass_Value*);
      static char const* get_constructor_name() { return "SassMap"; }
      static Sass_Value* construct(const std::vector<Local<v8::Value>>);

      static void initPrototype(Handle<ObjectTemplate>);

      static NAN_METHOD(GetValue);
      static NAN_METHOD(SetValue);
      static NAN_METHOD(GetKey);
      static NAN_METHOD(SetKey);
      static NAN_METHOD(GetLength);
  };
}

#endif
