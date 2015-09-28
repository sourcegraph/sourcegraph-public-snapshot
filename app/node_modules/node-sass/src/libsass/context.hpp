#ifndef SASS_CONTEXT_H
#define SASS_CONTEXT_H

#include <string>
#include <vector>
#include <map>

#define BUFFERSIZE 255
#include "b64/encode.h"

#include "ast_fwd_decl.hpp"
#include "kwd_arg_macros.hpp"
#include "memory_manager.hpp"
#include "environment.hpp"
#include "source_map.hpp"
#include "subset_map.hpp"
#include "output.hpp"
#include "plugins.hpp"
#include "sass_functions.h"

struct Sass_Function;

namespace Sass {
  using namespace std;
  struct Sass_Queued {
    string abs_path;
    string load_path;
    const char* source;
  public:
    Sass_Queued(const string& load_path, const string& abs_path, const char* source);
  };

  class Context {
  public:
    size_t head_imports;
    Memory_Manager<AST_Node> mem;

    struct Sass_Options* c_options;
    struct Sass_Compiler* c_compiler;
    const char* source_c_str;

    // c-strs containing Sass file contents
    // we will overtake ownership of memory
    vector<const char*> sources;
    // absolute paths to includes
    vector<string> included_files;
    // relative links to includes
    vector<string> include_links;
    // vectors above have same size

    vector<string> plugin_paths; // relative paths to load plugins
    vector<string> include_paths; // lookup paths for includes
    vector<Sass_Queued> queue; // queue of files to be parsed
    map<string, Block*> style_sheets; // map of paths to ASTs
    // SourceMap source_map;
    Output emitter;

    vector<Sass_Importer_Entry> c_headers;
    vector<Sass_Importer_Entry> c_importers;
    vector<Sass_Function_Entry> c_functions;

    void add_c_header(Sass_Importer_Entry header);
    void add_c_importer(Sass_Importer_Entry importer);
    void add_c_function(Sass_Function_Entry function);

    string       indent; // String to be used for indentation
    string       linefeed; // String to be used for line feeds
    string       input_path; // for relative paths in src-map
    string       output_path; // for relative paths to the output
    bool         source_comments; // for inline debug comments in css output
    Output_Style output_style; // output style for the generated css code
    string       source_map_file; // path to source map file (enables feature)
    string       source_map_root; // path for sourceRoot property (pass-through)
    bool         source_map_embed; // embed in sourceMappingUrl (as data-url)
    bool         source_map_contents; // insert included contents into source map
    bool         omit_source_map_url; // disable source map comment in css output
    bool         is_indented_syntax_src; // treat source string as sass

    // overload import calls
    vector<Sass_Import_Entry> import_stack;

    map<string, Color*> names_to_colors;
    map<int, string>    colors_to_names;

    size_t precision; // precision for outputting fractional numbers

    KWD_ARG_SET(Data) {
      KWD_ARG(Data, struct Sass_Options*, c_options);
      KWD_ARG(Data, struct Sass_Compiler*, c_compiler);
      KWD_ARG(Data, const char*,     source_c_str);
      KWD_ARG(Data, string,          entry_point);
      KWD_ARG(Data, string,          input_path);
      KWD_ARG(Data, string,          output_path);
      KWD_ARG(Data, string,          indent);
      KWD_ARG(Data, string,          linefeed);
      KWD_ARG(Data, const char*,     include_paths_c_str);
      KWD_ARG(Data, const char*,     plugin_paths_c_str);
      // KWD_ARG(Data, const char**,    include_paths_array);
      // KWD_ARG(Data, const char**,    plugin_paths_array);
      KWD_ARG(Data, vector<string>,  include_paths);
      KWD_ARG(Data, vector<string>,  plugin_paths);
      KWD_ARG(Data, bool,            source_comments);
      KWD_ARG(Data, Output_Style,    output_style);
      KWD_ARG(Data, string,          source_map_file);
      KWD_ARG(Data, string,          source_map_root);
      KWD_ARG(Data, bool,            omit_source_map_url);
      KWD_ARG(Data, bool,            is_indented_syntax_src);
      KWD_ARG(Data, size_t,          precision);
      KWD_ARG(Data, bool,            source_map_embed);
      KWD_ARG(Data, bool,            source_map_contents);
    };

    Context(Data);
    ~Context();
    static string get_cwd();
    void setup_color_map();

    Block* parse_file();
    Block* parse_string();
    void add_source(string, string, const char*);

    string add_file(const string& file);
    string add_file(const string& base, const string& file);


    // allow to optionally overwrite the input path
    // default argument for input_path is string("stdin")
    // usefull to influence the source-map generating etc.
    char* compile_file();
    char* compile_string();
    char* compile_block(Block* root);
    char* generate_source_map();

    vector<string> get_included_files(size_t skip = 0);

  private:
    void collect_plugin_paths(const char* paths_str);
    void collect_plugin_paths(const char** paths_array);
    void collect_include_paths(const char* paths_str);
    void collect_include_paths(const char** paths_array);
    string format_source_mapping_url(const string& file);

    string cwd;
    Plugins plugins;

    // void register_built_in_functions(Env* env);
    // void register_function(Signature sig, Native_Function f, Env* env);
    // void register_function(Signature sig, Native_Function f, size_t arity, Env* env);
    // void register_overload_stub(string name, Env* env);

  public:
    Subset_Map<string, pair<Complex_Selector*, Compound_Selector*> > subset_map;
  };

}

#endif
