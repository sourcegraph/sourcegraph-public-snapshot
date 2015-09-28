#include "ast.hpp"
#include "output.hpp"
#include "to_string.hpp"

namespace Sass {
  using namespace std;

  Output::Output(Context* ctx)
  : Inspect(Emitter(ctx)),
    charset(""),
    top_nodes(0)
  {}

  Output::~Output() { }

  void Output::fallback_impl(AST_Node* n)
  {
    return n->perform(this);
  }

  void Output::operator()(Import* imp)
  {
    top_nodes.push_back(imp);
  }

  OutputBuffer Output::get_buffer(void)
  {

    Emitter emitter(ctx);
    Inspect inspect(emitter);

    size_t size_nodes = top_nodes.size();
    for (size_t i = 0; i < size_nodes; i++) {
      top_nodes[i]->perform(&inspect);
      inspect.append_mandatory_linefeed();
    }

    // flush scheduled outputs
    inspect.finalize();
    // prepend buffer on top
    prepend_output(inspect.output());
    // make sure we end with a linefeed
    if (!ends_with(wbuf.buffer, ctx->linefeed)) {
      // if the output is not completely empty
      if (!wbuf.buffer.empty()) append_string(ctx->linefeed);
    }

    // search for unicode char
    for(const char& chr : wbuf.buffer) {
      // skip all ascii chars
      if (chr >= 0) continue;
      // declare the charset
      if (output_style() != COMPRESSED)
        charset = "@charset \"UTF-8\";"
                  + ctx->linefeed;
      else charset = "\xEF\xBB\xBF";
      // abort search
      break;
    }

    // add charset as first line, before comments and imports
    if (!charset.empty()) prepend_string(charset);

    return wbuf;

  }

  void Output::operator()(Comment* c)
  {
    To_String to_string(ctx);
    string txt = c->text()->perform(&to_string);
    // if (indentation && txt == "/**/") return;
    bool important = c->is_important();
    if (output_style() != COMPRESSED || important) {
      if (buffer().size() == 0) {
        top_nodes.push_back(c);
      } else {
        in_comment = true;
        append_indentation();
        c->text()->perform(this);
        in_comment = false;
        if (indentation == 0) {
          append_mandatory_linefeed();
        } else {
          append_optional_linefeed();
        }
      }
    }
  }

  void Output::operator()(Ruleset* r)
  {
    Selector* s     = r->selector();
    Block*    b     = r->block();
    bool      decls = false;

    // Filter out rulesets that aren't printable (process its children though)
    if (!Util::isPrintable(r, output_style())) {
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (dynamic_cast<Has_Block*>(stm)) {
          stm->perform(this);
        }
      }
      return;
    }

    if (b->has_non_hoistable()) {
      decls = true;
      if (output_style() == NESTED) indentation += r->tabs();
      if (ctx && ctx->source_comments) {
        stringstream ss;
        append_indentation();
        ss << "/* line " << r->pstate().line+1 << ", " << r->pstate().path << " */";
        append_string(ss.str());
        append_optional_linefeed();
      }
      s->perform(this);
      append_scope_opener(b);
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        bool bPrintExpression = true;
        // Check print conditions
        if (typeid(*stm) == typeid(Declaration)) {
          Declaration* dec = static_cast<Declaration*>(stm);
          if (dec->value()->concrete_type() == Expression::STRING) {
            String_Constant* valConst = static_cast<String_Constant*>(dec->value());
            string val(valConst->value());
            if (dynamic_cast<String_Quoted*>(valConst)) {
              if (!valConst->quote_mark() && val.empty()) {
                bPrintExpression = false;
              }
            }
          }
          else if (dec->value()->concrete_type() == Expression::LIST) {
            List* list = static_cast<List*>(dec->value());
            bool all_invisible = true;
            for (size_t list_i = 0, list_L = list->length(); list_i < list_L; ++list_i) {
              Expression* item = (*list)[list_i];
              if (!item->is_invisible()) all_invisible = false;
            }
            if (all_invisible) bPrintExpression = false;
          }
        }
        // Print if OK
        if (!stm->is_hoistable() && bPrintExpression) {
          stm->perform(this);
        }
      }
      if (output_style() == NESTED) indentation -= r->tabs();
      append_scope_closer(b);
    }

    if (b->has_hoistable()) {
      if (decls) ++indentation;
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (stm->is_hoistable()) {
          stm->perform(this);
        }
      }
      if (decls) --indentation;
    }
  }

  void Output::operator()(Keyframe_Rule* r)
  {
    Block* b = r->block();
    Selector* v = r->selector();

    if (v) {
      v->perform(this);
    }

    if (!b) {
      append_colon_separator();
      return;
    }

    append_scope_opener();
    for (size_t i = 0, L = b->length(); i < L; ++i) {
      Statement* stm = (*b)[i];
      if (!stm->is_hoistable()) {
        stm->perform(this);
        if (i < L - 1) append_special_linefeed();
      }
    }

    for (size_t i = 0, L = b->length(); i < L; ++i) {
      Statement* stm = (*b)[i];
      if (stm->is_hoistable()) {
        stm->perform(this);
      }
    }

    append_scope_closer();
  }

  void Output::operator()(Feature_Block* f)
  {
    if (f->is_invisible()) return;

    Feature_Query* q    = f->feature_queries();
    Block* b            = f->block();

    // Filter out feature blocks that aren't printable (process its children though)
    if (!Util::isPrintable(f, output_style())) {
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (dynamic_cast<Has_Block*>(stm)) {
          stm->perform(this);
        }
      }
      return;
    }

    if (output_style() == NESTED) indentation += f->tabs();
    append_indentation();
    append_token("@supports", f);
    append_mandatory_space();
    q->perform(this);
    append_scope_opener();

    Selector* e = f->selector();
    if (e && b->has_non_hoistable()) {
      // JMA - hoisted, output the non-hoistable in a nested block, followed by the hoistable
      e->perform(this);
      append_scope_opener();

      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (!stm->is_hoistable()) {
          stm->perform(this);
        }
      }

      append_scope_closer();

      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (stm->is_hoistable()) {
          stm->perform(this);
        }
      }
    }
    else {
      // JMA - not hoisted, just output in order
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        stm->perform(this);
        if (i < L - 1) append_special_linefeed();
      }
    }

    if (output_style() == NESTED) indentation -= f->tabs();

    append_scope_closer();

  }

  void Output::operator()(Media_Block* m)
  {
    if (m->is_invisible()) return;

    List*  q     = m->media_queries();
    Block* b     = m->block();

    // Filter out media blocks that aren't printable (process its children though)
    if (!Util::isPrintable(m, output_style())) {
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (dynamic_cast<Has_Block*>(stm)) {
          stm->perform(this);
        }
      }
      return;
    }
    if (output_style() == NESTED) indentation += m->tabs();
    append_indentation();
    append_token("@media", m);
    append_mandatory_space();
    in_media_block = true;
    q->perform(this);
    in_media_block = false;
    append_scope_opener();

    Selector* e = m->selector();
    if (e && b->has_non_hoistable()) {
      // JMA - hoisted, output the non-hoistable in a nested block, followed by the hoistable
      e->perform(this);
      append_scope_opener();

      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (!stm->is_hoistable()) {
          stm->perform(this);
        }
      }

      append_scope_closer();

      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (stm->is_hoistable()) {
          stm->perform(this);
        }
      }
    }
    else {
      // JMA - not hoisted, just output in order
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        stm->perform(this);
        if (i < L - 1) append_special_linefeed();
      }
    }

    if (output_style() == NESTED) indentation -= m->tabs();
    append_scope_closer();
  }

  void Output::operator()(At_Rule* a)
  {
    string      kwd   = a->keyword();
    Selector*   s     = a->selector();
    Expression* v     = a->value();
    Block*      b     = a->block();

    append_indentation();
    append_token(kwd, a);
    if (s) {
      append_mandatory_space();
      in_wrapped = true;
      s->perform(this);
      in_wrapped = false;
    }
    else if (v) {
      append_mandatory_space();
      v->perform(this);
    }
    if (!b) {
      append_delimiter();
      return;
    }

    if (b->is_invisible() || b->length() == 0) {
      return append_string(" {}");
    }

    append_scope_opener();

    bool format = kwd != "@font-face";;

    for (size_t i = 0, L = b->length(); i < L; ++i) {
      Statement* stm = (*b)[i];
      if (!stm->is_hoistable()) {
        stm->perform(this);
        if (i < L - 1 && format) append_special_linefeed();
      }
    }

    for (size_t i = 0, L = b->length(); i < L; ++i) {
      Statement* stm = (*b)[i];
      if (stm->is_hoistable()) {
        stm->perform(this);
        if (i < L - 1 && format) append_special_linefeed();
      }
    }

    append_scope_closer();
  }

  void Output::operator()(String_Quoted* s)
  {
    if (s->quote_mark()) {
      append_token(quote(s->value(), s->quote_mark()), s);
    } else if (!in_comment) {
      append_token(string_to_output(s->value()), s);
    } else {
      append_token(s->value(), s);
    }
  }

  void Output::operator()(String_Constant* s)
  {
    if (String_Quoted* quoted = dynamic_cast<String_Quoted*>(s)) {
      return Output::operator()(quoted);
    } else {
      string value(s->value());
      if (s->can_compress_whitespace() && output_style() == COMPRESSED) {
        value.erase(std::remove_if(value.begin(), value.end(), ::isspace), value.end());
      }
      if (!in_comment) {
        append_token(string_to_output(value), s);
      } else {
        append_token(value, s);
      }
    }
  }

}
