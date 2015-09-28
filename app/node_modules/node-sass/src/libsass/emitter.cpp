#include "util.hpp"
#include "context.hpp"
#include "output.hpp"
#include "emitter.hpp"
#include "utf8_string.hpp"

namespace Sass {
  using namespace std;

  Emitter::Emitter(Context* ctx)
  : wbuf(),
    ctx(ctx),
    indentation(0),
    scheduled_space(0),
    scheduled_linefeed(0),
    scheduled_delimiter(false),
    in_comment(false),
    in_wrapped(false),
    in_media_block(false),
    in_declaration(false),
    in_space_array(false),
    in_comma_array(false)
  { }

  // return buffer as string
  string Emitter::get_buffer(void)
  {
    return wbuf.buffer;
  }

  Output_Style Emitter::output_style(void)
  {
    return ctx ? ctx->output_style : COMPRESSED;
  }

  // PROXY METHODS FOR SOURCE MAPS

  void Emitter::add_source_index(size_t idx)
  { wbuf.smap.source_index.push_back(idx); }

  string Emitter::generate_source_map(Context &ctx)
  { return wbuf.smap.generate_source_map(ctx); }

  void Emitter::set_filename(const string& str)
  { wbuf.smap.file = str; }

  void Emitter::add_open_mapping(AST_Node* node)
  { wbuf.smap.add_open_mapping(node); }
  void Emitter::add_close_mapping(AST_Node* node)
  { wbuf.smap.add_close_mapping(node); }
  ParserState Emitter::remap(const ParserState& pstate)
  { return wbuf.smap.remap(pstate); }

  // MAIN BUFFER MANIPULATION

  // add outstanding delimiter
  void Emitter::finalize(void)
  {
    scheduled_space = 0;
    if (scheduled_linefeed)
      scheduled_linefeed = 1;
    flush_schedules();
  }

  // flush scheduled space/linefeed
  void Emitter::flush_schedules(void)
  {
    // check the schedule
    if (scheduled_linefeed) {
      string linefeeds = "";

      for (size_t i = 0; i < scheduled_linefeed; i++)
        linefeeds += ctx ? ctx->linefeed : "\n";
      scheduled_space = 0;
      scheduled_linefeed = 0;
      append_string(linefeeds);

    } else if (scheduled_space) {
      string spaces(scheduled_space, ' ');
      scheduled_space = 0;
      append_string(spaces);
    }
    if (scheduled_delimiter) {
      scheduled_delimiter = false;
      append_string(";");
    }
  }

  // prepend some text or token to the buffer
  void Emitter::prepend_output(const OutputBuffer& output)
  {
    wbuf.smap.prepend(output);
    wbuf.buffer = output.buffer + wbuf.buffer;
  }

  // prepend some text or token to the buffer
  void Emitter::prepend_string(const string& text)
  {
    wbuf.smap.prepend(Offset(text));
    wbuf.buffer = text + wbuf.buffer;
  }

  // append some text or token to the buffer
  void Emitter::append_string(const string& text)
  {
    // write space/lf
    flush_schedules();

    if (in_comment && output_style() == COMPACT) {
      // unescape comment nodes
      string out = comment_to_string(text);
      // add to buffer
      wbuf.buffer += out;
      // account for data in source-maps
      wbuf.smap.append(Offset(out));
    } else {
      // add to buffer
      wbuf.buffer += text;
      // account for data in source-maps
      wbuf.smap.append(Offset(text));
    }
  }

  // append some white-space only text
  void Emitter::append_wspace(const string& text)
  {
    if (text.empty()) return;
    if (peek_linefeed(text.c_str())) {
      scheduled_space = 0;
      append_mandatory_linefeed();
    }
  }

  // append some text or token to the buffer
  // this adds source-mappings for node start and end
  void Emitter::append_token(const string& text, AST_Node* node)
  {
    flush_schedules();
    add_open_mapping(node);
    append_string(text);
    add_close_mapping(node);
  }

  // HELPER METHODS

  void Emitter::append_indentation()
  {
    if (output_style() == COMPRESSED) return;
    if (output_style() == COMPACT) return;
    if (scheduled_linefeed && indentation)
      scheduled_linefeed = 1;
    string indent = "";
    for (size_t i = 0; i < indentation; i++)
      indent += ctx ? ctx->indent : "  ";
    append_string(indent);
  }

  void Emitter::append_delimiter()
  {
    scheduled_delimiter = true;
    if (output_style() == COMPACT) {
      if (indentation == 0) {
        append_mandatory_linefeed();
      } else {
        append_mandatory_space();
      }
    } else if (output_style() != COMPRESSED) {
      append_optional_linefeed();
    }
  }

  void Emitter::append_comma_separator()
  {
    scheduled_space = 0;
    append_string(",");
    append_optional_space();
  }

  void Emitter::append_colon_separator()
  {
    scheduled_space = 0;
    append_string(":");
    append_optional_space();
  }

  void Emitter::append_mandatory_space()
  {
    scheduled_space = 1;
  }

  void Emitter::append_optional_space()
  {
    if (output_style() != COMPRESSED && buffer().size()) {
      char lst = buffer().at(buffer().length() - 1);
      if (!isspace(lst) || scheduled_delimiter) {
        append_mandatory_space();
      }
    }
  }

  void Emitter::append_special_linefeed()
  {
    if (output_style() == COMPACT) {
      append_mandatory_linefeed();
      for (size_t p = 0; p < indentation; p++)
        append_string(ctx ? ctx->indent : "  ");
    }
  }

  void Emitter::append_optional_linefeed()
  {
    if (output_style() == COMPACT) {
      append_mandatory_space();
    } else {
      append_mandatory_linefeed();
    }
  }

  void Emitter::append_mandatory_linefeed()
  {
    if (output_style() != COMPRESSED) {
      scheduled_linefeed = 1;
      scheduled_space = 0;
      // flush_schedules();
    }
  }

  void Emitter::append_scope_opener(AST_Node* node)
  {
    append_optional_space();
    flush_schedules();
    if (node) add_open_mapping(node);
    append_string("{");
    append_optional_linefeed();
    // append_optional_space();
    ++ indentation;
  }
  void Emitter::append_scope_closer(AST_Node* node)
  {
    -- indentation;
    scheduled_linefeed = 0;
    if (output_style() == COMPRESSED)
      scheduled_delimiter = false;
    if (output_style() == EXPANDED) {
      append_optional_linefeed();
      append_indentation();
    } else {
      append_optional_space();
    }
    append_string("}");
    if (node) add_close_mapping(node);
    append_optional_linefeed();
    if (indentation != 0) return;
    if (output_style() != COMPRESSED)
      scheduled_linefeed = 2;
  }

}
