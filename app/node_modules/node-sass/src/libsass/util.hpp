#ifndef SASS_UTIL_H
#define SASS_UTIL_H

#include <cstdio>
#include <vector>
#include <string>
#include "ast_fwd_decl.hpp"

namespace Sass {
  using namespace std;

  char* sass_strdup(const char* str);
  double sass_atof(const char* str);
  string string_escape(const string& str);
  string string_unescape(const string& str);
  string string_eval_escapes(const string& str);
  string read_css_string(const string& str);
  string evacuate_quotes(const string& str);
  string evacuate_escapes(const string& str);
  string string_to_output(const string& str);
  string comment_to_string(const string& text);
  string normalize_wspace(const string& str);

  string quote(const string&, char q = 0, bool keep_linefeed_whitespace = false);
  string unquote(const string&, char* q = 0);
  char detect_best_quotemark(const char* s, char qm = '"');

  bool is_hex_doublet(double n);
  bool is_color_doublet(double r, double g, double b);

  bool peek_linefeed(const char* start);

  namespace Util {

    string normalize_underscores(const string& str);
    string normalize_decimals(const string& str);
    string normalize_sixtuplet(const string& col);

    string vecJoin(const vector<string>& vec, const string& sep);
    bool containsAnyPrintableStatements(Block* b);

    bool isPrintable(Ruleset* r, Output_Style style = NESTED);
    bool isPrintable(Feature_Block* r, Output_Style style = NESTED);
    bool isPrintable(Media_Block* r, Output_Style style = NESTED);
    bool isPrintable(Block* b, Output_Style style = NESTED);
    bool isPrintable(String_Constant* s, Output_Style style = NESTED);
    bool isPrintable(String_Quoted* s, Output_Style style = NESTED);
    bool isPrintable(Declaration* d, Output_Style style = NESTED);
    bool isPrintable(Expression* e, Output_Style style = NESTED);
    bool isAscii(const char chr);

  }
}
#endif
