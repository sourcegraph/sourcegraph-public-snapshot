#ifndef SASS_AST_FWD_DECL_H
#define SASS_AST_FWD_DECL_H

/////////////////////////////////////////////
// Forward declarations for the AST visitors.
/////////////////////////////////////////////
namespace Sass {

  enum Output_Style { NESTED, EXPANDED, COMPACT, COMPRESSED, FORMATTED };

  class AST_Node;
  // statements
  class Statement;
  class Block;
  class Ruleset;
  class Propset;
  class Bubble;
  class Media_Block;
  class Feature_Block;
  class At_Rule;
  class Keyframe_Rule;
  class At_Root_Block;
  class Declaration;
  class Assignment;
  class Import;
  class Import_Stub;
  class Warning;
  class Error;
  class Debug;
  class Comment;
  class If;
  class For;
  class Each;
  class While;
  class Return;
  class Content;
  class Extension;
  class Definition;
  class Mixin_Call;
  // expressions
  class Expression;
  class List;
  class Map;
  class Binary_Expression;
  class Unary_Expression;
  class Function_Call;
  class Function_Call_Schema;
  class Variable;
  class Textual;
  class Number;
  class Color;
  class Boolean;
  class String_Schema;
  class String;
  class String_Constant;
  class String_Quoted;
  class Media_Query;
  class Media_Query_Expression;
  class Feature_Query;
  class Feature_Query_Condition;
  class At_Root_Expression;
  class Null;
  class Parent_Selector;
  // parameters and arguments
  class Parameter;
  class Parameters;
  class Argument;
  class Arguments;
  // selectors
  class Selector;
  class Selector_Schema;
  class Selector_Reference;
  class Selector_Placeholder;
  class Type_Selector;
  class Selector_Qualifier;
  class Attribute_Selector;
  class Pseudo_Selector;
  class Wrapped_Selector;
  class Compound_Selector;
  class Complex_Selector;
  class Selector_List;

}

#endif
