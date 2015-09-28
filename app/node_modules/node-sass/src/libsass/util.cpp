#include<stdint.h>
#include "ast.hpp"
#include "util.hpp"
#include "prelexer.hpp"
#include "utf8/checked.h"

namespace Sass {

  #define out_of_memory() do {                    \
      fprintf(stderr, "Out of memory.\n");    \
      exit(EXIT_FAILURE);                     \
    } while (0)

  /* Sadly, sass_strdup is not portable. */
  char *sass_strdup(const char *str)
  {
    char *ret = (char*) malloc(strlen(str) + 1);
    if (ret == NULL)
      out_of_memory();
    strcpy(ret, str);
    return ret;
  }

  /* Locale unspecific atof function. */
  double sass_atof(const char *str)
  {
    char separator = *(localeconv()->decimal_point);
    if(separator != '.'){
      // The current locale specifies another
      // separator. convert the separator to the
      // one understood by the locale if needed
      const char *found = strchr(str, '.');
      if(found != NULL){
        // substitution is required. perform the substitution on a copy
        // of the string. This is slower but it is thread safe.
        char *copy = sass_strdup(str);
        *(copy + (found - str)) = separator;
        double res = atof(copy);
        free(copy);
        return res;
      }
    }

    return atof(str);
  }

  string string_eval_escapes(const string& s)
  {

    string out("");
    bool esc = false;
    for (size_t i = 0, L = s.length(); i < L; ++i) {
      if(s[i] == '\\' && esc == false) {
        esc = true;

        // escape length
        size_t len = 1;

        // parse as many sequence chars as possible
        // ToDo: Check if ruby aborts after possible max
        while (i + len < L && s[i + len] && isxdigit(s[i + len])) ++ len;

        // hex string?
        if (len > 1) {

          // convert the extracted hex string to code point value
          // ToDo: Maybe we could do this without creating a substring
          uint32_t cp = strtol(s.substr (i + 1, len - 1).c_str(), nullptr, 16);

          if (cp == 0) cp = 0xFFFD;

          // assert invalid code points
          if (cp >= 1) {

            // use a very simple approach to convert via utf8 lib
            // maybe there is a more elegant way; maybe we shoud
            // convert the whole output from string to a stream!?
            // allocate memory for utf8 char and convert to utf8
            unsigned char u[5] = {0,0,0,0,0}; utf8::append(cp, u);
            for(size_t m = 0; u[m] && m < 5; m++) out.push_back(u[m]);

            // skip some more chars?
            i += len - 1; esc = false;
            if (cp == 10) out += ' ';

          }

        }

      }
      else {
        out += s[i];
        esc = false;
      }
    }
    return out;

  }

  // double escape every escape sequences
  // escape unescaped quotes and backslashes
  string string_escape(const string& str)
  {
    string out("");
    for (auto i : str) {
      // escape some characters
      if (i == '"') out += '\\';
      if (i == '\'') out += '\\';
      if (i == '\\') out += '\\';
      out += i;
    }
    return out;
  }

  // unescape every escape sequence
  // only removes unescaped backslashes
  string string_unescape(const string& str)
  {
    string out("");
    bool esc = false;
    for (auto i : str) {
      if (esc || i != '\\') {
        esc = false;
        out += i;
      } else {
        esc = true;
      }
    }
    // open escape sequence at end
    // maybe it should thow an error
    if (esc) { out += '\\'; }
    return out;
  }

  // read css string (handle multiline DELIM)
  string read_css_string(const string& str)
  {
    string out("");
    bool esc = false;
    for (auto i : str) {
      if (i == '\\') {
        esc = ! esc;
      } else if (esc && i == '\r') {
        continue;
      } else if (esc && i == '\n') {
        out.resize (out.size () - 1);
        esc = false;
        continue;
      } else {
        esc = false;
      }
      out.push_back(i);
    }
    if (esc) out += '\\';
    return out;
  }

  // evacuate unescaped quoted
  // leave everything else untouched
  string evacuate_quotes(const string& str)
  {
    string out("");
    bool esc = false;
    for (auto i : str) {
      if (!esc) {
        // ignore next character
        if (i == '\\') esc = true;
        // evacuate unescaped quotes
        else if (i == '"') out += '\\';
        else if (i == '\'') out += '\\';
      }
      // get escaped char now
      else { esc = false; }
      // remove nothing
      out += i;
    }
    return out;
  }

  // double escape all escape sequences
  // keep unescaped quotes and backslashes
  string evacuate_escapes(const string& str)
  {
    string out("");
    bool esc = false;
    for (auto i : str) {
      if (i == '\\' && !esc) {
        out += '\\';
        out += '\\';
        esc = true;
      } else if (esc && i == '"') {
        out += '\\';
        out += i;
        esc = false;
      } else if (esc && i == '\'') {
        out += '\\';
        out += i;
        esc = false;
      } else if (esc && i == '\\') {
        out += '\\';
        out += i;
        esc = false;
      } else {
        esc = false;
        out += i;
      }
    }
    // happens when parsing does not correctly skip
    // over escaped sequences for ie. interpolations
    // one example: foo\#{interpolate}
    // if (esc) out += '\\';
    return out;
  }

  // bell character is replaces with space
  string string_to_output(const string& str)
  {
    string out("");
    for (auto i : str) {
      if (i == 10) {
        out += ' ';
      } else {
        out += i;
      }
    }
    return out;
  }

  string comment_to_string(const string& text)
  {
    string str = "";
    size_t has = 0;
    char prev = 0;
    bool clean = false;
    for (auto i : text) {
      if (clean) {
        if (i == '\n') { has = 0; }
        else if (i == '\r') { has = 0; }
        else if (i == '\t') { ++ has; }
        else if (i == ' ') { ++ has; }
        else if (i == '*') {}
        else {
          clean = false;
          str += ' ';
          if (prev == '*' && i == '/') str += "*/";
          else str += i;
        }
      } else if (i == '\n') {
        clean = true;
      } else if (i == '\r') {
        clean = true;
      } else {
        str += i;
      }
      prev = i;
    }
    if (has) return str;
    else return text;
  }

   string normalize_wspace(const string& str)
  {
    bool ws = false;
    bool esc = false;
    string text = "";
    for(const char& i : str) {
      if (!esc && i == '\\') {
        esc = true;
        ws = false;
        text += i;
      } else if (esc) {
        esc = false;
        ws = false;
        text += i;
      } else if (
        i == ' ' ||
        i == '\r' ||
        i == '\n' ||
        i == '	'
      ) {
        // only add one space
        if (!ws) text += ' ';
        ws = true;
      } else {
        ws = false;
        text += i;
      }
    }
    if (esc) text += '\\';
    return text;
  }

  // find best quote_mark by detecting if the string contains any single
  // or double quotes. When a single quote is found, we not we want a double
  // quote as quote_mark. Otherwise we check if the string cotains any double
  // quotes, which will trigger the use of single quotes as best quote_mark.
  char detect_best_quotemark(const char* s, char qm)
  {
    // ensure valid fallback quote_mark
    char quote_mark = qm && qm != '*' ? qm : '"';
    while (*s) {
      // force double quotes as soon
      // as one single quote is found
      if (*s == '\'') { return '"'; }
      // a single does not force quote_mark
      // maybe we see a double quote later
      else if (*s == '"') { quote_mark = '\''; }
      ++ s;
    }
    return quote_mark;
  }

  string unquote(const string& s, char* qd)
  {

    // not enough room for quotes
    // no possibility to unquote
    if (s.length() < 2) return s;

    char q;
    bool skipped = false;

    // this is no guarantee that the unquoting will work
    // what about whitespace before/after the quote_mark?
    if      (*s.begin() == '"'  && *s.rbegin() == '"')  q = '"';
    else if (*s.begin() == '\'' && *s.rbegin() == '\'') q = '\'';
    else                                                return s;

    string unq;
    unq.reserve(s.length()-2);

    for (size_t i = 1, L = s.length() - 1; i < L; ++i) {

      // implement the same strange ruby sass behavior
      // an escape sequence can also mean a unicode char
      if (s[i] == '\\' && !skipped) {
        // remember
        skipped = true;

        // skip it
        // ++ i;

        // if (i == L) break;

        // escape length
        size_t len = 1;

        // parse as many sequence chars as possible
        // ToDo: Check if ruby aborts after possible max
        while (i + len < L && s[i + len] && isxdigit(s[i + len])) ++ len;

        // hex string?
        if (len > 1) {

          // convert the extracted hex string to code point value
          // ToDo: Maybe we could do this without creating a substring
          uint32_t cp = strtol(s.substr (i + 1, len - 1).c_str(), nullptr, 16);

          // assert invalid code points
          if (cp == 0) cp = 0xFFFD;
          // replace bell character
          // if (cp == 10) cp = 32;

          // use a very simple approach to convert via utf8 lib
          // maybe there is a more elegant way; maybe we shoud
          // convert the whole output from string to a stream!?
          // allocate memory for utf8 char and convert to utf8
          unsigned char u[5] = {0,0,0,0,0}; utf8::append(cp, u);
          for(size_t m = 0; u[m] && m < 5; m++) unq.push_back(u[m]);

          // skip some more chars?
          i += len - 1; skipped = false;

        }


      }
      // check for unexpected delimiter
      // be strict and throw error back
      // else if (!skipped && q == s[i]) {
      //   // don't be that strict
      //   return s;
      //   // this basically always means an internal error and not users fault
      //   error("Unescaped delimiter in string to unquote found. [" + s + "]", ParserState("[UNQUOTE]"));
      // }
      else {
        skipped = false;
        unq.push_back(s[i]);
      }

    }
    if (skipped) { return s; }
    if (qd) *qd = q;
    return unq;

  }

  string quote(const string& s, char q, bool keep_linefeed_whitespace)
  {

    // autodetect with fallback to given quote
    q = detect_best_quotemark(s.c_str(), q);

    // return an empty quoted string
    if (s.empty()) return string(2, q ? q : '"');

    string quoted;
    quoted.reserve(s.length()+2);
    quoted.push_back(q);

    const char* it = s.c_str();
    const char* end = it + strlen(it) + 1;
    while (*it && it < end) {
      const char* now = it;

      if (*it == q) {
        quoted.push_back('\\');
      } else if (*it == '\\') {
        quoted.push_back('\\');
      }

      int cp = utf8::next(it, end);

      if (cp == 10) {
        quoted.push_back('\\');
        quoted.push_back('a');
        // we hope we can remove this flag once we figure out
        // why ruby sass has these different output behaviors
        if (keep_linefeed_whitespace)
          quoted.push_back(' ');
      } else if (cp < 127) {
        quoted.push_back((char) cp);
      } else {
        while (now < it) {
          quoted.push_back(*now);
          ++ now;
        }
      }
    }

    quoted.push_back(q);
    return quoted;
  }

  bool is_hex_doublet(double n)
  {
    return n == 0x00 || n == 0x11 || n == 0x22 || n == 0x33 ||
           n == 0x44 || n == 0x55 || n == 0x66 || n == 0x77 ||
           n == 0x88 || n == 0x99 || n == 0xAA || n == 0xBB ||
           n == 0xCC || n == 0xDD || n == 0xEE || n == 0xFF ;
  }

  bool is_color_doublet(double r, double g, double b)
  {
    return is_hex_doublet(r) && is_hex_doublet(g) && is_hex_doublet(b);
  }

  bool peek_linefeed(const char* start)
  {
    while (*start) {
      if (*start == '\n' || *start == '\r') return true;
      if (*start != ' ' && *start != '\t') return false;
      ++ start;
    }
    return false;
  }

  namespace Util {
    using std::string;

    string normalize_underscores(const string& str) {
      string normalized = str;
      for(size_t i = 0, L = normalized.length(); i < L; ++i) {
        if(normalized[i] == '_') {
          normalized[i] = '-';
        }
      }
      return normalized;
    }

    string normalize_decimals(const string& str) {
      string prefix = "0";
      string normalized = str;

      return normalized[0] == '.' ? normalized.insert(0, prefix) : normalized;
    }

    // compress a color sixtuplet if possible
    // input: "#CC9900" -> output: "#C90"
    string normalize_sixtuplet(const string& col) {
      if(
        col.substr(1, 1) == col.substr(2, 1) &&
        col.substr(3, 1) == col.substr(4, 1) &&
        col.substr(5, 1) == col.substr(6, 1)
      ) {
        return string("#" + col.substr(1, 1)
                          + col.substr(3, 1)
                          + col.substr(5, 1));
      } else {
        return string(col);
      }
    }

    bool isPrintable(Ruleset* r, Output_Style style) {
      if (r == NULL) {
        return false;
      }

      Block* b = r->block();

      bool hasSelectors = static_cast<Selector_List*>(r->selector())->length() > 0;

      if (!hasSelectors) {
        return false;
      }

      bool hasDeclarations = false;
      bool hasPrintableChildBlocks = false;
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (dynamic_cast<Has_Block*>(stm)) {
          Block* pChildBlock = ((Has_Block*)stm)->block();
          if (isPrintable(pChildBlock, style)) {
            hasPrintableChildBlocks = true;
          }
        } else if (Comment* c = dynamic_cast<Comment*>(stm)) {
          if (style == COMPRESSED) {
            hasDeclarations = c->is_important();
          } else {
            hasDeclarations = true;
          }
        } else if (Declaration* d = dynamic_cast<Declaration*>(stm)) {
          return isPrintable(d, style);
        } else {
          hasDeclarations = true;
        }

        if (hasDeclarations || hasPrintableChildBlocks) {
          return true;
        }
      }

      return false;
    }

    bool isPrintable(String_Constant* s, Output_Style style)
    {
      return ! s->value().empty();
    }

    bool isPrintable(String_Quoted* s, Output_Style style)
    {
      return true;
    }

    bool isPrintable(Declaration* d, Output_Style style)
    {
      Expression* val = d->value();
      if (String_Quoted* sq = dynamic_cast<String_Quoted*>(val)) return isPrintable(sq, style);
      if (String_Constant* sc = dynamic_cast<String_Constant*>(val)) return isPrintable(sc, style);
      return true;
    }

    bool isPrintable(Expression* e, Output_Style style) {
      return isPrintable(e, style);
    }

    bool isPrintable(Feature_Block* f, Output_Style style) {
      if (f == NULL) {
        return false;
      }

      Block* b = f->block();

      bool hasSelectors = f->selector() && static_cast<Selector_List*>(f->selector())->length() > 0;

      bool hasDeclarations = false;
      bool hasPrintableChildBlocks = false;
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (!stm->is_hoistable() && f->selector() != NULL && !hasSelectors) {
          // If a statement isn't hoistable, the selectors apply to it. If there are no selectors (a selector list of length 0),
          // then those statements aren't considered printable. That means there was a placeholder that was removed. If the selector
          // is NULL, then that means there was never a wrapping selector and it is printable (think of a top level media block with
          // a declaration in it).
        }
        else if (typeid(*stm) == typeid(Declaration) || typeid(*stm) == typeid(At_Rule)) {
          hasDeclarations = true;
        }
        else if (dynamic_cast<Has_Block*>(stm)) {
          Block* pChildBlock = ((Has_Block*)stm)->block();
          if (isPrintable(pChildBlock, style)) {
            hasPrintableChildBlocks = true;
          }
        }

        if (hasDeclarations || hasPrintableChildBlocks) {
          return true;
        }
      }

      return false;
    }

    bool isPrintable(Media_Block* m, Output_Style style) {
      if (m == NULL) {
        return false;
      }

      Block* b = m->block();

      bool hasSelectors = m->selector() && static_cast<Selector_List*>(m->selector())->length() > 0;

      bool hasDeclarations = false;
      bool hasPrintableChildBlocks = false;
      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (!stm->is_hoistable() && m->selector() != NULL && !hasSelectors) {
          // If a statement isn't hoistable, the selectors apply to it. If there are no selectors (a selector list of length 0),
          // then those statements aren't considered printable. That means there was a placeholder that was removed. If the selector
          // is NULL, then that means there was never a wrapping selector and it is printable (think of a top level media block with
          // a declaration in it).
        }
        else if (typeid(*stm) == typeid(Declaration) || typeid(*stm) == typeid(At_Rule)) {
          hasDeclarations = true;
        }
        else if (dynamic_cast<Has_Block*>(stm)) {
          Block* pChildBlock = ((Has_Block*)stm)->block();
          if (isPrintable(pChildBlock, style)) {
            hasPrintableChildBlocks = true;
          }
        }

        if (hasDeclarations || hasPrintableChildBlocks) {
          return true;
        }
      }

      return false;
    }

    bool isPrintable(Block* b, Output_Style style) {
      if (b == NULL) {
        return false;
      }

      for (size_t i = 0, L = b->length(); i < L; ++i) {
        Statement* stm = (*b)[i];
        if (typeid(*stm) == typeid(Declaration) || typeid(*stm) == typeid(At_Rule)) {
          return true;
        }
        else if (typeid(*stm) == typeid(Comment)) {

        }
        else if (typeid(*stm) == typeid(Ruleset)) {
          Ruleset* r = (Ruleset*) stm;
          if (isPrintable(r, style)) {
            return true;
          }
        }
        else if (typeid(*stm) == typeid(Feature_Block)) {
          Feature_Block* f = (Feature_Block*) stm;
          if (isPrintable(f, style)) {
            return true;
          }
        }
        else if (typeid(*stm) == typeid(Media_Block)) {
          Media_Block* m = (Media_Block*) stm;
          if (isPrintable(m, style)) {
            return true;
          }
        }
        else if (dynamic_cast<Has_Block*>(stm) && isPrintable(((Has_Block*)stm)->block(), style)) {
          return true;
        }
      }

      return false;
    }

    string vecJoin(const vector<string>& vec, const string& sep)
    {
      switch (vec.size())
      {
        case 0:
            return string("");
        case 1:
            return vec[0];
        default:
            std::ostringstream os;
            os << vec[0];
            for (size_t i = 1; i < vec.size(); i++) {
              os << sep << vec[i];
            }
            return os.str();
      }
    }

     bool isAscii(const char chr) {
       return unsigned(chr) < 128;
     }

  }
}
