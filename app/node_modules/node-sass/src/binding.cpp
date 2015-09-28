#include <nan.h>
#include <vector>
#include "sass_context_wrapper.h"
#include "custom_function_bridge.h"
#include "create_string.h"
#include "sass_types/factory.h"

Sass_Import_List sass_importer(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
{
  void* cookie = sass_importer_get_cookie(cb);
  struct Sass_Import* previous = sass_compiler_get_last_import(comp);
  const char* prev_path = sass_import_get_path(previous);
  sass_context_wrapper* ctx_w = static_cast<sass_context_wrapper*>(cookie);
  CustomImporterBridge& bridge = *(static_cast<CustomImporterBridge*>(cookie));

  std::vector<void*> argv;
  argv.push_back((void*)cur_path);
  argv.push_back((void*)prev_path);

  return bridge(argv);
}

union Sass_Value* sass_custom_function(const union Sass_Value* s_args, Sass_Function_Entry cb, struct Sass_Options* opts)
{
  void* cookie = sass_function_get_cookie(cb);
  CustomFunctionBridge& bridge = *(static_cast<CustomFunctionBridge*>(cookie));

  std::vector<void*> argv;
  for (unsigned l = sass_list_get_length(s_args), i = 0; i < l; i++) {
    argv.push_back((void*)sass_list_get_value(s_args, i));
  }

  try {
    return bridge(argv);
  }
  catch (const std::exception& e) {
    return sass_make_error(e.what());
  }
}

void ExtractOptions(Local<Object> options, void* cptr, sass_context_wrapper* ctx_w, bool is_file, bool is_sync) {
  NanScope();

  struct Sass_Context* ctx;

  NanAssignPersistent(ctx_w->result, options->Get(NanNew("result"))->ToObject());

  if (is_file) {
    ctx_w->fctx = (struct Sass_File_Context*) cptr;
    ctx = sass_file_context_get_context(ctx_w->fctx);
  }
  else {
    ctx_w->dctx = (struct Sass_Data_Context*) cptr;
    ctx = sass_data_context_get_context(ctx_w->dctx);
  }

  struct Sass_Options* sass_options = sass_context_get_options(ctx);

  ctx_w->is_sync = is_sync;

  if (!is_sync) {
    ctx_w->request.data = ctx_w;

    // async (callback) style
    Local<Function> success_callback = Local<Function>::Cast(options->Get(NanNew("success")));
    Local<Function> error_callback = Local<Function>::Cast(options->Get(NanNew("error")));

    ctx_w->success_callback = new NanCallback(success_callback);
    ctx_w->error_callback = new NanCallback(error_callback);
  }

  if (!is_file) {
    ctx_w->file = create_string(options->Get(NanNew("file")));
    sass_option_set_input_path(sass_options, ctx_w->file);
  }

  int indent_len = options->Get(NanNew("indentWidth"))->Int32Value();

  ctx_w->indent = (char*)malloc(indent_len + 1);

  strcpy(ctx_w->indent, std::string(
    indent_len,
    options->Get(NanNew("indentType"))->Int32Value() == 1 ? '\t' : ' '
    ).c_str());

  ctx_w->linefeed = create_string(options->Get(NanNew("linefeed")));
  ctx_w->include_path = create_string(options->Get(NanNew("includePaths")));
  ctx_w->out_file = create_string(options->Get(NanNew("outFile")));
  ctx_w->source_map = create_string(options->Get(NanNew("sourceMap")));
  ctx_w->source_map_root = create_string(options->Get(NanNew("sourceMapRoot")));

  sass_option_set_output_path(sass_options, ctx_w->out_file);
  sass_option_set_output_style(sass_options, (Sass_Output_Style)options->Get(NanNew("style"))->Int32Value());
  sass_option_set_is_indented_syntax_src(sass_options, options->Get(NanNew("indentedSyntax"))->BooleanValue());
  sass_option_set_source_comments(sass_options, options->Get(NanNew("sourceComments"))->BooleanValue());
  sass_option_set_omit_source_map_url(sass_options, options->Get(NanNew("omitSourceMapUrl"))->BooleanValue());
  sass_option_set_source_map_embed(sass_options, options->Get(NanNew("sourceMapEmbed"))->BooleanValue());
  sass_option_set_source_map_contents(sass_options, options->Get(NanNew("sourceMapContents"))->BooleanValue());
  sass_option_set_source_map_file(sass_options, ctx_w->source_map);
  sass_option_set_source_map_root(sass_options, ctx_w->source_map_root);
  sass_option_set_include_path(sass_options, ctx_w->include_path);
  sass_option_set_precision(sass_options, options->Get(NanNew("precision"))->Int32Value());
  sass_option_set_indent(sass_options, ctx_w->indent);
  sass_option_set_linefeed(sass_options, ctx_w->linefeed);

  Local<Value> importer_callback = options->Get(NanNew("importer"));

  if (importer_callback->IsFunction()) {
    Local<Function> importer = Local<Function>::Cast(importer_callback);
    auto bridge = std::make_shared<CustomImporterBridge>(new NanCallback(importer), ctx_w->is_sync);
    ctx_w->importer_bridges.push_back(bridge);

    Sass_Importer_List c_importers = sass_make_importer_list(1);
    c_importers[0] = sass_make_importer(sass_importer, 0, bridge.get());

    sass_option_set_c_importers(sass_options, c_importers);
  }
  else if (importer_callback->IsArray()) {
    Handle<Array> importers = Handle<Array>::Cast(importer_callback);
    Sass_Importer_List c_importers = sass_make_importer_list(importers->Length());

    for (size_t i = 0; i < importers->Length(); ++i) {
      Local<Function> callback = Local<Function>::Cast(importers->Get(static_cast<uint32_t>(i)));

      auto bridge = std::make_shared<CustomImporterBridge>(new NanCallback(callback), ctx_w->is_sync);
      ctx_w->importer_bridges.push_back(bridge);

      c_importers[i] = sass_make_importer(sass_importer, importers->Length() - i - 1, bridge.get());
    }

    sass_option_set_c_importers(sass_options, c_importers);
  }

  Local<Value> custom_functions = options->Get(NanNew("functions"));

  if (custom_functions->IsObject()) {
    Local<Object> functions = Local<Object>::Cast(custom_functions);
    Local<Array> signatures = functions->GetOwnPropertyNames();
    unsigned num_signatures = signatures->Length();
    Sass_Function_List fn_list = sass_make_function_list(num_signatures);

    for (unsigned i = 0; i < num_signatures; i++) {
      Local<String> signature = Local<String>::Cast(signatures->Get(NanNew(i)));
      Local<Function> callback = Local<Function>::Cast(functions->Get(signature));

      auto bridge = std::make_shared<CustomFunctionBridge>(new NanCallback(callback), ctx_w->is_sync);
      ctx_w->function_bridges.push_back(bridge);

      Sass_Function_Entry fn = sass_make_function(create_string(signature), sass_custom_function, bridge.get());
      sass_function_set_list_entry(fn_list, i, fn);
    }

    sass_option_set_c_functions(sass_options, fn_list);
  }
}

void GetStats(sass_context_wrapper* ctx_w, Sass_Context* ctx) {
  NanScope();

  char** included_files = sass_context_get_included_files(ctx);
  Handle<Array> arr = NanNew<Array>();

  if (included_files) {
    for (int i = 0; included_files[i] != nullptr; ++i) {
      arr->Set(i, NanNew<String>(included_files[i]));
    }
  }

  NanNew(ctx_w->result)->Get(NanNew("stats"))->ToObject()->Set(NanNew("includedFiles"), arr);
}

int GetResult(sass_context_wrapper* ctx_w, Sass_Context* ctx, bool is_sync = false) {
  NanScope();

  int status = sass_context_get_error_status(ctx);

  if (status == 0) {
    const char* css = sass_context_get_output_string(ctx);
    const char* map = sass_context_get_source_map_string(ctx);

    NanNew(ctx_w->result)->Set(NanNew("css"), NanNewBufferHandle(css, static_cast<uint32_t>(strlen(css))));

    GetStats(ctx_w, ctx);

    if (map) {
      NanNew(ctx_w->result)->Set(NanNew("map"), NanNewBufferHandle(map, static_cast<uint32_t>(strlen(map))));
    }
  }
  else if (is_sync) {
    NanNew(ctx_w->result)->Set(NanNew("error"), NanNew<String>(sass_context_get_error_json(ctx)));
  }

  return status;
}

void MakeCallback(uv_work_t* req) {
  NanScope();

  TryCatch try_catch;
  sass_context_wrapper* ctx_w = static_cast<sass_context_wrapper*>(req->data);
  struct Sass_Context* ctx;

  if (ctx_w->dctx) {
    ctx = sass_data_context_get_context(ctx_w->dctx);
  }
  else {
    ctx = sass_file_context_get_context(ctx_w->fctx);
  }

  int status = GetResult(ctx_w, ctx);

  if (status == 0 && ctx_w->success_callback) {
    // if no error, do callback(null, result)
    ctx_w->success_callback->Call(0, 0);
  }
  else if (ctx_w->error_callback) {
    // if error, do callback(error)
    const char* err = sass_context_get_error_json(ctx);
    Local<Value> argv[] = {
      NanNew<String>(err)
    };
    ctx_w->error_callback->Call(1, argv);
  }
  if (try_catch.HasCaught()) {
    node::FatalException(try_catch);
  }

  sass_free_context_wrapper(ctx_w);
}

NAN_METHOD(render) {
  NanScope();

  Local<Object> options = args[0]->ToObject();
  char* source_string = create_string(options->Get(NanNew("data")));
  struct Sass_Data_Context* dctx = sass_make_data_context(source_string);
  sass_context_wrapper* ctx_w = sass_make_context_wrapper();

  ExtractOptions(options, dctx, ctx_w, false, false);

  int status = uv_queue_work(uv_default_loop(), &ctx_w->request, compile_it, (uv_after_work_cb)MakeCallback);

  assert(status == 0);

  NanReturnUndefined();
}

NAN_METHOD(render_sync) {
  NanScope();

  Local<Object> options = args[0]->ToObject();
  char* source_string = create_string(options->Get(NanNew("data")));
  struct Sass_Data_Context* dctx = sass_make_data_context(source_string);
  struct Sass_Context* ctx = sass_data_context_get_context(dctx);
  sass_context_wrapper* ctx_w = sass_make_context_wrapper();

  ExtractOptions(options, dctx, ctx_w, false, true);

  compile_data(dctx);

  int result = GetResult(ctx_w, ctx, true);

  sass_free_context_wrapper(ctx_w);

  NanReturnValue(NanNew<Boolean>(result == 0));
}

NAN_METHOD(render_file) {
  NanScope();

  Local<Object> options = args[0]->ToObject();
  char* input_path = create_string(options->Get(NanNew("file")));
  struct Sass_File_Context* fctx = sass_make_file_context(input_path);
  sass_context_wrapper* ctx_w = sass_make_context_wrapper();

  ExtractOptions(options, fctx, ctx_w, true, false);

  int status = uv_queue_work(uv_default_loop(), &ctx_w->request, compile_it, (uv_after_work_cb)MakeCallback);

  assert(status == 0);

  NanReturnUndefined();
}

NAN_METHOD(render_file_sync) {
  NanScope();

  Local<Object> options = args[0]->ToObject();
  char* input_path = create_string(options->Get(NanNew("file")));
  struct Sass_File_Context* fctx = sass_make_file_context(input_path);
  struct Sass_Context* ctx = sass_file_context_get_context(fctx);
  sass_context_wrapper* ctx_w = sass_make_context_wrapper();

  ExtractOptions(options, fctx, ctx_w, true, true);
  compile_file(fctx);

  int result = GetResult(ctx_w, ctx, true);

  free(input_path);
  sass_free_context_wrapper(ctx_w);

  NanReturnValue(NanNew<Boolean>(result == 0));
}

NAN_METHOD(libsass_version) {
  NanScope();
  NanReturnValue(NanNew<String>(libsass_version()));
}

void RegisterModule(v8::Handle<v8::Object> target) {
  NODE_SET_METHOD(target, "render", render);
  NODE_SET_METHOD(target, "renderSync", render_sync);
  NODE_SET_METHOD(target, "renderFile", render_file);
  NODE_SET_METHOD(target, "renderFileSync", render_file_sync);
  NODE_SET_METHOD(target, "libsassVersion", libsass_version);
  SassTypes::Factory::initExports(target);
}

NODE_MODULE(binding, RegisterModule);
