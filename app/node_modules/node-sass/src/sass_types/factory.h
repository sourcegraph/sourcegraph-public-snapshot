#ifndef SASS_TYPES_FACTORY_H
#define SASS_TYPES_FACTORY_H

#include <nan.h>
#include <sass_values.h>
#include "value.h"

namespace SassTypes
{
  using namespace v8;

  // This is the guru that knows everything about instantiating the right subclass of SassTypes::Value
  // to wrap a given Sass_Value object.
  class Factory {
    public:
      static void initExports(Handle<Object>);
      static Value* create(Sass_Value*);
      static Value* unwrap(Handle<v8::Value>);
  };
}

#endif
