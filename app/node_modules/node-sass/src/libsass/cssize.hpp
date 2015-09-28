#ifndef SASS_CSSIZE_H
#define SASS_CSSIZE_H

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

  class Cssize : public Operation_CRTP<Statement*, Cssize> {

    Context&            ctx;
    Env*                env;
    vector<Block*>      block_stack;
    vector<Statement*>  p_stack;
    Backtrace*          backtrace;

    Statement* fallback_impl(AST_Node* n);

  public:
    Cssize(Context&, Env*, Backtrace*);
    virtual ~Cssize() { }

    using Operation<Statement*>::operator();

    Statement* operator()(Block*);
    Statement* operator()(Ruleset*);
    // Statement* operator()(Propset*);
    // Statement* operator()(Bubble*);
    Statement* operator()(Media_Block*);
    Statement* operator()(Feature_Block*);
    Statement* operator()(At_Root_Block*);
    Statement* operator()(At_Rule*);
    Statement* operator()(Keyframe_Rule*);
    // Statement* operator()(Declaration*);
    // Statement* operator()(Assignment*);
    // Statement* operator()(Import*);
    // Statement* operator()(Import_Stub*);
    // Statement* operator()(Warning*);
    // Statement* operator()(Error*);
    // Statement* operator()(Comment*);
    // Statement* operator()(If*);
    // Statement* operator()(For*);
    // Statement* operator()(Each*);
    // Statement* operator()(While*);
    // Statement* operator()(Return*);
    // Statement* operator()(Extension*);
    // Statement* operator()(Definition*);
    // Statement* operator()(Mixin_Call*);
    // Statement* operator()(Content*);

    Statement* parent();
    vector<pair<bool, Block*>> slice_by_bubble(Statement*);
    Statement* bubble(At_Rule*);
    Statement* bubble(At_Root_Block*);
    Statement* bubble(Media_Block*);
    Statement* bubble(Feature_Block*);
    Statement* shallow_copy(Statement*);
    Statement* debubble(Block* children, Statement* parent = 0);
    Statement* flatten(Statement*);
    bool bubblable(Statement*);

    List* merge_media_queries(Media_Block*, Media_Block*);
    Media_Query* merge_media_query(Media_Query*, Media_Query*);

    template <typename U>
    Statement* fallback(U x) { return fallback_impl(x); }

    void append_block(Block*);
  };

}

#endif
