#ifndef SASS_ERROR_HANDLING_H
#define SASS_ERROR_HANDLING_H

#include <string>

#include "position.hpp"

namespace Sass {
  using namespace std;

  struct Backtrace;

  struct Sass_Error {
    enum Type { read, write, syntax, evaluation };

    Type type;
    ParserState pstate;
    string message;

    Sass_Error(Type type, ParserState pstate, string message);

  };

  void warn(string msg, ParserState pstate);
  void warn(string msg, ParserState pstate, Backtrace* bt);

  void error(string msg, ParserState pstate);
  void error(string msg, ParserState pstate, Backtrace* bt);

}

#endif