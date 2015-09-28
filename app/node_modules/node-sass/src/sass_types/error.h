#ifndef SASS_TYPES_ERROR_H
#define SASS_TYPES_ERROR_H

#include <nan.h>
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class Error : public SassValueWrapper<Error> {
    public:
      Error(Sass_Value*);
      static char const* get_constructor_name() { return "SassError"; }
      static Sass_Value* construct(const std::vector<Local<v8::Value>>);

      static void initPrototype(Handle<ObjectTemplate>);
  };
}

#endif
