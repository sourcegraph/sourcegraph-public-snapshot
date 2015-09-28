#ifndef SASS_TYPES_BOOLEAN_H
#define SASS_TYPES_BOOLEAN_H

#include <nan.h>
#include "value.h"
#include "sass_value_wrapper.h"

namespace SassTypes
{
  using namespace v8;

  class Boolean : public Value {
    public:
      static Boolean& get_singleton(bool);
      static Handle<Function> get_constructor();

      Sass_Value* get_sass_value();
      Local<Object> get_js_object();

      static NAN_METHOD(New);
      static NAN_METHOD(GetValue);

    private:
      Boolean(bool);

      bool value;
      Persistent<Object> js_object;

      static Persistent<Function> constructor;
      static bool constructor_locked;
  };
}

#endif
