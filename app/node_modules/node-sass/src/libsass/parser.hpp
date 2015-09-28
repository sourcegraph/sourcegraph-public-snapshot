#ifndef SASS_PARSER_H
#define SASS_PARSER_H

#include <map>
#include <vector>
#include <iostream>

#include "ast.hpp"
#include "position.hpp"
#include "context.hpp"
#include "position.hpp"
#include "prelexer.hpp"

struct Selector_Lookahead {
  const char* found;
  bool has_interpolants;
};

namespace Sass {
  using std::string;
  using std::vector;
  using std::map;
  using namespace Prelexer;

  class Parser : public ParserState {
  private:
    void add_single_file (Import* imp, string import_path);
    void import_single_file (Import* imp, string import_path);
  public:
    class AST_Node;

    enum Syntactic_Context { nothing, mixin_def, function_def };
    bool do_import(const string& import_path, Import* imp, vector<Sass_Importer_Entry> importers, bool only_one = true);

    Context& ctx;
    vector<Block*> block_stack;
    vector<Syntactic_Context> stack;
    Media_Block* last_media_block;
    const char* source;
    const char* position;
    const char* end;
    Position before_token;
    Position after_token;
    ParserState pstate;
    int indentation;


    Token lexed;
    bool in_at_root;

    Parser(Context& ctx, const ParserState& pstate)
    : ParserState(pstate), ctx(ctx), block_stack(0), stack(0), last_media_block(0),
      source(0), position(0), end(0), before_token(pstate), after_token(pstate), pstate(pstate), indentation(0)
    { in_at_root = false; stack.push_back(nothing); }

    // static Parser from_string(const string& src, Context& ctx, ParserState pstate = ParserState("[STRING]"));
    static Parser from_c_str(const char* src, Context& ctx, ParserState pstate = ParserState("[CSTRING]"));
    static Parser from_c_str(const char* beg, const char* end, Context& ctx, ParserState pstate = ParserState("[CSTRING]"));
    static Parser from_token(Token t, Context& ctx, ParserState pstate = ParserState("[TOKEN]"));
    // special static parsers to convert strings into certain selectors
    static Selector_List* parse_selector(const char* src, Context& ctx, ParserState pstate = ParserState("[SELECTOR]"));

#ifdef __clang__

    // lex and peak uses the template parameter to branch on the action, which
    // triggers clangs tautological comparison on the single-comparison
    // branches. This is not a bug, just a merging of behaviour into
    // one function

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wtautological-compare"

#endif


    bool peek_newline(const char* start = 0);

    // skip over spaces, tabs and line comments
    template <prelexer mx>
    const char* sneak(const char* start = 0)
    {

      // maybe use optional start position from arguments?
      const char* it_position = start ? start : position;

      // skip white-space?
      if (mx == spaces ||
          mx == no_spaces ||
          mx == css_comments ||
          mx == css_whitespace ||
          mx == optional_spaces ||
          mx == optional_css_comments ||
          mx == optional_css_whitespace
      ) {
        return it_position;
      }

      // skip over spaces, tabs and sass line comments
      const char* pos = optional_css_whitespace(it_position);
      // always return a valid position
      return pos ? pos : it_position;

    }

    // peek will only skip over space, tabs and line comment
    // return the position where the lexer match will occur
    template <prelexer mx>
    const char* peek(const char* start = 0)
    {

      // sneak up to the actual token we want to lex
      // this should skip over white-space if desired
      const char* it_before_token = sneak < mx >(start);

      // match the given prelexer
      return mx(it_before_token);

    }

    // white-space handling is built into the lexer
    // this way you do not need to parse it yourself
    // some matchers don't accept certain white-space
    // we do not support start arg, since we manipulate
    // sourcemap offset and we modify the position pointer!
    // lex will only skip over space, tabs and line comment
    template <prelexer mx>
    const char* lex(bool lazy = true)
    {

      // position considered before lexed token
      // we can skip whitespace or comments for
      // lazy developers (but we need control)
      const char* it_before_token = position;

      // sneak up to the actual token we want to lex
      // this should skip over white-space if desired
      if (lazy) it_before_token = sneak < mx >(position);

      // now call matcher to get position after token
      const char* it_after_token = mx(it_before_token);

      // assertion that we got a valid match
      if (it_after_token == 0) return 0;
      // assertion that we actually lexed something
      if (it_after_token == it_before_token) return 0;

      // create new lexed token object (holds all parse result information)
      lexed = Token(position, it_before_token, it_after_token);

      // advance position (add whitespace before current token)
      before_token = after_token.add(position, it_before_token);

      // update after_token position for current token
      after_token.add(it_before_token, it_after_token);

      // ToDo: could probably do this incremetal on original object (API wants offset?)
      pstate = ParserState(path, source, lexed, before_token, after_token - before_token);

      // advance internal char iterator
      return position = it_after_token;

    }

    // lex_css skips over space, tabs, line and block comment
    // all block comments will be consumed and thrown away
    // source-map position will point to token after the comment
    template <prelexer mx>
    const char* lex_css()
    {
      // copy old token
      Token prev = lexed;
      // throw away comments
      // update srcmap position
      lex < css_comments >();
      // now lex a new token
      const char* pos = lex< mx >();
      // maybe restore prev token
      if (pos == 0) lexed = prev;
      // return match
      return pos;
    }

    // all block comments will be skipped and thrown away
    template <prelexer mx>
    const char* peek_css(const char* start = 0)
    {
      // now peek a token (skip comments first)
      return peek< mx >(peek < css_comments >(start));
    }

#ifdef __clang__

#pragma clang diagnostic pop

#endif

    void error(string msg, Position pos);
    // generate message with given and expected sample
    // text before and in the middle are configurable
    void css_error(const string& msg,
                   const string& prefix = " after ",
                   const string& middle = ", was: ");
    void read_bom();

    Block* parse();
    Import* parse_import();
    Definition* parse_definition();
    Parameters* parse_parameters();
    Parameter* parse_parameter();
    Mixin_Call* parse_mixin_call();
    Arguments* parse_arguments(bool has_url = false);
    Argument* parse_argument(bool has_url = false);
    Assignment* parse_assignment();
    // Propset* parse_propset();
    Ruleset* parse_ruleset(Selector_Lookahead lookahead);
    Selector_Schema* parse_selector_schema(const char* end_of_selector);
    Selector_List* parse_selector_group();
    Complex_Selector* parse_selector_combination();
    Compound_Selector* parse_simple_selector_sequence();
    Simple_Selector* parse_simple_selector();
    Wrapped_Selector* parse_negated_selector();
    Simple_Selector* parse_pseudo_selector();
    Attribute_Selector* parse_attribute_selector();
    Block* parse_block();
    bool parse_number_prefix();
    Declaration* parse_declaration();
    Expression* parse_map_value();
    Expression* parse_map();
    Expression* parse_list();
    Expression* parse_comma_list();
    Expression* parse_space_list();
    Expression* parse_disjunction();
    Expression* parse_conjunction();
    Expression* parse_relation();
    Expression* parse_expression();
    Expression* parse_term();
    Expression* parse_factor();
    Expression* parse_value();
    Function_Call* parse_calc_function();
    Function_Call* parse_function_call();
    Function_Call_Schema* parse_function_call_schema();
    String* parse_interpolated_chunk(Token, bool constant = false);
    String* parse_string();
    String_Constant* parse_static_expression();
    String_Constant* parse_static_value();
    String* parse_ie_property();
    String* parse_ie_keyword_arg();
    String_Schema* parse_value_schema(const char* stop);
    Expression* parse_operators(Expression* factor);
    String* parse_identifier_schema();
    // String_Schema* parse_url_schema();
    If* parse_if_directive(bool else_if = false);
    For* parse_for_directive();
    Each* parse_each_directive();
    While* parse_while_directive();
    Media_Block* parse_media_block();
    List* parse_media_queries();
    Media_Query* parse_media_query();
    Media_Query_Expression* parse_media_expression();
    Feature_Block* parse_feature_block();
    Feature_Query* parse_feature_queries();
    Feature_Query_Condition* parse_feature_query();
    Feature_Query_Condition* parse_feature_query_in_parens();
    Feature_Query_Condition* parse_supports_negation();
    Feature_Query_Condition* parse_supports_conjunction();
    Feature_Query_Condition* parse_supports_disjunction();
    Feature_Query_Condition* parse_supports_declaration();
    At_Root_Block* parse_at_root_block();
    At_Root_Expression* parse_at_root_expression();
    At_Rule* parse_at_rule();
    Warning* parse_warning();
    Error* parse_error();
    Debug* parse_debug();

    void parse_block_comments(Block* block);

    Selector_Lookahead lookahead_for_value(const char* start = 0);
    Selector_Lookahead lookahead_for_selector(const char* start = 0);
    Selector_Lookahead lookahead_for_extension_target(const char* start = 0);

    Expression* fold_operands(Expression* base, vector<Expression*>& operands, Binary_Expression::Type op);
    Expression* fold_operands(Expression* base, vector<Expression*>& operands, vector<Binary_Expression::Type>& ops);

    void throw_syntax_error(string message, size_t ln = 0);
    void throw_read_error(string message, size_t ln = 0);
  };

  size_t check_bom_chars(const char* src, const char *end, const unsigned char* bom, size_t len);
}

#endif
