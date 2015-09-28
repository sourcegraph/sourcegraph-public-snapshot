#ifndef CUSTOM_IMPORTER_BRIDGE_H
#define CUSTOM_IMPORTER_BRIDGE_H

#include <nan.h>
#include <sass_context.h>
#include "callback_bridge.h"

using namespace v8;

typedef Sass_Import_List SassImportList;

class CustomImporterBridge : public CallbackBridge<SassImportList> {
  public:
    CustomImporterBridge(NanCallback* cb, bool is_sync) : CallbackBridge<SassImportList>(cb, is_sync) {}

  private:
    SassImportList post_process_return_value(Handle<Value>) const;
    Sass_Import* get_importer_entry(const Local<Object>&) const;
    std::vector<Handle<Value>> pre_process_args(std::vector<void*>) const;
};

#endif
