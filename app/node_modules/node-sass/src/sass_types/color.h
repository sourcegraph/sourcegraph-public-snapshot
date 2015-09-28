#ifndef SASS_TYPES_COLOR_H
#define SASS_TYPES_COLOR_H

#include <nan.h>
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class Color : public SassValueWrapper<Color> {
    public:
      Color(Sass_Value*);
      static char const* get_constructor_name() { return "SassColor"; }
      static Sass_Value* construct(const std::vector<Local<v8::Value>>);

      static void initPrototype(Handle<ObjectTemplate>);

      static NAN_METHOD(GetR);
      static NAN_METHOD(GetG);
      static NAN_METHOD(GetB);
      static NAN_METHOD(GetA);
      static NAN_METHOD(SetR);
      static NAN_METHOD(SetG);
      static NAN_METHOD(SetB);
      static NAN_METHOD(SetA);
  };
}

#endif
