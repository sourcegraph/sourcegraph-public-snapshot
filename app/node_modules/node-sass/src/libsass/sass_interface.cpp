#ifdef _WIN32
#include <io.h>
#define LFEED "\n"
#else
#include <unistd.h>
#define LFEED "\n"
#endif

#include <string>
#include <cstdlib>
#include <cstring>
#include <sstream>
#include <iostream>

#include "util.hpp"
#include "context.hpp"
#include "inspect.hpp"
#include "error_handling.hpp"
#include "sass_interface.h"


extern "C" {
  using namespace std;

  sass_context* sass_new_context()
  { return (sass_context*) calloc(1, sizeof(sass_context)); }

  // helper for safe access to c_ctx
  static const char* safe_str (const char* str) {
    return str == NULL ? "" : str;
  }

  static void copy_strings(const std::vector<std::string>& strings, char*** array, int skip = 0) {
    int num = static_cast<int>(strings.size());
    char** arr = (char**) malloc(sizeof(char*) * (num + 1));

    for(int i = skip; i < num; i++) {
      arr[i-skip] = (char*) malloc(sizeof(char) * (strings[i].size() + 1));
      std::copy(strings[i].begin(), strings[i].end(), arr[i-skip]);
      arr[i-skip][strings[i].size()] = '\0';
    }

    arr[num-skip] = 0;
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

  void sass_free_context(sass_context* ctx)
  {
    if (ctx->output_string)     free(ctx->output_string);
    if (ctx->source_map_string) free(ctx->source_map_string);
    if (ctx->error_message)     free(ctx->error_message);
    if (ctx->c_functions)       free(ctx->c_functions);

    free_string_array(ctx->included_files);

    free(ctx);
  }

  sass_file_context* sass_new_file_context()
  { return (sass_file_context*) calloc(1, sizeof(sass_file_context)); }

  void sass_free_file_context(sass_file_context* ctx)
  {
    if (ctx->output_string)     free(ctx->output_string);
    if (ctx->source_map_string) free(ctx->source_map_string);
    if (ctx->error_message)     free(ctx->error_message);
    if (ctx->c_functions)       free(ctx->c_functions);

    free_string_array(ctx->included_files);

    free(ctx);
  }

  sass_folder_context* sass_new_folder_context()
  { return (sass_folder_context*) calloc(1, sizeof(sass_folder_context)); }

  void sass_free_folder_context(sass_folder_context* ctx)
  {
    free_string_array(ctx->included_files);
    free(ctx);
  }

  int sass_compile(sass_context* c_ctx)
  {
    using namespace Sass;
    try {
      string input_path = safe_str(c_ctx->input_path);
      int lastindex = static_cast<int>(input_path.find_last_of("."));
      string output_path;
      if (!c_ctx->output_path) {
        if (input_path != "") {
          output_path = (lastindex > -1 ? input_path.substr(0, lastindex) : input_path) + ".css";
        }
      }
      else {
          output_path = c_ctx->output_path;
      }
      Context cpp_ctx(
        Context::Data().source_c_str(c_ctx->source_string)
                       .output_path(output_path)
                       .output_style((Output_Style) c_ctx->options.output_style)
                       .is_indented_syntax_src(c_ctx->options.is_indented_syntax_src)
                       .source_comments(c_ctx->options.source_comments)
                       .source_map_file(safe_str(c_ctx->options.source_map_file))
                       .source_map_root(safe_str(c_ctx->options.source_map_root))
                       .source_map_embed(c_ctx->options.source_map_embed)
                       .source_map_contents(c_ctx->options.source_map_contents)
                       .omit_source_map_url(c_ctx->options.omit_source_map_url)
                       .include_paths_c_str(c_ctx->options.include_paths)
                       .plugin_paths_c_str(c_ctx->options.plugin_paths)
                       // .include_paths_array(0)
                       // .plugin_paths_array(0)
                       .include_paths(vector<string>())
                       .plugin_paths(vector<string>())
                       .precision(c_ctx->options.precision ? c_ctx->options.precision : 5)
                       .indent(c_ctx->options.indent ? c_ctx->options.indent : "  ")
                       .linefeed(c_ctx->options.linefeed ? c_ctx->options.linefeed : LFEED)
      );
      if (c_ctx->c_functions) {
        Sass_Function_List this_func_data = c_ctx->c_functions;
        while ((this_func_data) && (*this_func_data)) {
          cpp_ctx.c_functions.push_back(*this_func_data);
          ++this_func_data;
        }
      }
      c_ctx->output_string = cpp_ctx.compile_string();
      c_ctx->source_map_string = cpp_ctx.generate_source_map();
      c_ctx->error_message = 0;
      c_ctx->error_status = 0;

      copy_strings(cpp_ctx.get_included_files(1), &c_ctx->included_files, 1);
    }
    catch (Sass_Error& e) {
      stringstream msg_stream;
      msg_stream << e.pstate.path << ":" << e.pstate.line << ": " << e.message << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch(bad_alloc& ba) {
      stringstream msg_stream;
      msg_stream << "Unable to allocate memory: " << ba.what() << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (std::exception& e) {
      stringstream msg_stream;
      msg_stream << "Error: " << e.what() << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (string& e) {
      stringstream msg_stream;
      msg_stream << "Error: " << e << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (...) {
      // couldn't find the specified file in the include paths; report an error
      stringstream msg_stream;
      msg_stream << "Unknown error occurred" << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    return 0;
  }

  int sass_compile_file(sass_file_context* c_ctx)
  {
    using namespace Sass;
    try {
      string input_path = safe_str(c_ctx->input_path);
      int lastindex = static_cast<int>(input_path.find_last_of("."));
      string output_path;
      if (!c_ctx->output_path) {
          output_path = (lastindex > -1 ? input_path.substr(0, lastindex) : input_path) + ".css";
      }
      else {
          output_path = c_ctx->output_path;
      }
      Context cpp_ctx(
        Context::Data().entry_point(input_path)
                       .output_path(output_path)
                       .output_style((Output_Style) c_ctx->options.output_style)
                       .is_indented_syntax_src(c_ctx->options.is_indented_syntax_src)
                       .source_comments(c_ctx->options.source_comments)
                       .source_map_file(safe_str(c_ctx->options.source_map_file))
                       .source_map_root(safe_str(c_ctx->options.source_map_root))
                       .source_map_embed(c_ctx->options.source_map_embed)
                       .source_map_contents(c_ctx->options.source_map_contents)
                       .omit_source_map_url(c_ctx->options.omit_source_map_url)
                       .include_paths_c_str(c_ctx->options.include_paths)
                       .plugin_paths_c_str(c_ctx->options.plugin_paths)
                       // .include_paths_array(0)
                       // .plugin_paths_array(0)
                       .include_paths(vector<string>())
                       .plugin_paths(vector<string>())
                       .precision(c_ctx->options.precision ? c_ctx->options.precision : 5)
                       .indent(c_ctx->options.indent ? c_ctx->options.indent : "  ")
                       .linefeed(c_ctx->options.linefeed ? c_ctx->options.linefeed : LFEED)
      );
      if (c_ctx->c_functions) {
        Sass_Function_List this_func_data = c_ctx->c_functions;
        while ((this_func_data) && (*this_func_data)) {
          cpp_ctx.c_functions.push_back(*this_func_data);
          ++this_func_data;
        }
      }
      c_ctx->output_string = cpp_ctx.compile_file();
      c_ctx->source_map_string = cpp_ctx.generate_source_map();
      c_ctx->error_message = 0;
      c_ctx->error_status = 0;

      copy_strings(cpp_ctx.get_included_files(), &c_ctx->included_files);
    }
    catch (Sass_Error& e) {
      stringstream msg_stream;
      msg_stream << e.pstate.path << ":" << e.pstate.line << ": " << e.message << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch(bad_alloc& ba) {
      stringstream msg_stream;
      msg_stream << "Unable to allocate memory: " << ba.what() << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (std::exception& e) {
      stringstream msg_stream;
      msg_stream << "Error: " << e.what() << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (string& e) {
      stringstream msg_stream;
      msg_stream << "Error: " << e << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    catch (...) {
      // couldn't find the specified file in the include paths; report an error
      stringstream msg_stream;
      msg_stream << "Unknown error occurred" << endl;
      c_ctx->error_message = sass_strdup(msg_stream.str().c_str());
      c_ctx->error_status = 1;
      c_ctx->output_string = 0;
      c_ctx->source_map_string = 0;
    }
    return 0;
  }

  int sass_compile_folder(sass_folder_context* c_ctx)
  {
    return 1;
  }

}
