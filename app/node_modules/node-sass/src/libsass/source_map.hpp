#ifndef SASS_SOURCE_MAP_H
#define SASS_SOURCE_MAP_H

#include <vector>

#include "ast_fwd_decl.hpp"
#include "base64vlq.hpp"
#include "position.hpp"
#include "mapping.hpp"

#define VECTOR_PUSH(vec, ins) vec.insert(vec.end(), ins.begin(), ins.end())
#define VECTOR_UNSHIFT(vec, ins) vec.insert(vec.begin(), ins.begin(), ins.end())

namespace Sass {
  using std::vector;

  class Context;
  class OutputBuffer;

  class SourceMap {

  public:
    vector<size_t> source_index;
    SourceMap();
    SourceMap(const string& file);

    void setFile(const string& str) {
      file = str;
    }
    void append(const Offset& offset);
    void prepend(const Offset& offset);
    void append(const OutputBuffer& out);
    void prepend(const OutputBuffer& out);
    void add_open_mapping(AST_Node* node);
    void add_close_mapping(AST_Node* node);

    string generate_source_map(Context &ctx);
    ParserState remap(const ParserState& pstate);

  private:

    string serialize_mappings();

    vector<Mapping> mappings;
    Position current_position;
public:
    string file;
private:
    Base64VLQ base64vlq;
  };

  class OutputBuffer {
    public:
      OutputBuffer(void)
      : buffer(""),
        smap()
      { }
    public:
      string buffer;
      SourceMap smap;
  };

}

#endif
