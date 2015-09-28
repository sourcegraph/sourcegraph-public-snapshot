#ifndef SASS_TYPES_STRING_H
#define SASS_TYPES_STRING_H

#include <nan.h>
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class String : public SassValueWrapper<String> {
    public:
      String(Sass_Value*);
      static char const* get_constructor_name() { return "SassString"; }
      static Sass_Value* construct(const std::vector<Local<v8::Value>>);

      static void initPrototype(Handle<ObjectTemplate>);

      static NAN_METHOD(GetValue);
      static NAN_METHOD(SetValue);
  };
}

#endif
