#ifndef SASS_TYPES_NULL_H
#define SASS_TYPES_NULL_H

#include <nan.h>
#include "value.h"

namespace SassTypes
{
  using namespace v8;

  class Null : public Value {
    public:
      static Null& get_singleton();
      static Handle<Function> get_constructor();

      Sass_Value* get_sass_value();
      Local<Object> get_js_object();

      static NAN_METHOD(New);

    private:
      Null();

      Persistent<Object> js_object;

      static Persistent<Function> constructor;
      static bool constructor_locked;
  };
}

#endif
