#include "contextualize_eval.hpp"
#include "ast.hpp"
#include "eval.hpp"
#include "backtrace.hpp"
#include "to_string.hpp"
#include "parser.hpp"

namespace Sass {

  Contextualize_Eval::Contextualize_Eval(Context& ctx, Eval* eval, Env* env, Backtrace* bt)
  : Contextualize(ctx, env, bt), eval(eval)
  { }

  Contextualize_Eval::~Contextualize_Eval() { }

  Selector* Contextualize_Eval::fallback_impl(AST_Node* n)
  {
    return Contextualize::fallback_impl(n);
  }

  Contextualize_Eval* Contextualize_Eval::with(Selector* s, Env* e, Backtrace* bt, Selector* p, Selector* ex)
  {
    Contextualize::with(s, e, bt, p, ex);
    eval = eval->with(s, e, bt, p, ex);
    return this;
  }

  Selector* Contextualize_Eval::operator()(Selector_Schema* s)
  {
    To_String to_string;
    string result_str(s->contents()->perform(eval)->perform(&to_string));
    result_str += '{'; // the parser looks for a brace to end the selector
    Selector* result_sel = Parser::from_c_str(result_str.c_str(), ctx, s->pstate()).parse_selector_group();
    return result_sel->perform(this);
  }

  Selector* Contextualize_Eval::operator()(Selector_List* s)
  {
    return Contextualize::operator ()(s);
  }

  Selector* Contextualize_Eval::operator()(Complex_Selector* s)
  {
    return Contextualize::operator ()(s);
  }

  Selector* Contextualize_Eval::operator()(Compound_Selector* s)
  {
    return Contextualize::operator ()(s);
  }

  Selector* Contextualize_Eval::operator()(Wrapped_Selector* s)
  {
    return Contextualize::operator ()(s);
  }

  Selector* Contextualize_Eval::operator()(Pseudo_Selector* s)
  {
    return Contextualize::operator ()(s);
  }

  Selector* Contextualize_Eval::operator()(Attribute_Selector* s)
  {
    // the value might be interpolated; evaluate it
    String* v = s->value();
    if (v && eval) {
     Eval* eval_with = eval->with(parent, env, backtrace);
     v = static_cast<String*>(v->perform(eval_with));
    }
    To_String toString;
    Attribute_Selector* ss = new (ctx.mem) Attribute_Selector(*s);
    ss->value(v);
    return ss;
  }

  Selector* Contextualize_Eval::operator()(Selector_Qualifier* s)
  {     return Contextualize::operator ()(s);
 }

  Selector* Contextualize_Eval::operator()(Type_Selector* s)
  {     return Contextualize::operator ()(s);
 }

  Selector* Contextualize_Eval::operator()(Selector_Placeholder* p)
  {
    return Contextualize::operator ()(p);
  }

  Selector* Contextualize_Eval::operator()(Selector_Reference* s)
  {
    return Contextualize::operator ()(s);
  }
}
