#include "sass_context_wrapper.h"

extern "C" {
  using namespace std;

  void compile_it(uv_work_t* req) {
    sass_context_wrapper* ctx_w = (sass_context_wrapper*)req->data;

    if (ctx_w->dctx) {
      compile_data(ctx_w->dctx);
    }
    else if (ctx_w->fctx) {
      compile_file(ctx_w->fctx);
    }
  }

  void compile_data(struct Sass_Data_Context* dctx) {
    sass_compile_data_context(dctx);
  }

  void compile_file(struct Sass_File_Context* fctx) {
    sass_compile_file_context(fctx);
  }

  sass_context_wrapper* sass_make_context_wrapper() {
    return (sass_context_wrapper*)calloc(1, sizeof(sass_context_wrapper));
  }

  void sass_free_context_wrapper(sass_context_wrapper* ctx_w) {
    if (ctx_w->dctx) {
      sass_delete_data_context(ctx_w->dctx);
    }
    else if (ctx_w->fctx) {
      sass_delete_file_context(ctx_w->fctx);
    }

    delete ctx_w->error_callback;
    delete ctx_w->success_callback;

    NanDisposePersistent(ctx_w->result);

    free(ctx_w->include_path);
    free(ctx_w->linefeed);
    free(ctx_w->out_file);
    free(ctx_w->source_map);
    free(ctx_w->source_map_root);
    free(ctx_w->indent);

    ctx_w->importer_bridges.resize(0);
    ctx_w->function_bridges.resize(0);

    free(ctx_w);
  }
}
