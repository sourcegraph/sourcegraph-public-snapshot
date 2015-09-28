#ifdef _WIN32
#include <io.h>
#define LFEED "\n"
#else
#include <unistd.h>
#define LFEED "\n"
#endif

#include <cstring>
#include <stdexcept>
#include "file.hpp"
#include "json.hpp"
#include "util.hpp"
#include "context.hpp"
#include "sass_values.h"
#include "sass_context.h"
#include "ast_fwd_decl.hpp"
#include "error_handling.hpp"

extern "C" {
  using namespace std;
  using namespace Sass;

  // Input behaviours
  enum Sass_Input_Style {
    SASS_CONTEXT_NULL,
    SASS_CONTEXT_FILE,
    SASS_CONTEXT_DATA,
    SASS_CONTEXT_FOLDER
  };

  // simple linked list
  struct string_list {
    string_list* next;
    char* string;
  };

  // sass config options structure
  struct Sass_Options {

    // Precision for fractional numbers
    int precision;

    // Output style for the generated css code
    // A value from above SASS_STYLE_* constants
    enum Sass_Output_Style output_style;

    // Emit comments in the generated CSS indicating
    // the corresponding source line.
    bool source_comments;

    // embed sourceMappingUrl as data uri
    bool source_map_embed;

    // embed include contents in maps
    bool source_map_contents;

    // Disable sourceMappingUrl in css output
    bool omit_source_map_url;

    // Treat source_string as sass (as opposed to scss)
    bool is_indented_syntax_src;

    // The input path is used for source map
    // generation. It can be used to define
    // something with string compilation or to
    // overload the input file path. It is
    // set to "stdin" for data contexts and
    // to the input file on file contexts.
    char* input_path;

    // The output path is used for source map
    // generation. Libsass will not write to
    // this file, it is just used to create
    // information in source-maps etc.
    char* output_path;

    // String to be used for indentation
    const char* indent;
    // String to be used to for line feeds
    const char* linefeed;

    // Colon-separated list of paths
    // Semicolon-separated on Windows
    // Maybe use array interface instead?
    char* include_path;
    char* plugin_path;

    // Include paths (linked string list)
    struct string_list* include_paths;
    // Plugin paths (linked string list)
    struct string_list* plugin_paths;

    // Path to source map file
    // Enables source map generation
    // Used to create sourceMappingUrl
    char* source_map_file;

    // Directly inserted in source maps
    char* source_map_root;

    // Custom functions that can be called from sccs code
    Sass_Function_List c_functions;

    // List of custom importers
    Sass_Importer_List c_importers;

    // List of custom headers
    Sass_Importer_List c_headers;

  };

  // base for all contexts
  struct Sass_Context : Sass_Options
  {

    // store context type info
    enum Sass_Input_Style type;

    // generated output data
    char* output_string;

    // generated source map json
    char* source_map_string;

    // error status
    int error_status;
    char* error_json;
    char* error_text;
    char* error_message;
    // error position
    char* error_file;
    size_t error_line;
    size_t error_column;
    const char* error_src;

    // report imported files
    char** included_files;

  };

  // struct for file compilation
  struct Sass_File_Context : Sass_Context {

    // no additional fields required
    // input_path is already on options

  };

  // struct for data compilation
  struct Sass_Data_Context : Sass_Context {

    // provided source string
    char* source_string;

  };

  // link c and cpp context
  struct Sass_Compiler {
    // progress status
    Sass_Compiler_State state;
    // original c context
    Sass_Context* c_ctx;
    // Sass::Context
    Context* cpp_ctx;
    // Sass::Block
    Block* root;
  };

  static void copy_options(struct Sass_Options* to, struct Sass_Options* from) { *to = *from; }

  #define IMPLEMENT_SASS_OPTION_ACCESSOR(type, option) \
    type ADDCALL sass_option_get_##option (struct Sass_Options* options) { return options->option; } \
    void ADDCALL sass_option_set_##option (struct Sass_Options* options, type option) { options->option = option; }
  #define IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(type, option) \
    type ADDCALL sass_option_get_##option (struct Sass_Options* options) { return options->option; } \
    void ADDCALL sass_option_set_##option (struct Sass_Options* options, type option) \
    { free(options->option); options->option = option ? sass_strdup(option) : 0; }

  #define IMPLEMENT_SASS_CONTEXT_GETTER(type, option) \
    type ADDCALL sass_context_get_##option (struct Sass_Context* ctx) { return ctx->option; }
  #define IMPLEMENT_SASS_CONTEXT_TAKER(type, option) \
    type sass_context_take_##option (struct Sass_Context* ctx) \
    { type foo = ctx->option; ctx->option = 0; return foo; }

  // helper for safe access to c_ctx
  static const char* safe_str (const char* str) {
    return str == NULL ? "" : str;
  }

  static void copy_strings(const std::vector<std::string>& strings, char*** array) {
    int num = static_cast<int>(strings.size());
    char** arr = (char**) malloc(sizeof(char*) * (num + 1));
    if (arr == 0) throw(bad_alloc());

    for(int i = 0; i < num; i++) {
      arr[i] = (char*) malloc(sizeof(char) * (strings[i].size() + 1));
      if (arr[i] == 0) throw(bad_alloc());
      std::copy(strings[i].begin(), strings[i].end(), arr[i]);
      arr[i][strings[i].size()] = '\0';
    }

    arr[num] = 0;
    *array = arr;
  }

  static void free_string_array(char ** arr) {
    if(!arr)
        return;

    char **it = arr;
    while (it && (*it)) {
      free(*it);
      ++it;
    }

    free(arr);
  }

  static int handle_errors(Sass_Context* c_ctx) {
    try {
     throw;
    }
    catch (Sass_Error& e) {
      stringstream msg_stream;
      string cwd(Sass::File::get_cwd());
      JsonNode* json_err = json_mkobject();
      json_append_member(json_err, "status", json_mknumber(1));
      json_append_member(json_err, "file", json_mkstring(e.pstate.path.c_str()));
      json_append_member(json_err, "line", json_mknumber(e.pstate.line+1));
      json_append_member(json_err, "column", json_mknumber(e.pstate.column+1));
      json_append_member(json_err, "message", json_mkstring(e.message.c_str()));
      string rel_path(Sass::File::resolve_relative_path(e.pstate.path, cwd, cwd));

      string msg_prefix("Error: ");
      bool got_newline = false;
      msg_stream << msg_prefix;
      for (char chr : e.message) {
        if (chr == '\n') {
          got_newline = true;
        } else if (got_newline) {
          msg_stream << string(msg_prefix.size(), ' ');
          got_newline = false;
        }
        msg_stream << chr;
      }
      if (!got_newline) msg_stream << "\n";
      msg_stream << string(msg_prefix.size(), ' ');
      msg_stream << " on line " << e.pstate.line+1 << " of " << rel_path << "\n";

      // now create the code trace (ToDo: maybe have util functions?)
      if (e.pstate.line != string::npos && e.pstate.column != string::npos) {
        size_t line = e.pstate.line;
        const char* line_beg = e.pstate.src;
        while (line_beg && *line_beg && line) {
          if (*line_beg == '\n') -- line;
          ++ line_beg;
        }
        const char* line_end = line_beg;
        while (line_end && *line_end && *line_end != '\n') {
          if (*line_end == '\n') break;
          if (*line_end == '\r') break;
          line_end ++;
        }
        size_t max_left = 42; size_t max_right = 78;
        size_t move_in = e.pstate.column > max_left ? e.pstate.column - max_left : 0;
        size_t shorten = (line_end - line_beg) - move_in > max_right ?
                         (line_end - line_beg) - move_in - max_right : 0;
        msg_stream << ">> " << string(line_beg + move_in, line_end - shorten) << "\n";
        msg_stream << "   " << string(e.pstate.column - move_in, '-') << "^\n";
      }

      c_ctx->error_json = json_stringify(json_err, "  ");;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_text = strdup(e.message.c_str());
      c_ctx->error_status = 1;
      c_ctx->error_file = sass_strdup(e.pstate.path.c_str());
      c_ctx->error_line = e.pstate.line+1;
      c_ctx->error_column = e.pstate.column+1;
      c_ctx->error_src = e.pstate.src;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
      json_delete(json_err);
    }
    catch(bad_alloc& ba) {
      stringstream msg_stream;
      JsonNode* json_err = json_mkobject();
      msg_stream << "Unable to allocate memory: " << ba.what() << endl;
      json_append_member(json_err, "status", json_mknumber(2));
      json_append_member(json_err, "message", json_mkstring(ba.what()));
      c_ctx->error_json = json_stringify(json_err, "  ");;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_text = strdup(ba.what());
      c_ctx->error_status = 2;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
      json_delete(json_err);
    }
    catch (std::exception& e) {
      stringstream msg_stream;
      JsonNode* json_err = json_mkobject();
      msg_stream << "Error: " << e.what() << endl;
      json_append_member(json_err, "status", json_mknumber(3));
      json_append_member(json_err, "message", json_mkstring(e.what()));
      c_ctx->error_json = json_stringify(json_err, "  ");;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_text = strdup(e.what());
      c_ctx->error_status = 3;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
      json_delete(json_err);
    }
    catch (string& e) {
      stringstream msg_stream;
      JsonNode* json_err = json_mkobject();
      msg_stream << "Error: " << e << endl;
      json_append_member(json_err, "status", json_mknumber(4));
      json_append_member(json_err, "message", json_mkstring(e.c_str()));
      c_ctx->error_json = json_stringify(json_err, "  ");;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_text = strdup(e.c_str());
      c_ctx->error_status = 4;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
      json_delete(json_err);
    }
    catch (...) {
      stringstream msg_stream;
      JsonNode* json_err = json_mkobject();
      msg_stream << "Unknown error occurred" << endl;
      json_append_member(json_err, "status", json_mknumber(5));
      json_append_member(json_err, "message", json_mkstring("unknown"));
      c_ctx->error_json = json_stringify(json_err, "  ");;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_text = strdup("unknown");
      c_ctx->error_status = 5;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
      json_delete(json_err);
    }
    return c_ctx->error_status;
  }

  // generic compilation function (not exported, use file/data compile instead)
  static Sass_Compiler* sass_prepare_context (Sass_Context* c_ctx, Context::Data cpp_opt) throw()
  {
    try {

      // get input/output path from options
      string input_path = safe_str(c_ctx->input_path);
      string output_path = safe_str(c_ctx->output_path);
      // maybe we can extract an output path from input path
      if (output_path == "" && input_path != "") {
        int lastindex = static_cast<int>(input_path.find_last_of("."));
        output_path = (lastindex > -1 ? input_path.substr(0, lastindex) : input_path) + ".css";
      }

      // convert include path linked list to static array
      struct string_list* inc = c_ctx->include_paths;
      // very poor loop to get the length of the linked list
      size_t inc_size = 0; while (inc) { inc_size ++; inc = inc->next; }
      // create char* array to hold all paths plus null terminator
      const char** include_paths = (const char**) calloc(inc_size + 1, sizeof(char*));
      if (include_paths == 0) throw(bad_alloc());
      // reset iterator
      inc = c_ctx->include_paths;
      // copy over the paths
      for (size_t i = 0; inc; i++) {
        include_paths[i] = inc->string;
        inc = inc->next;
      }

      // convert plugin path linked list to static array
      struct string_list* imp = c_ctx->plugin_paths;
      // very poor loop to get the length of the linked list
      size_t imp_size = 0; while (imp) { imp_size ++; imp = imp->next; }
      // create char* array to hold all paths plus null terminator
      const char** plugin_paths = (const char**) calloc(imp_size + 1, sizeof(char*));
      if (plugin_paths == 0) throw(bad_alloc());
      // reset iterator
      imp = c_ctx->plugin_paths;
      // copy over the paths
      for (size_t i = 0; imp; i++) {
        plugin_paths[i] = imp->string;
        imp = imp->next;
      }

      // transfer the options to c++
      cpp_opt.c_compiler(0)
             .c_options(c_ctx)
             .input_path(input_path)
             .output_path(output_path)
             .output_style((Output_Style) c_ctx->output_style)
             .is_indented_syntax_src(c_ctx->is_indented_syntax_src)
             .source_comments(c_ctx->source_comments)
             .source_map_file(safe_str(c_ctx->source_map_file))
             .source_map_root(safe_str(c_ctx->source_map_root))
             .source_map_embed(c_ctx->source_map_embed)
             .source_map_contents(c_ctx->source_map_contents)
             .omit_source_map_url(c_ctx->omit_source_map_url)
             .include_paths_c_str(c_ctx->include_path)
             .plugin_paths_c_str(c_ctx->plugin_path)
             // .include_paths_array(include_paths)
             // .plugin_paths_array(plugin_paths)
             .include_paths(vector<string>())
             .plugin_paths(vector<string>())
             .precision(c_ctx->precision)
             .linefeed(c_ctx->linefeed)
             .indent(c_ctx->indent);

      // create new c++ Context
      Context* cpp_ctx = new Context(cpp_opt);
      // free intermediate data
      free(include_paths);
      free(plugin_paths);

      // register our custom functions
      if (c_ctx->c_functions) {
        auto this_func_data = c_ctx->c_functions;
        while (this_func_data && *this_func_data) {
          cpp_ctx->add_c_function(*this_func_data);
          ++this_func_data;
        }
      }

      // register our custom headers
      if (c_ctx->c_headers) {
        auto this_head_data = c_ctx->c_headers;
        while (this_head_data && *this_head_data) {
          cpp_ctx->add_c_header(*this_head_data);
          ++this_head_data;
        }
      }

      // register our custom importers
      if (c_ctx->c_importers) {
        auto this_imp_data = c_ctx->c_importers;
        while (this_imp_data && *this_imp_data) {
          cpp_ctx->add_c_importer(*this_imp_data);
          ++this_imp_data;
        }
      }

      // reset error status
      c_ctx->error_json = 0;
      c_ctx->error_text = 0;
      c_ctx->error_message = 0;
      c_ctx->error_status = 0;
      // reset error position
      c_ctx->error_src = 0;
      c_ctx->error_file = 0;
      c_ctx->error_line = string::npos;
      c_ctx->error_column = string::npos;

      // allocate a new compiler instance
      Sass_Compiler* compiler = (struct Sass_Compiler*) calloc(1, sizeof(struct Sass_Compiler));
      compiler->state = SASS_COMPILER_CREATED;

      // store in sass compiler
      compiler->c_ctx = c_ctx;
      compiler->cpp_ctx = cpp_ctx;
      cpp_ctx->c_compiler = compiler;

      // use to parse block
      return compiler;

    }
    // pass errors to generic error handler
    catch (...) { handle_errors(c_ctx); }

    // error
    return 0;

  }

  static Block* sass_parse_block (Sass_Compiler* compiler) throw()
  {

    // assert valid pointer
    if (compiler == 0) return 0;
    // The cpp context must be set by now
    Context* cpp_ctx = compiler->cpp_ctx;
    Sass_Context* c_ctx = compiler->c_ctx;
    // We will take care to wire up the rest
    compiler->cpp_ctx->c_compiler = compiler;
    compiler->state = SASS_COMPILER_PARSED;

    try {

      // get input/output path from options
      string input_path = safe_str(c_ctx->input_path);
      string output_path = safe_str(c_ctx->output_path);

      // parsed root block
      Block* root = 0;

      // maybe skip some entries of included files
      // we do not include stdin for data contexts
      size_t skip = 0;

      // dispatch to the correct render function
      if (c_ctx->type == SASS_CONTEXT_FILE) {
        root = cpp_ctx->parse_file();
      } else if (c_ctx->type == SASS_CONTEXT_DATA) {
        root = cpp_ctx->parse_string();
        skip = 1; // skip first entry of includes
      }

      // skip all prefixed files?
      skip += cpp_ctx->head_imports;

      // copy the included files on to the context (dont forget to free)
      if (root) copy_strings(cpp_ctx->get_included_files(skip), &c_ctx->included_files);

      // return parsed block
      return root;

    }
    // pass errors to generic error handler
    catch (...) { handle_errors(c_ctx); }

    // error
    return 0;

  }

  // generic compilation function (not exported, use file/data compile instead)
  static int sass_compile_context (Sass_Context* c_ctx, Context::Data cpp_opt)
  {

    // prepare sass compiler with context and options
    Sass_Compiler* compiler = sass_prepare_context(c_ctx, cpp_opt);

    try {
      // call each compiler step
      sass_compiler_parse(compiler);
      sass_compiler_execute(compiler);
    }
    // pass errors to generic error handler
    catch (...) { handle_errors(c_ctx); }

    sass_delete_compiler(compiler);

    return c_ctx->error_status;
  }

  inline void init_options (struct Sass_Options* options)
  {
    options->precision = 5;
    options->indent = "  ";
    options->linefeed = LFEED;
  }

  Sass_Options* ADDCALL sass_make_options (void)
  {
    struct Sass_Options* options = (struct Sass_Options*) calloc(1, sizeof(struct Sass_Options));
    if (options == 0) { cerr << "Error allocating memory for options" << endl; return 0; }
    init_options(options);
    return options;
  }

  Sass_File_Context* ADDCALL sass_make_file_context(const char* input_path)
  {
    struct Sass_File_Context* ctx = (struct Sass_File_Context*) calloc(1, sizeof(struct Sass_File_Context));
    if (ctx == 0) { cerr << "Error allocating memory for file context" << endl; return 0; }
    ctx->type = SASS_CONTEXT_FILE;
    init_options(ctx);
    try {
      if (input_path == 0) { throw(runtime_error("File context created without an input path")); }
      if (*input_path == 0) { throw(runtime_error("File context created with empty input path")); }
      sass_option_set_input_path(ctx, input_path);
    } catch (...) {
      handle_errors(ctx);
    }
    return ctx;
  }

  Sass_Data_Context* ADDCALL sass_make_data_context(char* source_string)
  {
    struct Sass_Data_Context* ctx = (struct Sass_Data_Context*) calloc(1, sizeof(struct Sass_Data_Context));
    if (ctx == 0) { cerr << "Error allocating memory for data context" << endl; return 0; }
    ctx->type = SASS_CONTEXT_DATA;
    init_options(ctx);
    try {
      if (source_string == 0) { throw(runtime_error("Data context created without a source string")); }
      if (*source_string == 0) { throw(runtime_error("Data context created with empty source string")); }
      ctx->source_string = source_string;
    } catch (...) {
      handle_errors(ctx);
    }
    return ctx;
  }

  struct Sass_Compiler* ADDCALL sass_make_file_compiler (struct Sass_File_Context* c_ctx)
  {
    if (c_ctx == 0) return 0;
    Context::Data cpp_opt = Context::Data();
    cpp_opt.entry_point(c_ctx->input_path);
    return sass_prepare_context(c_ctx, cpp_opt);
  }

  struct Sass_Compiler* ADDCALL sass_make_data_compiler (struct Sass_Data_Context* c_ctx)
  {
    if (c_ctx == 0) return 0;
    Context::Data cpp_opt = Context::Data();
    cpp_opt.source_c_str(c_ctx->source_string);
    return sass_prepare_context(c_ctx, cpp_opt);
  }

  int ADDCALL sass_compile_data_context(Sass_Data_Context* data_ctx)
  {
    if (data_ctx == 0) return 1;
    Sass_Context* c_ctx = data_ctx;
    if (c_ctx->error_status)
      return c_ctx->error_status;
    Context::Data cpp_opt = Context::Data();
    try {
      if (data_ctx->source_string == 0) { throw(runtime_error("Data context has no source string")); }
      if (*data_ctx->source_string == 0) { throw(runtime_error("Data context has empty source string")); }
      cpp_opt.source_c_str(data_ctx->source_string);
    }
    catch (...) { return handle_errors(c_ctx) | 1; }
    return sass_compile_context(c_ctx, cpp_opt);
  }

  int ADDCALL sass_compile_file_context(Sass_File_Context* file_ctx)
  {
    if (file_ctx == 0) return 1;
    Sass_Context* c_ctx = file_ctx;
    if (c_ctx->error_status)
      return c_ctx->error_status;
    Context::Data cpp_opt = Context::Data();
    try {
      if (file_ctx->input_path == 0) { throw(runtime_error("File context has no input path")); }
      if (*file_ctx->input_path == 0) { throw(runtime_error("File context has empty input path")); }
      cpp_opt.entry_point(file_ctx->input_path);
    }
    catch (...) { return handle_errors(c_ctx) | 1; }
    return sass_compile_context(c_ctx, cpp_opt);
  }

  int ADDCALL sass_compiler_parse(struct Sass_Compiler* compiler)
  {
    if (compiler == 0) return 1;
    if (compiler->state == SASS_COMPILER_PARSED) return 0;
    if (compiler->state != SASS_COMPILER_CREATED) return -1;
    if (compiler->c_ctx == NULL) return 1;
    if (compiler->cpp_ctx == NULL) return 1;
    if (compiler->c_ctx->error_status)
      return compiler->c_ctx->error_status;
    // parse the context we have set up (file or data)
    compiler->root = sass_parse_block(compiler);
    // success
    return 0;
  }

  int ADDCALL sass_compiler_execute(struct Sass_Compiler* compiler)
  {
    if (compiler == 0) return 1;
    if (compiler->state == SASS_COMPILER_EXECUTED) return 0;
    if (compiler->state != SASS_COMPILER_PARSED) return -1;
    if (compiler->c_ctx == NULL) return 1;
    if (compiler->cpp_ctx == NULL) return 1;
    if (compiler->root == NULL) return 1;
    if (compiler->c_ctx->error_status)
      return compiler->c_ctx->error_status;
    compiler->state = SASS_COMPILER_EXECUTED;
    Context* cpp_ctx = (Context*) compiler->cpp_ctx;
    Block* root = (Block*) compiler->root;
    // compile the parsed root block
    try { compiler->c_ctx->output_string = cpp_ctx->compile_block(root); }
    // pass catched errors to generic error handler
    catch (...) { return handle_errors(compiler->c_ctx) | 1; }
    // generate source map json and store on context
    compiler->c_ctx->source_map_string = cpp_ctx->generate_source_map();
    // success
    return 0;
  }

  // helper function, not exported, only accessible locally
  static void sass_clear_options (struct Sass_Options* options)
  {
    if (options == 0) return;
    // Deallocate custom functions
    if (options->c_functions) {
      Sass_Function_List this_func_data = options->c_functions;
      while (this_func_data && *this_func_data) {
        free(*this_func_data);
        ++this_func_data;
      }
    }
    // Deallocate custom headers
    if (options->c_headers) {
      Sass_Importer_List this_head_data = options->c_headers;
      while (this_head_data && *this_head_data) {
        free(*this_head_data);
        ++this_head_data;
      }
    }
    // Deallocate custom importers
    if (options->c_importers) {
      Sass_Importer_List this_imp_data = options->c_importers;
      while (this_imp_data && *this_imp_data) {
        free(*this_imp_data);
        ++this_imp_data;
      }
    }
    // Deallocate inc paths
    if (options->plugin_paths) {
      struct string_list* cur;
      struct string_list* next;
      cur = options->plugin_paths;
      while (cur) {
        next = cur->next;
        free(cur->string);
        free(cur);
        cur = next;
      }
    }
    // Deallocate inc paths
    if (options->include_paths) {
      struct string_list* cur;
      struct string_list* next;
      cur = options->include_paths;
      while (cur) {
        next = cur->next;
        free(cur->string);
        free(cur);
        cur = next;
      }
    }
    // Free custom functions
    free(options->c_functions);
    // Free custom importers
    free(options->c_importers);
    free(options->c_headers);
    // Reset our pointers
    options->c_functions = 0;
    options->c_importers = 0;
    options->c_headers = 0;
    options->plugin_paths = 0;
    options->include_paths = 0;
  }

  // helper function, not exported, only accessible locally
  // sass_free_context is also defined in old sass_interface
  static void sass_clear_context (struct Sass_Context* ctx)
  {
    if (ctx == 0) return;
    // release the allocated memory (mostly via sass_strdup)
    if (ctx->output_string)     free(ctx->output_string);
    if (ctx->source_map_string) free(ctx->source_map_string);
    if (ctx->error_message)     free(ctx->error_message);
    if (ctx->error_text)        free(ctx->error_text);
    if (ctx->error_json)        free(ctx->error_json);
    if (ctx->error_file)        free(ctx->error_file);
    if (ctx->input_path)        free(ctx->input_path);
    if (ctx->output_path)       free(ctx->output_path);
    if (ctx->include_path)      free(ctx->include_path);
    if (ctx->source_map_file)   free(ctx->source_map_file);
    if (ctx->source_map_root)   free(ctx->source_map_root);
    free_string_array(ctx->included_files);
    // play safe and reset properties
    ctx->output_string = 0;
    ctx->source_map_string = 0;
    ctx->error_message = 0;
    ctx->error_text = 0;
    ctx->error_json = 0;
    ctx->error_file = 0;
    ctx->input_path = 0;
    ctx->output_path = 0;
    ctx->include_path = 0;
    ctx->source_map_file = 0;
    ctx->source_map_root = 0;
    ctx->included_files = 0;
    // now clear the options
    sass_clear_options(ctx);
  }

  void ADDCALL sass_delete_compiler (struct Sass_Compiler* compiler)
  {
    if (compiler == 0) return;
    Context* cpp_ctx = (Context*) compiler->cpp_ctx;
    compiler->cpp_ctx = 0;
    delete cpp_ctx;
    free(compiler);
  }

  // Deallocate all associated memory with contexts
  void ADDCALL sass_delete_file_context (struct Sass_File_Context* ctx) { sass_clear_context(ctx); free(ctx); }
  void ADDCALL sass_delete_data_context (struct Sass_Data_Context* ctx) { sass_clear_context(ctx); free(ctx); }

  // Getters for sass context from specific implementations
  struct Sass_Context* ADDCALL sass_file_context_get_context(struct Sass_File_Context* ctx) { return ctx; }
  struct Sass_Context* ADDCALL sass_data_context_get_context(struct Sass_Data_Context* ctx) { return ctx; }

  // Getters for context options from Sass_Context
  struct Sass_Options* ADDCALL sass_context_get_options(struct Sass_Context* ctx) { return ctx; }
  struct Sass_Options* ADDCALL sass_file_context_get_options(struct Sass_File_Context* ctx) { return ctx; }
  struct Sass_Options* ADDCALL sass_data_context_get_options(struct Sass_Data_Context* ctx) { return ctx; }
  void ADDCALL sass_file_context_set_options (struct Sass_File_Context* ctx, struct Sass_Options* opt) { copy_options(ctx, opt); }
  void ADDCALL sass_data_context_set_options (struct Sass_Data_Context* ctx, struct Sass_Options* opt) { copy_options(ctx, opt); }

  // Getters for Sass_Compiler options (get conected sass context)
  enum Sass_Compiler_State ADDCALL sass_compiler_get_state(struct Sass_Compiler* compiler) { return compiler->state; }
  struct Sass_Context* ADDCALL sass_compiler_get_context(struct Sass_Compiler* compiler) { return compiler->c_ctx; }
  // Getters for Sass_Compiler options (query import stack)
  size_t ADDCALL sass_compiler_get_import_stack_size(struct Sass_Compiler* compiler) { return compiler->cpp_ctx->import_stack.size(); }
  Sass_Import_Entry ADDCALL sass_compiler_get_last_import(struct Sass_Compiler* compiler) { return compiler->cpp_ctx->import_stack.back(); }
  Sass_Import_Entry ADDCALL sass_compiler_get_import_entry(struct Sass_Compiler* compiler, size_t idx) { return compiler->cpp_ctx->import_stack[idx]; }

  // Calculate the size of the stored null terminated array
  size_t ADDCALL sass_context_get_included_files_size (struct Sass_Context* ctx)
  { size_t l = 0; auto i = ctx->included_files; while (i && *i) { ++i; ++l; } return l; }

  // Create getter and setters for options
  IMPLEMENT_SASS_OPTION_ACCESSOR(int, precision);
  IMPLEMENT_SASS_OPTION_ACCESSOR(enum Sass_Output_Style, output_style);
  IMPLEMENT_SASS_OPTION_ACCESSOR(bool, source_comments);
  IMPLEMENT_SASS_OPTION_ACCESSOR(bool, source_map_embed);
  IMPLEMENT_SASS_OPTION_ACCESSOR(bool, source_map_contents);
  IMPLEMENT_SASS_OPTION_ACCESSOR(bool, omit_source_map_url);
  IMPLEMENT_SASS_OPTION_ACCESSOR(bool, is_indented_syntax_src);
  IMPLEMENT_SASS_OPTION_ACCESSOR(Sass_Function_List, c_functions);
  IMPLEMENT_SASS_OPTION_ACCESSOR(Sass_Importer_List, c_importers);
  IMPLEMENT_SASS_OPTION_ACCESSOR(Sass_Importer_List, c_headers);
  IMPLEMENT_SASS_OPTION_ACCESSOR(const char*, indent);
  IMPLEMENT_SASS_OPTION_ACCESSOR(const char*, linefeed);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, input_path);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, output_path);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, plugin_path);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, include_path);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, source_map_file);
  IMPLEMENT_SASS_OPTION_STRING_ACCESSOR(const char*, source_map_root);

  // Create getter and setters for context
  IMPLEMENT_SASS_CONTEXT_GETTER(int, error_status);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, error_json);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, error_message);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, error_text);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, error_file);
  IMPLEMENT_SASS_CONTEXT_GETTER(size_t, error_line);
  IMPLEMENT_SASS_CONTEXT_GETTER(size_t, error_column);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, error_src);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, output_string);
  IMPLEMENT_SASS_CONTEXT_GETTER(const char*, source_map_string);
  IMPLEMENT_SASS_CONTEXT_GETTER(char**, included_files);

  // Take ownership of memory (value on context is set to 0)
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, error_json);
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, error_message);
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, error_text);
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, error_file);
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, output_string);
  IMPLEMENT_SASS_CONTEXT_TAKER(char*, source_map_string);
  IMPLEMENT_SASS_CONTEXT_TAKER(char**, included_files);

  // Push function for include paths (no manipulation support for now)
  void ADDCALL sass_option_push_include_path(struct Sass_Options* options, const char* path)
  {

    struct string_list* include_path = (struct string_list*) calloc(1, sizeof(struct string_list));
    if (include_path == 0) return;
    include_path->string = path ? sass_strdup(path) : 0;
    struct string_list* last = options->include_paths;
    if (!options->include_paths) {
      options->include_paths = include_path;
    } else {
      while (last->next)
        last = last->next;
      last->next = include_path;
    }

  }

  // Push function for plugin paths (no manipulation support for now)
  void ADDCALL sass_option_push_plugin_path(struct Sass_Options* options, const char* path)
  {

    struct string_list* plugin_path = (struct string_list*) calloc(1, sizeof(struct string_list));
    if (plugin_path == 0) return;
    plugin_path->string = path ? sass_strdup(path) : 0;
    struct string_list* last = options->plugin_paths;
    if (!options->plugin_paths) {
      options->plugin_paths = plugin_path;
    } else {
      while (last->next)
        last = last->next;
      last->next = plugin_path;
    }

  }

}
