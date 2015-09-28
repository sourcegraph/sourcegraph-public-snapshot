#include "prelexer.hpp"
#include "backtrace.hpp"
#include "error_handling.hpp"

namespace Sass {

  Sass_Error::Sass_Error(Type type, ParserState pstate, string message)
  : type(type), pstate(pstate), message(message)
  { }

  void warn(string msg, ParserState pstate)
  {
    cerr << "Warning: " << msg<< endl;
  }

  void warn(string msg, ParserState pstate, Backtrace* bt)
  {
    Backtrace top(bt, pstate, "");
    msg += top.to_string();
    warn(msg, pstate);
  }

  void error(string msg, ParserState pstate)
  {
    throw Sass_Error(Sass_Error::syntax, pstate, msg);
  }

  void error(string msg, ParserState pstate, Backtrace* bt)
  {
    Backtrace top(bt, pstate, "");
    msg += "\n" + top.to_string();
    error(msg, pstate);
  }

}
