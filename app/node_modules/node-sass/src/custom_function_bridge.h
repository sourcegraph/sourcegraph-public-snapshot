#ifndef CUSTOM_FUNCTION_BRIDGE_H
#define CUSTOM_FUNCTION_BRIDGE_H

#include <nan.h>
#include <sass_context.h>
#include "callback_bridge.h"

using namespace v8;

class CustomFunctionBridge : public CallbackBridge<Sass_Value*> {
  public:
    CustomFunctionBridge(NanCallback* cb, bool is_sync) : CallbackBridge<Sass_Value*>(cb, is_sync) {}

  private:
    Sass_Value* post_process_return_value(Handle<Value>) const;
    std::vector<Handle<Value>> pre_process_args(std::vector<void*>) const;
};

#endif
