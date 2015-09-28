#ifndef SASS_TO_STRING_H
#define SASS_TO_STRING_H

#include <string>

#include "operation.hpp"

namespace Sass {
  using namespace std;

  class Context;
  class Null;

  class To_String : public Operation_CRTP<string, To_String> {
    // import all the class-specific methods and override as desired
    using Operation<string>::operator();
    // override this to define a catch-all
    string fallback_impl(AST_Node* n);

    Context* ctx;
    bool in_declaration;

  public:
    To_String(Context* ctx = 0, bool in_declaration = true);
    virtual ~To_String();

    string operator()(Null* n);
    string operator()(String_Constant*);

    template <typename U>
    string fallback(U n) { return fallback_impl(n); }
  };
}

#endif
