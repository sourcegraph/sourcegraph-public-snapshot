#ifndef SASS_DEBUGGER_H
#define SASS_DEBUGGER_H

#include <string>
#include <sstream>
#include "ast_fwd_decl.hpp"

using namespace std;
using namespace Sass;

inline string str_replace(std::string str, const std::string& oldStr, const std::string& newStr)
{
  size_t pos = 0;
  while((pos = str.find(oldStr, pos)) != std::string::npos)
  {
     str.replace(pos, oldStr.length(), newStr);
     pos += newStr.length();
  }
  return str;
}

inline string prettyprint(const string& str) {
  string clean = str_replace(str, "\n", "\\n");
  clean = str_replace(clean, "	", "\\t");
  clean = str_replace(clean, "\r", "\\r");
  return clean;
}

inline string longToHex(long long t) {
  std::stringstream is;
  is << std::hex << t;
  return is.str();
}

inline string pstate_source_position(AST_Node* node)
{
  stringstream str;
  Position start(node->pstate());
  Position end(start + node->pstate().offset);
  str << (start.file == string::npos ? -1 : start.file)
    << "@[" << start.line << ":" << start.column << "]"
    << "-[" << end.line << ":" << end.column << "]";
  return str.str();
}

inline void debug_ast(AST_Node* node, string ind = "", Env* env = 0)
{
  if (node == 0) return;
  if (ind == "") cerr << "####################################################################\n";
  if (dynamic_cast<Bubble*>(node)) {
    Bubble* bubble = dynamic_cast<Bubble*>(node);
    cerr << ind << "Bubble " << bubble;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << bubble->tabs();
    cerr << endl;
  } else if (dynamic_cast<At_Root_Block*>(node)) {
    At_Root_Block* root_block = dynamic_cast<At_Root_Block*>(node);
    cerr << ind << "At_Root_Block " << root_block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << root_block->tabs();
    cerr << endl;
    if (root_block->block()) for(auto i : root_block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Selector_List*>(node)) {
    Selector_List* selector = dynamic_cast<Selector_List*>(node);
    cerr << ind << "Selector_List " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [block:" << selector->last_block() << "]";
    cerr << (selector->last_block() && selector->last_block()->is_root() ? " [root]" : "");
    cerr << " [@media:" << selector->media_block() << "]";
    cerr << (selector->is_optional() ? " [is_optional]": " -");
    cerr << (selector->has_line_break() ? " [line-break]": " -");
    cerr << (selector->has_line_feed() ? " [line-feed]": " -");
    cerr << endl;

    for(auto i : selector->elements()) { debug_ast(i, ind + " ", env); }

//  } else if (dynamic_cast<Expression*>(node)) {
//    Expression* expression = dynamic_cast<Expression*>(node);
//    cerr << ind << "Expression " << expression << " " << expression->concrete_type() << endl;

  } else if (dynamic_cast<Parent_Selector*>(node)) {
    Parent_Selector* selector = dynamic_cast<Parent_Selector*>(node);
    cerr << ind << "Parent_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <" << prettyprint(selector->pstate().token.ws_before()) << ">" << endl;
    debug_ast(selector->selector(), ind + "->", env);

  } else if (dynamic_cast<Complex_Selector*>(node)) {
    Complex_Selector* selector = dynamic_cast<Complex_Selector*>(node);
    cerr << ind << "Complex_Selector " << selector
      << " (" << pstate_source_position(node) << ")"
      << " [block:" << selector->last_block() << "]"
      << " [weight:" << longToHex(selector->specificity()) << "]"
      << (selector->last_block() && selector->last_block()->is_root() ? " [root]" : "")
      << " [@media:" << selector->media_block() << "]"
      << (selector->is_optional() ? " [is_optional]": " -")
      << (selector->has_line_break() ? " [line-break]": " -")
      << (selector->has_line_feed() ? " [line-feed]": " -") << " -> ";
      switch (selector->combinator()) {
        case Complex_Selector::PARENT_OF:   cerr << "{>}"; break;
        case Complex_Selector::PRECEDES:    cerr << "{~}"; break;
        case Complex_Selector::ADJACENT_TO: cerr << "{+}"; break;
        case Complex_Selector::ANCESTOR_OF: cerr << "{ }"; break;
      }
    cerr << " <" << prettyprint(selector->pstate().token.ws_before()) << ">" << endl;
    debug_ast(selector->head(), ind + " ", env);
    debug_ast(selector->tail(), ind + "-", env);
  } else if (dynamic_cast<Compound_Selector*>(node)) {
    Compound_Selector* selector = dynamic_cast<Compound_Selector*>(node);
    cerr << ind << "Compound_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [block:" << selector->last_block() << "]";
    cerr << " [weight:" << longToHex(selector->specificity()) << "]";
    // cerr << (selector->last_block() && selector->last_block()->is_root() ? " [root]" : "");
    cerr << " [@media:" << selector->media_block() << "]";
    cerr << (selector->is_optional() ? " [is_optional]": " -");
    cerr << (selector->has_line_break() ? " [line-break]": " -");
    cerr << (selector->has_line_feed() ? " [line-feed]": " -");
    cerr << " <" << prettyprint(selector->pstate().token.ws_before()) << ">" << endl;
    for(auto i : selector->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Propset*>(node)) {
    Propset* selector = dynamic_cast<Propset*>(node);
    cerr << ind << "Propset " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << selector->tabs() << endl;
    if (selector->block()) for(auto i : selector->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Wrapped_Selector*>(node)) {
    Wrapped_Selector* selector = dynamic_cast<Wrapped_Selector*>(node);
    cerr << ind << "Wrapped_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << selector->name() << ">>" << (selector->has_line_break() ? " [line-break]": " -") << (selector->has_line_feed() ? " [line-feed]": " -") << endl;
    debug_ast(selector->selector(), ind + " () ", env);
  } else if (dynamic_cast<Pseudo_Selector*>(node)) {
    Pseudo_Selector* selector = dynamic_cast<Pseudo_Selector*>(node);
    cerr << ind << "Pseudo_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << selector->name() << ">>" << (selector->has_line_break() ? " [line-break]": " -") << (selector->has_line_feed() ? " [line-feed]": " -") << endl;
    debug_ast(selector->expression(), ind + " <= ", env);
  } else if (dynamic_cast<Attribute_Selector*>(node)) {
    Attribute_Selector* selector = dynamic_cast<Attribute_Selector*>(node);
    cerr << ind << "Attribute_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << selector->name() << ">>" << (selector->has_line_break() ? " [line-break]": " -") << (selector->has_line_feed() ? " [line-feed]": " -") << endl;
    debug_ast(selector->value(), ind + "[" + selector->matcher() + "] ", env);
  } else if (dynamic_cast<Selector_Qualifier*>(node)) {
    Selector_Qualifier* selector = dynamic_cast<Selector_Qualifier*>(node);
    cerr << ind << "Selector_Qualifier " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << selector->name() << ">>" << (selector->has_line_break() ? " [line-break]": " -") << (selector->has_line_feed() ? " [line-feed]": " -") << endl;
  } else if (dynamic_cast<Type_Selector*>(node)) {
    Type_Selector* selector = dynamic_cast<Type_Selector*>(node);
    cerr << ind << "Type_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << selector->name() << ">>" << (selector->has_line_break() ? " [line-break]": " -") <<
      " <" << prettyprint(selector->pstate().token.ws_before()) << ">" << endl;
  } else if (dynamic_cast<Selector_Placeholder*>(node)) {

    Selector_Placeholder* selector = dynamic_cast<Selector_Placeholder*>(node);
    cerr << ind << "Selector_Placeholder [" << selector->name() << "] " << selector
      << " [block:" << selector->last_block() << "]"
      << " [@media:" << selector->media_block() << "]"
      << (selector->is_optional() ? " [is_optional]": " -")
      << (selector->has_line_break() ? " [line-break]": " -")
      << (selector->has_line_feed() ? " [line-feed]": " -")
    << endl;

  } else if (dynamic_cast<Selector_Reference*>(node)) {
    Selector_Reference* selector = dynamic_cast<Selector_Reference*>(node);
    cerr << ind << "Selector_Reference " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " @ref " << selector->selector() << endl;
  } else if (dynamic_cast<Simple_Selector*>(node)) {
    Simple_Selector* selector = dynamic_cast<Simple_Selector*>(node);
    cerr << ind << "Simple_Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << (selector->has_line_break() ? " [line-break]": " -") << (selector->has_line_feed() ? " [line-feed]": " -") << endl;

  } else if (dynamic_cast<Selector_Schema*>(node)) {
    Selector_Schema* selector = dynamic_cast<Selector_Schema*>(node);
    cerr << ind << "Selector_Schema " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [block:" << selector->last_block() << "]"
      << (selector->last_block() && selector->last_block()->is_root() ? " [root]" : "")
      << " [@media:" << selector->media_block() << "]"
      << (selector->has_line_break() ? " [line-break]": " -")
      << (selector->has_line_feed() ? " [line-feed]": " -")
    << endl;

    debug_ast(selector->contents(), ind + " ");
    // for(auto i : selector->elements()) { debug_ast(i, ind + " ", env); }

  } else if (dynamic_cast<Selector*>(node)) {
    Selector* selector = dynamic_cast<Selector*>(node);
    cerr << ind << "Selector " << selector;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << (selector->has_line_break() ? " [line-break]": " -")
      << (selector->has_line_feed() ? " [line-feed]": " -")
    << endl;

  } else if (dynamic_cast<Media_Query_Expression*>(node)) {
    Media_Query_Expression* block = dynamic_cast<Media_Query_Expression*>(node);
    cerr << ind << "Media_Query_Expression " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << (block->is_interpolated() ? " [is_interpolated]": " -")
    << endl;
    debug_ast(block->feature(), ind + " feature) ");
    debug_ast(block->value(), ind + " value) ");

  } else if (dynamic_cast<Media_Query*>(node)) {
    Media_Query* block = dynamic_cast<Media_Query*>(node);
    cerr << ind << "Media_Query " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << (block->is_negated() ? " [is_negated]": " -")
      << (block->is_restricted() ? " [is_restricted]": " -")
    << endl;
    debug_ast(block->media_type(), ind + " ");
    for(auto i : block->elements()) { debug_ast(i, ind + " ", env); }

  } else if (dynamic_cast<Media_Block*>(node)) {
    Media_Block* block = dynamic_cast<Media_Block*>(node);
    cerr << ind << "Media_Block " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    debug_ast(block->media_queries(), ind + " =@ ");
    debug_ast(block->selector(), ind + " -@ ");
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Feature_Block*>(node)) {
    Feature_Block* block = dynamic_cast<Feature_Block*>(node);
    cerr << ind << "Feature_Block " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Block*>(node)) {
    Block* root_block = dynamic_cast<Block*>(node);
    cerr << ind << "Block " << root_block;
    cerr << " (" << pstate_source_position(node) << ")";
    if (root_block->is_root()) cerr << " [root]";
    cerr << " " << root_block->tabs() << endl;
    if (root_block->block()) for(auto i : root_block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Warning*>(node)) {
    Warning* block = dynamic_cast<Warning*>(node);
    cerr << ind << "Warning " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Error*>(node)) {
    Error* block = dynamic_cast<Error*>(node);
    cerr << ind << "Error " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Debug*>(node)) {
    Debug* block = dynamic_cast<Debug*>(node);
    cerr << ind << "Debug " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Comment*>(node)) {
    Comment* block = dynamic_cast<Comment*>(node);
    cerr << ind << "Comment " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() <<
      " <" << prettyprint(block->pstate().token.ws_before()) << ">" << endl;
    debug_ast(block->text(), ind + "// ", env);
  } else if (dynamic_cast<If*>(node)) {
    If* block = dynamic_cast<If*>(node);
    cerr << ind << "If " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    debug_ast(block->predicate(), ind + " = ");
    debug_ast(block->consequent(), ind + " <>");
    debug_ast(block->alternative(), ind + " ><");
  } else if (dynamic_cast<Return*>(node)) {
    Return* block = dynamic_cast<Return*>(node);
    cerr << ind << "Return " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Extension*>(node)) {
    Extension* block = dynamic_cast<Extension*>(node);
    cerr << ind << "Extension " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    debug_ast(block->selector(), ind + "-> ", env);
  } else if (dynamic_cast<Content*>(node)) {
    Content* block = dynamic_cast<Content*>(node);
    cerr << ind << "Content " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Import_Stub*>(node)) {
    Import_Stub* block = dynamic_cast<Import_Stub*>(node);
    cerr << ind << "Import_Stub " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
  } else if (dynamic_cast<Import*>(node)) {
    Import* block = dynamic_cast<Import*>(node);
    cerr << ind << "Import " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    // debug_ast(block->media_queries(), ind + " @ ");
    // vector<string>         files_;
    for (auto imp : block->urls()) debug_ast(imp, "@ ", env);
  } else if (dynamic_cast<Assignment*>(node)) {
    Assignment* block = dynamic_cast<Assignment*>(node);
    cerr << ind << "Assignment " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " <<" << block->variable() << ">> " << block->tabs() << endl;
    debug_ast(block->value(), ind + "=", env);
  } else if (dynamic_cast<Declaration*>(node)) {
    Declaration* block = dynamic_cast<Declaration*>(node);
    cerr << ind << "Declaration " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    debug_ast(block->property(), ind + " prop: ", env);
    debug_ast(block->value(), ind + " value: ", env);
  } else if (dynamic_cast<Keyframe_Rule*>(node)) {
    Keyframe_Rule* has_block = dynamic_cast<Keyframe_Rule*>(node);
    cerr << ind << "Keyframe_Rule " << has_block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << has_block->tabs() << endl;
    if (has_block->selector()) debug_ast(has_block->selector(), ind + "@");
    if (has_block->block()) for(auto i : has_block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<At_Rule*>(node)) {
    At_Rule* block = dynamic_cast<At_Rule*>(node);
    cerr << ind << "At_Rule " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << block->keyword() << "] " << block->tabs() << endl;
    debug_ast(block->value(), ind + "+", env);
    debug_ast(block->selector(), ind + "~", env);
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Each*>(node)) {
    Each* block = dynamic_cast<Each*>(node);
    cerr << ind << "Each " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<For*>(node)) {
    For* block = dynamic_cast<For*>(node);
    cerr << ind << "For " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<While*>(node)) {
    While* block = dynamic_cast<While*>(node);
    cerr << ind << "While " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Definition*>(node)) {
    Definition* block = dynamic_cast<Definition*>(node);
    cerr << ind << "Definition " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [name: " << block->name() << "] ";
    cerr << " [type: " << (block->type() == Sass::Definition::Type::MIXIN ? "Mixin " : "Function ") << "] ";
    // this seems to lead to segfaults some times?
    // cerr << " [signature: " << block->signature() << "] ";
    cerr << " [native: " << block->native_function() << "] ";
    cerr << " " << block->tabs() << endl;
    debug_ast(block->parameters(), ind + " params: ", env);
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Mixin_Call*>(node)) {
    Mixin_Call* block = dynamic_cast<Mixin_Call*>(node);
    cerr << ind << "Mixin_Call " << block << " " << block->tabs();
    cerr << " [" <<  block->name() << "]" << endl;
    debug_ast(block->arguments(), ind + " args: ");
    if (block->block()) for(auto i : block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Ruleset*>(node)) {
    Ruleset* ruleset = dynamic_cast<Ruleset*>(node);
    cerr << ind << "Ruleset " << ruleset;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << ruleset->tabs() << endl;
    debug_ast(ruleset->selector(), ind + " ");
    if (ruleset->block()) for(auto i : ruleset->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Block*>(node)) {
    Block* block = dynamic_cast<Block*>(node);
    cerr << ind << "Block " << block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << block->tabs() << endl;
    for(auto i : block->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Textual*>(node)) {
    Textual* expression = dynamic_cast<Textual*>(node);
    cerr << ind << "Textual ";
    if (expression->type() == Textual::NUMBER) cerr << " [NUMBER]";
    else if (expression->type() == Textual::PERCENTAGE) cerr << " [PERCENTAGE]";
    else if (expression->type() == Textual::DIMENSION) cerr << " [DIMENSION]";
    else if (expression->type() == Textual::HEX) cerr << " [HEX]";
    cerr << expression << " [" << expression->value() << "]";
    if (expression->is_delayed()) cerr << " [delayed]";
    cerr << endl;
  } else if (dynamic_cast<Variable*>(node)) {
    Variable* expression = dynamic_cast<Variable*>(node);
    cerr << ind << "Variable " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->name() << "]" << endl;
    string name(expression->name());
    if (env && env->has(name)) debug_ast(static_cast<Expression*>((*env)[name]), ind + " -> ", env);
  } else if (dynamic_cast<Function_Call_Schema*>(node)) {
    Function_Call_Schema* expression = dynamic_cast<Function_Call_Schema*>(node);
    cerr << ind << "Function_Call_Schema " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << "" << endl;
    debug_ast(expression->name(), ind + "name: ", env);
    debug_ast(expression->arguments(), ind + " args: ", env);
  } else if (dynamic_cast<Function_Call*>(node)) {
    Function_Call* expression = dynamic_cast<Function_Call*>(node);
    cerr << ind << "Function_Call " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->name() << "]" << endl;
    debug_ast(expression->arguments(), ind + " args: ", env);
  } else if (dynamic_cast<Arguments*>(node)) {
    Arguments* expression = dynamic_cast<Arguments*>(node);
    cerr << ind << "Arguments " << expression;
    if (expression->is_delayed()) cerr << " [delayed]";
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << endl;
    for(auto i : expression->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Argument*>(node)) {
    Argument* expression = dynamic_cast<Argument*>(node);
    cerr << ind << "Argument " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->value() << "]";
    cerr << " [name: " << expression->name() << "] ";
    cerr << " [rest: " << expression->is_rest_argument() << "] ";
    cerr << " [keyword: " << expression->is_keyword_argument() << "] " << endl;
    debug_ast(expression->value(), ind + " value: ", env);
  } else if (dynamic_cast<Parameters*>(node)) {
    Parameters* expression = dynamic_cast<Parameters*>(node);
    cerr << ind << "Parameters " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [has_optional: " << expression->has_optional_parameters() << "] ";
    cerr << " [has_rest: " << expression->has_rest_parameter() << "] ";
    cerr << endl;
    for(auto i : expression->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Parameter*>(node)) {
    Parameter* expression = dynamic_cast<Parameter*>(node);
    cerr << ind << "Parameter " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [name: " << expression->name() << "] ";
    cerr << " [default: " << expression->default_value() << "] ";
    cerr << " [rest: " << expression->is_rest_parameter() << "] " << endl;
  } else if (dynamic_cast<Unary_Expression*>(node)) {
    Unary_Expression* expression = dynamic_cast<Unary_Expression*>(node);
    cerr << ind << "Unary_Expression " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->type() << "]" << endl;
    debug_ast(expression->operand(), ind + " operand: ", env);
  } else if (dynamic_cast<Binary_Expression*>(node)) {
    Binary_Expression* expression = dynamic_cast<Binary_Expression*>(node);
    cerr << ind << "Binary_Expression " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->type_name() << "]" << endl;
    debug_ast(expression->left(), ind + " left:  ", env);
    debug_ast(expression->right(), ind + " right: ", env);
  } else if (dynamic_cast<Map*>(node)) {
    Map* expression = dynamic_cast<Map*>(node);
    cerr << ind << "Map " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [Hashed]" << endl;
    for (auto i : expression->elements()) {
      debug_ast(i.first, ind + " key: ");
      debug_ast(i.second, ind + " val: ");
    }
  } else if (dynamic_cast<List*>(node)) {
    List* expression = dynamic_cast<List*>(node);
    cerr << ind << "List " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " (" << expression->length() << ") " <<
      (expression->separator() == Sass::List::Separator::COMMA ? "Comma " : "Space ") <<
      " [delayed: " << expression->is_delayed() << "] " <<
      " [interpolant: " << expression->is_interpolant() << "] " <<
      " [arglist: " << expression->is_arglist() << "] " <<
      endl;
    for(auto i : expression->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Content*>(node)) {
    Content* expression = dynamic_cast<Content*>(node);
    cerr << ind << "Content " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [Statement]" << endl;
  } else if (dynamic_cast<Boolean*>(node)) {
    Boolean* expression = dynamic_cast<Boolean*>(node);
    cerr << ind << "Boolean " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->value() << "]" << endl;
  } else if (dynamic_cast<Color*>(node)) {
    Color* expression = dynamic_cast<Color*>(node);
    cerr << ind << "Color " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->r() << ":"  << expression->g() << ":" << expression->b() << "@" << expression->a() << "]" << endl;
  } else if (dynamic_cast<Number*>(node)) {
    Number* expression = dynamic_cast<Number*>(node);
    cerr << ind << "Number " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << expression->value() << expression->unit() << "]" << endl;
  } else if (dynamic_cast<String_Quoted*>(node)) {
    String_Quoted* expression = dynamic_cast<String_Quoted*>(node);
    cerr << ind << "String_Quoted " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << prettyprint(expression->value()) << "]";
    if (expression->is_delayed()) cerr << " [delayed]";
    if (expression->sass_fix_1291()) cerr << " [sass_fix_1291]";
    if (expression->quote_mark()) cerr << " [quote_mark: " << expression->quote_mark() << "]";
    cerr << " <" << prettyprint(expression->pstate().token.ws_before()) << ">" << endl;
  } else if (dynamic_cast<String_Constant*>(node)) {
    String_Constant* expression = dynamic_cast<String_Constant*>(node);
    cerr << ind << "String_Constant " << expression;
    if (expression->concrete_type()) {
      cerr << " " << expression->concrete_type();
    }
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " [" << prettyprint(expression->value()) << "]";
    if (expression->is_delayed()) cerr << " [delayed]";
    if (expression->sass_fix_1291()) cerr << " [sass_fix_1291]";
    cerr << " <" << prettyprint(expression->pstate().token.ws_before()) << ">" << endl;
  } else if (dynamic_cast<String_Schema*>(node)) {
    String_Schema* expression = dynamic_cast<String_Schema*>(node);
    cerr << ind << "String_Schema " << expression;
    cerr << " " << expression->concrete_type();
    if (expression->is_delayed()) cerr << " [delayed]";
    if (expression->has_interpolants()) cerr << " [has_interpolants]";
    cerr << " <" << prettyprint(expression->pstate().token.ws_before()) << ">" << endl;
    for(auto i : expression->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<String*>(node)) {
    String* expression = dynamic_cast<String*>(node);
    cerr << ind << "String " << expression;
    cerr << " " << expression->concrete_type();
    cerr << " (" << pstate_source_position(node) << ")";
    if (expression->sass_fix_1291()) cerr << " [sass_fix_1291]";
    cerr << " <" << prettyprint(expression->pstate().token.ws_before()) << ">" << endl;
  } else if (dynamic_cast<Expression*>(node)) {
    Expression* expression = dynamic_cast<Expression*>(node);
    cerr << ind << "Expression " << expression;
    cerr << " (" << pstate_source_position(node) << ")";
    switch (expression->concrete_type()) {
      case Expression::Concrete_Type::NONE: cerr << " [NONE]"; break;
      case Expression::Concrete_Type::BOOLEAN: cerr << " [BOOLEAN]"; break;
      case Expression::Concrete_Type::NUMBER: cerr << " [NUMBER]"; break;
      case Expression::Concrete_Type::COLOR: cerr << " [COLOR]"; break;
      case Expression::Concrete_Type::STRING: cerr << " [STRING]"; break;
      case Expression::Concrete_Type::LIST: cerr << " [LIST]"; break;
      case Expression::Concrete_Type::MAP: cerr << " [MAP]"; break;
      case Expression::Concrete_Type::SELECTOR: cerr << " [SELECTOR]"; break;
      case Expression::Concrete_Type::NULL_VAL: cerr << " [NULL_VAL]"; break;
      case Expression::Concrete_Type::NUM_TYPES: cerr << " [NUM_TYPES]"; break;
    }
    cerr << endl;
  } else if (dynamic_cast<Has_Block*>(node)) {
    Has_Block* has_block = dynamic_cast<Has_Block*>(node);
    cerr << ind << "Has_Block " << has_block;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << has_block->tabs() << endl;
    if (has_block->block()) for(auto i : has_block->block()->elements()) { debug_ast(i, ind + " ", env); }
  } else if (dynamic_cast<Statement*>(node)) {
    Statement* statement = dynamic_cast<Statement*>(node);
    cerr << ind << "Statement " << statement;
    cerr << " (" << pstate_source_position(node) << ")";
    cerr << " " << statement->tabs() << endl;
  }

  if (ind == "") cerr << "####################################################################\n";
}

#endif // SASS_DEBUGGER
