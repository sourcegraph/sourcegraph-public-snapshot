#ifndef SASS_LISTIZE_H
#define SASS_LISTIZE_H

#include <vector>
#include <iostream>

#include "ast.hpp"
#include "context.hpp"
#include "operation.hpp"
#include "environment.hpp"

namespace Sass {
  using namespace std;

  typedef Environment<AST_Node*> Env;
  struct Backtrace;

  class Listize : public Operation_CRTP<Expression*, Listize> {

    Context&            ctx;

    Expression* fallback_impl(AST_Node* n);

  public:
    Listize(Context&);
    virtual ~Listize() { }

    using Operation<Expression*>::operator();

    Expression* operator()(Selector_List*);
    Expression* operator()(Complex_Selector*);
    Expression* operator()(Compound_Selector*);
    Expression* operator()(Selector_Reference*);

    template <typename U>
    Expression* fallback(U x) { return fallback_impl(x); }
  };

}

#endif
