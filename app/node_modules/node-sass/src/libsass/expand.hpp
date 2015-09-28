#ifndef SASS_EXPAND_H
#define SASS_EXPAND_H

#include <map>
#include <vector>
#include <iostream>

#include "ast.hpp"
#include "eval.hpp"
#include "operation.hpp"
#include "environment.hpp"
#include "contextualize.hpp"

namespace Sass {
  using namespace std;

  class Context;
  class Eval;
  class Contextualize_Eval;
  typedef Environment<AST_Node*> Env;
  struct Backtrace;

  class Expand : public Operation_CRTP<Statement*, Expand> {

    Context&          ctx;
    Eval*             eval;
    Contextualize_Eval*    contextualize_eval;
    Env*              env;
    vector<Block*>    block_stack;
    vector<String*>   property_stack;
    vector<Selector*> selector_stack;
    vector<Selector*> at_root_selector_stack;
    bool              in_at_root;
    bool              in_keyframes;
    Backtrace*        backtrace;

    Statement* fallback_impl(AST_Node* n);

  public:
    Expand(Context&, Eval*, Contextualize_Eval*, Env*, Backtrace*);
    virtual ~Expand() { }

    using Operation<Statement*>::operator();

    Statement* operator()(Block*);
    Statement* operator()(Ruleset*);
    Statement* operator()(Propset*);
    Statement* operator()(Media_Block*);
    Statement* operator()(Feature_Block*);
    Statement* operator()(At_Root_Block*);
    Statement* operator()(At_Rule*);
    Statement* operator()(Declaration*);
    Statement* operator()(Assignment*);
    Statement* operator()(Import*);
    Statement* operator()(Import_Stub*);
    Statement* operator()(Warning*);
    Statement* operator()(Error*);
    Statement* operator()(Debug*);
    Statement* operator()(Comment*);
    Statement* operator()(If*);
    Statement* operator()(For*);
    Statement* operator()(Each*);
    Statement* operator()(While*);
    Statement* operator()(Return*);
    Statement* operator()(Extension*);
    Statement* operator()(Definition*);
    Statement* operator()(Mixin_Call*);
    Statement* operator()(Content*);

    template <typename U>
    Statement* fallback(U x) { return fallback_impl(x); }

    void append_block(Block*);
  };

}

#endif
