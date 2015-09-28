#include <cmath>
#include <sstream>
#include <iomanip>
#include <iostream>

#include "ast.hpp"
#include "inspect.hpp"
#include "context.hpp"
#include "to_string.hpp"

namespace Sass {
  using namespace std;

  To_String::To_String(Context* ctx, bool in_declaration)
  : ctx(ctx), in_declaration(in_declaration) { }
  To_String::~To_String() { }

  inline string To_String::fallback_impl(AST_Node* n)
  {
    Emitter emitter(ctx);
    Inspect i(emitter);
    i.in_declaration = in_declaration;
    n->perform(&i);
    return i.get_buffer();
  }

  inline string To_String::operator()(String_Constant* s)
  {
    return s->value();
  }

  inline string To_String::operator()(Null* n)
  { return ""; }
}
