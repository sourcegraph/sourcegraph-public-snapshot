#include <nan.h>
#include "custom_importer_bridge.h"
#include "create_string.h"

SassImportList CustomImporterBridge::post_process_return_value(Handle<Value> val) const {
  SassImportList imports = 0;
  NanScope();

  Local<Value> returned_value = NanNew(val);

  if (returned_value->IsArray()) {
    Handle<Array> array = Handle<Array>::Cast(returned_value);

    imports = sass_make_import_list(array->Length());

    for (size_t i = 0; i < array->Length(); ++i) {
      Local<Value> value = array->Get(static_cast<uint32_t>(i));

      if (!value->IsObject()) {
        auto entry = sass_make_import_entry(0, 0, 0);
        sass_import_set_error(entry, "returned array must only contain object literals", -1, -1);
        continue;
      }

      Local<Object> object = Local<Object>::Cast(value);

      if (value->IsNativeError()) {
        char* message = create_string(object->Get(NanNew<String>("message")));

        imports[i] = sass_make_import_entry(0, 0, 0);

        sass_import_set_error(imports[i], message, -1, -1);
      }
      else {
        imports[i] = get_importer_entry(object);
      }
    }
  }
  else if (returned_value->IsNativeError()) {
    imports = sass_make_import_list(1);
    Local<Object> object = Local<Object>::Cast(returned_value);
    char* message = create_string(object->Get(NanNew<String>("message")));

    imports[0] = sass_make_import_entry(0, 0, 0);

    sass_import_set_error(imports[0], message, -1, -1);
  }
  else if (returned_value->IsObject()) {
    imports = sass_make_import_list(1);
    imports[0] = get_importer_entry(Local<Object>::Cast(returned_value));
  }

  return imports;
}

Sass_Import* CustomImporterBridge::get_importer_entry(const Local<Object>& object) const {
  auto returned_file = object->Get(NanNew<String>("file"));

  if (!returned_file->IsUndefined() && !returned_file->IsString()) {
    auto entry = sass_make_import_entry(0, 0, 0);
    sass_import_set_error(entry, "returned value of `file` must be a string", -1, -1);
    return entry;
  }

  auto returned_contents = object->Get(NanNew<String>("contents"));

  if (!returned_contents->IsUndefined() && !returned_contents->IsString()) {
    auto entry = sass_make_import_entry(0, 0, 0);
    sass_import_set_error(entry, "returned value of `contents` must be a string", -1, -1);
    return entry;
  }

  auto returned_map = object->Get(NanNew<String>("map"));

  if (!returned_map->IsUndefined() && !returned_map->IsString()) {
    auto entry = sass_make_import_entry(0, 0, 0);
    sass_import_set_error(entry, "returned value of `map` must be a string", -1, -1);
    return entry;
  }

  char* path = create_string(returned_file);
  char* contents = create_string(returned_contents);
  char* srcmap = create_string(returned_map);

  return sass_make_import_entry(path, contents, srcmap);
}

std::vector<Handle<Value>> CustomImporterBridge::pre_process_args(std::vector<void*> in) const {
  std::vector<Handle<Value>> out;

  for (void* ptr : in) {
    out.push_back(NanNew<String>((char const*)ptr));
  }

  return out;
}
