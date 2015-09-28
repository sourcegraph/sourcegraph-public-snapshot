#ifndef SASS_CONTEXTUALIZE_H
#define SASS_CONTEXTUALIZE_H

#include "context.hpp"
#include "operation.hpp"
#include "environment.hpp"
#include "ast_fwd_decl.hpp"

namespace Sass {
  struct Backtrace;

  typedef Environment<AST_Node*> Env;

  class Contextualize : public Operation_CRTP<Selector*, Contextualize> {


  public:
    Context&   ctx;
    Env*       env;
    Backtrace* backtrace;
    Selector*  parent;
    Selector* placeholder;
    Selector* extender;

    Selector* fallback_impl(AST_Node* n);
    Contextualize(Context&, Env*, Backtrace*, Selector* placeholder = 0, Selector* extender = 0);
    virtual ~Contextualize();
    Contextualize* with(Selector*, Env*, Backtrace*, Selector* placeholder = 0, Selector* extender = 0);
    using Operation<Selector*>::operator();

    Selector* operator()(Selector_List*);
    Selector* operator()(Complex_Selector*);
    Selector* operator()(Compound_Selector*);
    Selector* operator()(Wrapped_Selector*);
    Selector* operator()(Pseudo_Selector*);
    Selector* operator()(Selector_Qualifier*);
    Selector* operator()(Type_Selector*);
    Selector* operator()(Selector_Placeholder*);
    Selector* operator()(Selector_Reference*);

    template <typename U>
    Selector* fallback(U x) { return fallback_impl(x); }
  };
}

#endif
