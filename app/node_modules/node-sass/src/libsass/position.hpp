#ifndef SASS_POSITION_H
#define SASS_POSITION_H

#include <string>
#include <cstring>
#include <cstdlib>
#include <sstream>
#include <iostream>

namespace Sass {

  using namespace std;

  class Offset {

    public: // c-tor
      Offset(const char* string);
      Offset(const string& text);
      Offset(const size_t line, const size_t column);

      // return new position, incremented by the given string
      Offset add(const char* begin, const char* end);
      Offset inc(const char* begin, const char* end) const;

      // init/create instance from const char substring
      static Offset init(const char* beg, const char* end);

    public: // overload operators for position
      void operator+= (const Offset &pos);
      bool operator== (const Offset &pos) const;
      bool operator!= (const Offset &pos) const;
      Offset operator+ (const Offset &off) const;
      Offset operator- (const Offset &off) const;

    public: // overload output stream operator
      // friend ostream& operator<<(ostream& strm, const Offset& off);

    public:
      Offset off() { return *this; };

    public:
      size_t line;
      size_t column;

  };

  class Position : public Offset {

    public: // c-tor
      Position(const size_t file); // line(0), column(0)
      Position(const size_t file, const Offset& offset);
      Position(const size_t line, const size_t column); // file(-1)
      Position(const size_t file, const size_t line, const size_t column);

    public: // overload operators for position
      void operator+= (const Offset &off);
      bool operator== (const Position &pos) const;
      bool operator!= (const Position &pos) const;
      const Position operator+ (const Offset &off) const;
      const Offset operator- (const Offset &off) const;
      // return new position, incremented by the given string
      Position add(const char* begin, const char* end);
      Position inc(const char* begin, const char* end) const;

    public: // overload output stream operator
      // friend ostream& operator<<(ostream& strm, const Position& pos);

    public:
      size_t file;

  };

  // Token type for representing lexed chunks of text
  class Token {
  public:
    const char* prefix;
    const char* begin;
    const char* end;

    Token()
    : prefix(0), begin(0), end(0) { }
    Token(const char* b, const char* e)
    : prefix(b), begin(b), end(e) { }
    Token(const char* str)
    : prefix(str), begin(str), end(str + strlen(str)) { }
    Token(const char* p, const char* b, const char* e)
    : prefix(p), begin(b), end(e) { }

    size_t length()    const { return end - begin; }
    string ws_before() const { return string(prefix, begin); }
    string to_string() const { return string(begin, end); }
    string time_wspace() const {
      string str(to_string());
      string whitespaces(" \t\f\v\n\r");
      return str.erase(str.find_last_not_of(whitespaces)+1);
    }

    operator bool()   { return begin && end && begin >= end; }
    operator string() { return to_string(); }

    bool operator==(Token t)  { return to_string() == t.to_string(); }
  };

  class ParserState : public Position {

    public: // c-tor
      ParserState(string path, const char* src = 0, const size_t file = string::npos);
      ParserState(string path, const char* src, Position position, Offset offset = Offset(0, 0));
      ParserState(string path, const char* src, Token token, Position position, Offset offset = Offset(0, 0));

    public: // down casts
      Offset off() { return *this; };
      Position pos() { return *this; };
      ParserState pstate() { return *this; };

    public:
      string path;
      const char* src;
      Offset offset;
      Token token;

  };

}

#endif
