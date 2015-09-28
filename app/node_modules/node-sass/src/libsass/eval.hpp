#ifndef SASS_EVAL_H
#define SASS_EVAL_H

#include <iostream>

#include "context.hpp"
#include "position.hpp"
#include "operation.hpp"
#include "environment.hpp"
#include "contextualize.hpp"
#include "listize.hpp"
#include "sass_values.h"

namespace Sass {
  using namespace std;

  typedef Environment<AST_Node*> Env;
  struct Backtrace;
  class Contextualize;
  class Listize;

  class Eval : public Operation_CRTP<Expression*, Eval> {

    Context&   ctx;

    Expression* fallback_impl(AST_Node* n);

  public:
    Contextualize* contextualize;
    Listize*   listize;
    Env*       env;
    Backtrace* backtrace;
    Eval(Context&, Contextualize*, Listize*, Env*, Backtrace*);
    virtual ~Eval();
    Eval* with(Env* e, Backtrace* bt); // for setting the env before eval'ing an expression
    Eval* with(Selector* c, Env* e, Backtrace* bt, Selector* placeholder = 0, Selector* extender = 0); // for setting the env before eval'ing an expression
    using Operation<Expression*>::operator();

    // for evaluating function bodies
    Expression* operator()(Block*);
    Expression* operator()(Assignment*);
    Expression* operator()(If*);
    Expression* operator()(For*);
    Expression* operator()(Each*);
    Expression* operator()(While*);
    Expression* operator()(Return*);
    Expression* operator()(Warning*);
    Expression* operator()(Error*);
    Expression* operator()(Debug*);

    Expression* operator()(List*);
    Expression* operator()(Map*);
    Expression* operator()(Binary_Expression*);
    Expression* operator()(Unary_Expression*);
    Expression* operator()(Function_Call*);
    Expression* operator()(Function_Call_Schema*);
    Expression* operator()(Variable*);
    Expression* operator()(Textual*);
    Expression* operator()(Number*);
    Expression* operator()(Boolean*);
    Expression* operator()(String_Schema*);
    Expression* operator()(String_Constant*);
    Expression* operator()(Media_Query*);
    Expression* operator()(Media_Query_Expression*);
    Expression* operator()(At_Root_Expression*);
    Expression* operator()(Feature_Query*);
    Expression* operator()(Feature_Query_Condition*);
    Expression* operator()(Null*);
    Expression* operator()(Argument*);
    Expression* operator()(Arguments*);
    Expression* operator()(Comment*);
    Expression* operator()(Parent_Selector* p);

    template <typename U>
    Expression* fallback(U x) { return fallback_impl(x); }

  private:
    string interpolation(Expression* s);

  };

  Expression* cval_to_astnode(Sass_Value* v, Context& ctx, Backtrace* backtrace, ParserState pstate = ParserState("[AST]"));

  bool eq(Expression*, Expression*, Context&);
  bool lt(Expression*, Expression*, Context&);
}

#endif
