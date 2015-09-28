#ifndef SASS_TYPES_NUMBER_H
#define SASS_TYPES_NUMBER_H

#include <nan.h>
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class Number : public SassValueWrapper<Number> {
    public:
      Number(Sass_Value*);
      static char const* get_constructor_name() { return "SassNumber"; }
      static Sass_Value* construct(const std::vector<Local<v8::Value>>);

      static void initPrototype(Handle<ObjectTemplate>);

      static NAN_METHOD(GetValue);
      static NAN_METHOD(GetUnit);
      static NAN_METHOD(SetValue);
      static NAN_METHOD(SetUnit);
  };
}

#endif
