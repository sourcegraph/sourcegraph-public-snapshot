#include <cstdlib>
#include <iostream>
#include <vector>
#include "parser.hpp"
#include "file.hpp"
#include "inspect.hpp"
#include "to_string.hpp"
#include "constants.hpp"
#include "util.hpp"
#include "prelexer.hpp"
#include "sass_functions.h"

#include <typeinfo>

namespace Sass {
  using namespace std;
  using namespace Constants;

  Parser Parser::from_c_str(const char* str, Context& ctx, ParserState pstate)
  {
    Parser p(ctx, pstate);
    p.source   = str;
    p.position = p.source;
    p.end      = str + strlen(str);
    Block* root = new (ctx.mem) Block(pstate);
    p.block_stack.push_back(root);
    root->is_root(true);
    return p;
  }

  Parser Parser::from_c_str(const char* beg, const char* end, Context& ctx, ParserState pstate)
  {
    Parser p(ctx, pstate);
    p.source   = beg;
    p.position = p.source;
    p.end      = end;
    Block* root = new (ctx.mem) Block(pstate);
    p.block_stack.push_back(root);
    root->is_root(true);
    return p;
  }

  Selector_List* Parser::parse_selector(const char* src, Context& ctx, ParserState pstate)
  {
    Parser p = Parser::from_c_str(src, ctx, pstate);
    // ToDo: ruby sass errors on parent references
    // ToDo: remap the source-map entries somehow
    return p.parse_selector_group();
  }

  bool Parser::peek_newline(const char* start)
  {
    return peek_linefeed(start ? start : position);
  }

  Parser Parser::from_token(Token t, Context& ctx, ParserState pstate)
  {
    Parser p(ctx, pstate);
    p.source   = t.begin;
    p.position = p.source;
    p.end      = t.end;
    Block* root = new (ctx.mem) Block(pstate);
    p.block_stack.push_back(root);
    root->is_root(true);
    return p;
  }

  Block* Parser::parse()
  {
    Block* root = new (ctx.mem) Block(pstate);
    block_stack.push_back(root);
    root->is_root(true);
    read_bom();

    if (ctx.queue.size() == 1) {
      Import* pre = new (ctx.mem) Import(pstate);
      string load_path(ctx.queue[0].load_path);
      do_import(load_path, pre, ctx.c_headers, false);
      ctx.head_imports = ctx.queue.size() - 1;
      if (!pre->urls().empty()) (*root) << pre;
      if (!pre->files().empty()) {
        for (size_t i = 0, S = pre->files().size(); i < S; ++i) {
          (*root) << new (ctx.mem) Import_Stub(pstate, pre->files()[i]);
        }
      }
    }

    bool semicolon = false;
    string(error_message);
    lex< optional_spaces >();
    Selector_Lookahead lookahead_result;
    while (position < end) {
      parse_block_comments(root);
      if (peek< kwd_import >()) {
        Import* imp = parse_import();
        if (!imp->urls().empty()) (*root) << imp;
        if (!imp->files().empty()) {
          for (size_t i = 0, S = imp->files().size(); i < S; ++i) {
            (*root) << new (ctx.mem) Import_Stub(pstate, imp->files()[i]);
          }
        }
        semicolon = true;
        error_message = "top-level @import directive must be terminated by ';'";
      }
      else if (peek< kwd_mixin >() || peek< kwd_function >()) {
        (*root) << parse_definition();
      }
      else if (peek< variable >()) {
        (*root) << parse_assignment();
        semicolon = true;
        error_message = "top-level variable binding must be terminated by ';'";
      }
      /*else if (peek< sequence< optional< exactly<'*'> >, alternatives< identifier_schema, identifier >, optional_spaces, exactly<':'>, optional_spaces, exactly<'{'> > >(position)) {
        (*root) << parse_propset();
      }*/
      else if (peek< kwd_include >() /* || peek< exactly<'+'> >() */) {
        Mixin_Call* mixin_call = parse_mixin_call();
        (*root) << mixin_call;
        if (!mixin_call->block()) {
          semicolon = true;
          error_message = "top-level @include directive must be terminated by ';'";
        }
      }
      else if (peek< kwd_if_directive >()) {
        (*root) << parse_if_directive();
      }
      else if (peek< kwd_for_directive >()) {
        (*root) << parse_for_directive();
      }
      else if (peek< kwd_each_directive >()) {
        (*root) << parse_each_directive();
      }
      else if (peek< kwd_while_directive >()) {
        (*root) << parse_while_directive();
      }
      else if (peek< kwd_media >()) {
        (*root) << parse_media_block();
      }
      else if (peek< kwd_at_root >()) {
        (*root) << parse_at_root_block();
      }
      else if (peek< kwd_supports >()) {
        (*root) << parse_feature_block();
      }
      else if (peek< kwd_warn >()) {
        (*root) << parse_warning();
        semicolon = true;
        error_message = "top-level @warn directive must be terminated by ';'";
      }
      else if (peek< kwd_err >()) {
        (*root) << parse_error();
        semicolon = true;
        error_message = "top-level @error directive must be terminated by ';'";
      }
      else if (peek< kwd_dbg >()) {
        (*root) << parse_debug();
        semicolon = true;
        error_message = "top-level @debug directive must be terminated by ';'";
      }
      // ignore the @charset directive for now
      else if (lex< exactly< charset_kwd > >()) {
        lex< quoted_string >();
        lex< one_plus< exactly<';'> > >();
      }
      else if (peek< at_keyword >()) {
        At_Rule* at_rule = parse_at_rule();
        (*root) << at_rule;
        if (!at_rule->block()){
          semicolon = true;
          error_message = "top-level directive must be terminated by ';'";
        }
      }
      else if ((lookahead_result = lookahead_for_selector(position)).found) {
        (*root) << parse_ruleset(lookahead_result);
      }
      else if (peek< exactly<';'> >()) {
        lex< one_plus< exactly<';'> > >();
      }
      else {
        lex< css_whitespace >();
        if (position >= end) break;
        error("invalid top-level expression", after_token);
      }
      if (semicolon) {
        if (!lex< one_plus< exactly<';'> > >() && peek_css< optional_css_whitespace >() != end)
        { error(error_message, pstate); }
        semicolon = false;
      }
      lex< optional_spaces >();
    }
    block_stack.pop_back();
    return root;
  }

  void Parser::add_single_file (Import* imp, string import_path) {

    string extension;
    string unquoted(unquote(import_path));
    if (unquoted.length() > 4) { // 2 quote marks + the 4 chars in .css
      // a string constant is guaranteed to end with a quote mark, so make sure to skip it when indexing from the end
      extension = unquoted.substr(unquoted.length() - 4, 4);
    }

    if (extension == ".css") {
      String_Constant* loc = new (ctx.mem) String_Constant(pstate, unquote(import_path));
      Argument* loc_arg = new (ctx.mem) Argument(pstate, loc);
      Arguments* loc_args = new (ctx.mem) Arguments(pstate);
      (*loc_args) << loc_arg;
      Function_Call* new_url = new (ctx.mem) Function_Call(pstate, "url", loc_args);
      imp->urls().push_back(new_url);
    }
    else {
      string current_dir = File::dir_name(path);
      string resolved(ctx.add_file(current_dir, unquoted));
      if (resolved.empty()) error("file to import not found or unreadable: " + unquoted + "\nCurrent dir: " + current_dir, pstate);
      imp->files().push_back(resolved);
    }

  }

  void Parser::import_single_file (Import* imp, string import_path) {

    if (!unquote(import_path).substr(0, 7).compare("http://") ||
        !unquote(import_path).substr(0, 8).compare("https://") ||
        !unquote(import_path).substr(0, 2).compare("//"))
    {
      imp->urls().push_back(new (ctx.mem) String_Quoted(pstate, import_path));
    }
    else {
      add_single_file(imp, import_path);
    }

  }

  bool Parser::do_import(const string& import_path, Import* imp, vector<Sass_Importer_Entry> importers, bool only_one)
  {
    bool has_import = false;
    string load_path = unquote(import_path);
    for (auto importer : importers) {
      // int priority = sass_importer_get_priority(importer);
      Sass_Importer_Fn fn = sass_importer_get_function(importer);
      if (Sass_Import_List includes =
          fn(load_path.c_str(), importer, ctx.c_compiler)
      ) {
        Sass_Import_List list = includes;
        while (*includes) {
          Sass_Import_Entry include = *includes;
          const char *file = sass_import_get_path(include);
          char* source = sass_import_take_source(include);
          size_t line = sass_import_get_error_line(include);
          size_t column = sass_import_get_error_column(include);
          const char* message = sass_import_get_error_message(include);
          if (message) {
            if (line == string::npos && column == string::npos) error(message, pstate);
            else error(message, ParserState(message, source, Position(line, column)));
          } else if (source) {
            if (file) {
              ctx.add_source(file, load_path, source);
              imp->files().push_back(file);
            } else {
              ctx.add_source(load_path, load_path, source);
              imp->files().push_back(load_path);
            }
          } else if(file) {
            import_single_file(imp, file);
          }
          ++includes;
        }
        // deallocate returned memory
        sass_delete_import_list(list);
        // set success flag
        has_import = true;
        // break import chain
        if (only_one) return true;
      }
    }
    // return result
    return has_import;
  }

  Import* Parser::parse_import()
  {
    lex< kwd_import >();
    Import* imp = new (ctx.mem) Import(pstate);
    bool first = true;
    do {
      while (lex< block_comment >());
      if (lex< quoted_string >()) {
        if (!do_import(lexed, imp, ctx.c_importers, true))
        {
          // push single file import
          import_single_file(imp, lexed);
        }
      }
      else if (lex< uri_prefix >()) {
        Arguments* args = new (ctx.mem) Arguments(pstate);
        Function_Call* result = new (ctx.mem) Function_Call(pstate, "url", args);
        if (lex< quoted_string >()) {
          Expression* the_url = parse_string();
          *args << new (ctx.mem) Argument(the_url->pstate(), the_url);
        }
        else if (lex < uri_value >(position)) { // chunk seems to work too!
          String* the_url = parse_interpolated_chunk(lexed);
          *args << new (ctx.mem) Argument(the_url->pstate(), the_url);
        }
        else if (peek < skip_over_scopes < exactly < '(' >, exactly < ')' > > >(position)) {
          Expression* the_url = parse_list(); // parse_interpolated_chunk(lexed);
          *args << new (ctx.mem) Argument(the_url->pstate(), the_url);
        }
        else {
          error("malformed URL", pstate);
        }
        if (!lex< exactly<')'> >()) error("URI is missing ')'", pstate);
        imp->urls().push_back(result);
      }
      else {
        if (first) error("@import directive requires a url or quoted path", pstate);
        else error("expecting another url or quoted path in @import list", pstate);
      }
      first = false;
    } while (lex_css< exactly<','> >());
    return imp;
  }

  Definition* Parser::parse_definition()
  {
    Definition::Type which_type = Definition::MIXIN;
    if      (lex< kwd_mixin >())    which_type = Definition::MIXIN;
    else if (lex< kwd_function >()) which_type = Definition::FUNCTION;
    string which_str(lexed);
    if (!lex< identifier >()) error("invalid name in " + which_str + " definition", pstate);
    string name(Util::normalize_underscores(lexed));
    if (which_type == Definition::FUNCTION && (name == "and" || name == "or" || name == "not"))
    { error("Invalid function name \"" + name + "\".", pstate); }
    ParserState source_position_of_def = pstate;
    Parameters* params = parse_parameters();
    if (!peek< exactly<'{'> >()) error("body for " + which_str + " " + name + " must begin with a '{'", pstate);
    if (which_type == Definition::MIXIN) stack.push_back(mixin_def);
    else stack.push_back(function_def);
    Block* body = parse_block();
    stack.pop_back();
    Definition* def = new (ctx.mem) Definition(source_position_of_def, name, params, body, &ctx, which_type);
    return def;
  }

  Parameters* Parser::parse_parameters()
  {
    string name(lexed);
    Position position = after_token;
    Parameters* params = new (ctx.mem) Parameters(pstate);
    if (lex_css< exactly<'('> >()) {
      // if there's anything there at all
      if (!peek_css< exactly<')'> >()) {
        do (*params) << parse_parameter();
        while (lex_css< exactly<','> >());
      }
      if (!lex_css< exactly<')'> >()) error("expected a variable name (e.g. $x) or ')' for the parameter list for " + name, position);
    }
    return params;
  }

  Parameter* Parser::parse_parameter()
  {
    while (lex< alternatives < spaces, block_comment > >());
    lex< variable >();
    string name(Util::normalize_underscores(lexed));
    ParserState pos = pstate;
    Expression* val = 0;
    bool is_rest = false;
    while (lex< alternatives < spaces, block_comment > >());
    if (lex< exactly<':'> >()) { // there's a default value
      while (lex< block_comment >());
      val = parse_space_list();
      val->is_delayed(false);
    }
    else if (lex< exactly< ellipsis > >()) {
      is_rest = true;
    }
    Parameter* p = new (ctx.mem) Parameter(pos, name, val, is_rest);
    return p;
  }

  Mixin_Call* Parser::parse_mixin_call()
  {
    lex< kwd_include >() /* || lex< exactly<'+'> >() */;
    if (!lex< identifier >()) error("invalid name in @include directive", pstate);
    ParserState source_position_of_call = pstate;
    string name(Util::normalize_underscores(lexed));
    Arguments* args = parse_arguments();
    Block* content = 0;
    if (peek< exactly<'{'> >()) {
      content = parse_block();
    }
    Mixin_Call* the_call = new (ctx.mem) Mixin_Call(source_position_of_call, name, args, content);
    return the_call;
  }

  Arguments* Parser::parse_arguments(bool has_url)
  {
    string name(lexed);
    Position position = after_token;
    Arguments* args = new (ctx.mem) Arguments(pstate);
    if (lex_css< exactly<'('> >()) {
      // if there's anything there at all
      if (!peek_css< exactly<')'> >()) {
        do (*args) << parse_argument(has_url);
        while (lex_css< exactly<','> >());
      }
      if (!lex_css< exactly<')'> >()) error("expected a variable name (e.g. $x) or ')' for the parameter list for " + name, position);
    }
    return args;
  }

  Argument* Parser::parse_argument(bool has_url)
  {

    Argument* arg;
    // some urls can look like line comments (parse literally - chunk would not work)
    if (has_url && lex< sequence < uri_value, lookahead < loosely<')'> > > >(false)) {
      String* the_url = parse_interpolated_chunk(lexed);
      arg = new (ctx.mem) Argument(the_url->pstate(), the_url);
    }
    else if (peek_css< sequence < variable, optional_css_comments, exactly<':'> > >()) {
      lex_css< variable >();
      string name(Util::normalize_underscores(lexed));
      ParserState p = pstate;
      lex_css< exactly<':'> >();
      Expression* val = parse_space_list();
      val->is_delayed(false);
      arg = new (ctx.mem) Argument(p, val, name);
    }
    else {
      bool is_arglist = false;
      bool is_keyword = false;
      Expression* val = parse_space_list();
      val->is_delayed(false);
      if (lex_css< exactly< ellipsis > >()) {
        if (val->concrete_type() == Expression::MAP) is_keyword = true;
        else is_arglist = true;
      }
      arg = new (ctx.mem) Argument(pstate, val, "", is_arglist, is_keyword);
    }
    return arg;
  }

  Assignment* Parser::parse_assignment()
  {
    lex< variable >();
    string name(Util::normalize_underscores(lexed));
    ParserState var_source_position = pstate;
    if (!lex< exactly<':'> >()) error("expected ':' after " + name + " in assignment statement", pstate);
    Expression* val;
    Selector_Lookahead lookahead = lookahead_for_value(position);
    if (lookahead.has_interpolants && lookahead.found) {
      val = parse_value_schema(lookahead.found);
    } else {
      val = parse_list();
    }
    val->is_delayed(false);
    bool is_default = false;
    bool is_global = false;
    while (peek< default_flag >() || peek< global_flag >()) {
      is_default = lex< default_flag >() || is_default;
      is_global = lex< global_flag >() || is_global;
    }
    Assignment* var = new (ctx.mem) Assignment(var_source_position, name, val, is_default, is_global);
    return var;
  }

  /* not used anymore - remove?
  Propset* Parser::parse_propset()
  {
    String* property_segment;
    if (peek< sequence< optional< exactly<'*'> >, identifier_schema > >()) {
      property_segment = parse_identifier_schema();
    }
    else {
      lex< sequence< optional< exactly<'*'> >, identifier > >();
      property_segment = new (ctx.mem) String_Quoted(pstate, lexed);
    }
    Propset* propset = new (ctx.mem) Propset(pstate, property_segment);
    lex< exactly<':'> >();

    if (!peek< exactly<'{'> >()) error("expected a '{' after namespaced property", pstate);

    propset->block(parse_block());

    propset->tabs(indentation);

    return propset;
  } */

  Ruleset* Parser::parse_ruleset(Selector_Lookahead lookahead)
  {
    Selector* sel;
    if (lookahead.has_interpolants) {
      sel = parse_selector_schema(lookahead.found);
    }
    else {
      sel = parse_selector_group();
    }
    bool old_in_at_root = in_at_root;
    ParserState r_source_position = pstate;
    lex < css_comments >();
    in_at_root = false;
    if (!peek< exactly<'{'> >()) error("expected a '{' after the selector", pstate);
    Block* block = parse_block();
    in_at_root = old_in_at_root;
    old_in_at_root = false;
    Ruleset* ruleset = new (ctx.mem) Ruleset(r_source_position, sel, block);
    return ruleset;
  }

  Selector_Schema* Parser::parse_selector_schema(const char* end_of_selector)
  {
    lex< optional_spaces >();
    const char* i = position;
    String_Schema* schema = new (ctx.mem) String_Schema(pstate);
    while (i < end_of_selector) {
      // try to parse mutliple interpolants
      if (const char* p = find_first_in_interval< exactly<hash_lbrace> >(i, end_of_selector)) {
        // accumulate the preceding segment if the position has advanced
        if (i < p) (*schema) << new (ctx.mem) String_Quoted(pstate, string(i, p));
        // skip to the delimiter by skipping occurences in quoted strings
        if (peek < sequence < optional_spaces, exactly<rbrace> > >(p+2)) { position = p+2;
          css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
        }
        const char* j = skip_over_scopes< exactly<hash_lbrace>, exactly<rbrace> >(p + 2, end_of_selector);
        Expression* interpolant = Parser::from_c_str(p+2, j, ctx, pstate).parse_list();
        interpolant->is_interpolant(true);
        (*schema) << interpolant;
        i = j;
      }
      // no more interpolants have been found
      // add the last segment if there is one
      else {
        if (i < end_of_selector) (*schema) << new (ctx.mem) String_Quoted(pstate, string(i, end_of_selector));
        break;
      }
    }
    position = end_of_selector;
    Selector_Schema* selector_schema = new (ctx.mem) Selector_Schema(pstate, schema);
    selector_schema->media_block(last_media_block);
    selector_schema->last_block(block_stack.back());
    return selector_schema;
  }

  Selector_List* Parser::parse_selector_group()
  {
    bool reloop = true;
    To_String to_string(&ctx);
    lex< css_whitespace >();
    Selector_List* group = new (ctx.mem) Selector_List(pstate);
    group->media_block(last_media_block);
    group->last_block(block_stack.back());
    do {
      reloop = false;
      if (peek< alternatives <
            exactly<'{'>,
            exactly<'}'>,
            exactly<')'>,
            exactly<';'>
          > >())
        break; // in case there are superfluous commas at the end
      Complex_Selector* comb = parse_selector_combination();
      if (!comb->has_reference() && !in_at_root) {
        ParserState sel_source_position = pstate;
        Selector_Reference* ref = new (ctx.mem) Selector_Reference(sel_source_position);
        Compound_Selector* ref_wrap = new (ctx.mem) Compound_Selector(sel_source_position);
        ref_wrap->media_block(last_media_block);
        ref_wrap->last_block(block_stack.back());
        (*ref_wrap) << ref;
        if (!comb->head()) {
          comb->head(ref_wrap);
          comb->has_reference(true);
        }
        else {
          comb = new (ctx.mem) Complex_Selector(sel_source_position, Complex_Selector::ANCESTOR_OF, ref_wrap, comb);
          comb->media_block(last_media_block);
          comb->last_block(block_stack.back());
          comb->has_reference(true);
        }
        if (peek_newline()) ref_wrap->has_line_break(true);
      }
      while (peek_css< exactly<','> >())
      {
        // consume everything up and including the comma speparator
        reloop = lex< sequence < optional_css_comments, exactly<','> > >() != 0;
        // remember line break (also between some commas)
        if (peek_newline()) comb->has_line_feed(true);
        if (comb->tail() && peek_newline()) comb->tail()->has_line_feed(true);
        if (comb->tail() && comb->tail()->head() && peek_newline()) comb->tail()->head()->has_line_feed(true);
        // remember line break (also between some commas)
      }
      (*group) << comb;
    }
    while (reloop);
    while (lex< optional >()) {
      group->is_optional(true);
    }
    return group;
  }

  Complex_Selector* Parser::parse_selector_combination()
  {
    Position sel_source_position(-1);
    Compound_Selector* lhs;
    if (peek_css< alternatives <
          exactly<'+'>,
          exactly<'~'>,
          exactly<'>'>
        > >())
    // no selector before the combinator
    { lhs = 0; }
    else {
      lhs = parse_simple_selector_sequence();
      sel_source_position = before_token;
      lhs->has_line_break(peek_newline());
    }

    Complex_Selector::Combinator cmb;
    if      (lex< exactly<'+'> >()) cmb = Complex_Selector::ADJACENT_TO;
    else if (lex< exactly<'~'> >()) cmb = Complex_Selector::PRECEDES;
    else if (lex< exactly<'>'> >()) cmb = Complex_Selector::PARENT_OF;
    else                            cmb = Complex_Selector::ANCESTOR_OF;
    bool cpx_lf = peek_newline();

    Complex_Selector* rhs;
    if (peek_css< alternatives <
                exactly<','>,
                exactly<')'>,
                exactly<'{'>,
                exactly<'}'>,
                exactly<';'>,
                optional
        > >())
    // no selector after the combinator
    { rhs = 0; }
    else {
      rhs = parse_selector_combination();
      sel_source_position = before_token;
    }
    if (!sel_source_position.line) sel_source_position = before_token;
    Complex_Selector* cpx = new (ctx.mem) Complex_Selector(ParserState(path, source, sel_source_position), cmb, lhs, rhs);
    cpx->media_block(last_media_block);
    cpx->last_block(block_stack.back());
    if (cpx_lf) cpx->has_line_break(cpx_lf);
    return cpx;
  }

  Compound_Selector* Parser::parse_simple_selector_sequence()
  {
    Compound_Selector* seq = new (ctx.mem) Compound_Selector(pstate);
    seq->media_block(last_media_block);
    seq->last_block(block_stack.back());
    bool sawsomething = false;
    if (lex_css< exactly<'&'> >()) {
      // check if we have a parent selector on the root level block
      if (block_stack.back() && block_stack.back()->is_root()) {
        //error("Base-level rules cannot contain the parent-selector-referencing character '&'.", pstate);
      }
      (*seq) << new (ctx.mem) Selector_Reference(pstate);
      sawsomething = true;
      // if you see a space after a &, then you're done
      if(peek< spaces >() || peek< alternatives < spaces, exactly<';'> > >()) {
        return seq;
      }
    }
    if (sawsomething && lex_css< sequence< negate< functional >, alternatives< identifier_alnums, universal, quoted_string, dimension, percentage, number > > >()) {
      // saw an ampersand, then allow type selectors with arbitrary number of hyphens at the beginning
      (*seq) << new (ctx.mem) Type_Selector(pstate, unquote(lexed));
    } else if (lex_css< sequence< negate< functional >, alternatives< type_selector, universal, quoted_string, dimension, percentage, number > > >()) {
      // if you see a type selector
      (*seq) << new (ctx.mem) Type_Selector(pstate, lexed);
      sawsomething = true;
    }
    if (!sawsomething) {
      // don't blindly do this if you saw a & or selector
      (*seq) << parse_simple_selector();
    }

    while (!peek< spaces >(position) &&
           !(peek_css < alternatives <
               exactly<'+'>,
               exactly<'~'>,
               exactly<'>'>,
               exactly<','>,
               exactly<')'>,
               exactly<'{'>,
               exactly<'}'>,
               exactly<';'>
             > >(position))) {
      (*seq) << parse_simple_selector();
    }
    return seq;
  }

  Simple_Selector* Parser::parse_simple_selector()
  {
    lex < css_comments >();
    if (lex< alternatives < id_name, class_name > >()) {
      return new (ctx.mem) Selector_Qualifier(pstate, unquote(lexed));
    }
    else if (lex< quoted_string >()) {
      return new (ctx.mem) Type_Selector(pstate, unquote(lexed));
    }
    else if (lex< alternatives < number, kwd_sel_deep > >()) {
      return new (ctx.mem) Type_Selector(pstate, lexed);
    }
    else if (peek< pseudo_not >()) {
      return parse_negated_selector();
    }
    else if (peek< exactly<':'> >(position) || peek< functional >()) {
      return parse_pseudo_selector();
    }
    else if (peek< exactly<'['> >(position)) {
      return parse_attribute_selector();
    }
    else if (lex< placeholder >()) {
      Selector_Placeholder* sel = new (ctx.mem) Selector_Placeholder(pstate, unquote(lexed));
      sel->media_block(last_media_block);
      sel->last_block(block_stack.back());
      return sel;
    }
    else {
      error("invalid selector after " + lexed.to_string(), pstate);
    }
    // unreachable statement
    return 0;
  }

  Wrapped_Selector* Parser::parse_negated_selector()
  {
    lex< pseudo_not >();
    string name(lexed);
    ParserState nsource_position = pstate;
    Selector* negated = parse_selector_group();
    if (!lex< exactly<')'> >()) {
      error("negated selector is missing ')'", pstate);
    }
    return new (ctx.mem) Wrapped_Selector(nsource_position, name, negated);
  }

  Simple_Selector* Parser::parse_pseudo_selector() {
    if (lex< sequence< pseudo_prefix, functional > >() || lex< functional >()) {
      string name(lexed);
      String* expr = 0;
      ParserState p = pstate;
      Selector* wrapped = 0;
      if (lex< alternatives< even, odd > >()) {
        expr = new (ctx.mem) String_Quoted(p, lexed);
      }
      else if (lex< binomial >(position)) {
        expr = new (ctx.mem) String_Constant(p, lexed);
        ((String_Constant*)expr)->can_compress_whitespace(true);
      }
      else if (peek< sequence< optional<sign>,
                               zero_plus<digit>,
                               exactly<'n'>,
                               optional_css_whitespace,
                               exactly<')'> > >()) {
        lex< sequence< optional<sign>,
                       zero_plus<digit>,
                       exactly<'n'> > >();
        expr = new (ctx.mem) String_Quoted(p, lexed);
      }
      else if (lex< sequence< optional<sign>, one_plus < digit > > >()) {
        expr = new (ctx.mem) String_Quoted(p, lexed);
      }
      else if (peek< sequence< identifier, optional_css_whitespace, exactly<')'> > >()) {
        lex< identifier >();
        expr = new (ctx.mem) String_Quoted(p, lexed);
      }
      else if (lex< quoted_string >()) {
        expr = new (ctx.mem) String_Quoted(p, lexed);
      }
      else if (peek< exactly<')'> >()) {
        expr = new (ctx.mem) String_Constant(p, "");
      }
      else {
        wrapped = parse_selector_group();
      }
      if (!lex< exactly<')'> >()) error("unterminated argument to " + name + "...)", pstate);
      if (wrapped) {
        return new (ctx.mem) Wrapped_Selector(p, name, wrapped);
      }
      return new (ctx.mem) Pseudo_Selector(p, name, expr);
    }
    else if (lex < sequence< pseudo_prefix, identifier > >()) {
      return new (ctx.mem) Pseudo_Selector(pstate, unquote(lexed));
    }
    else {
      error("unrecognized pseudo-class or pseudo-element", pstate);
    }
    // unreachable statement
    return 0;
  }

  Attribute_Selector* Parser::parse_attribute_selector()
  {
    lex_css< exactly<'['> >();
    ParserState p = pstate;
    if (!lex_css< attribute_name >()) error("invalid attribute name in attribute selector", pstate);
    string name(lexed);
    if (lex_css< exactly<']'> >()) return new (ctx.mem) Attribute_Selector(p, name, "", 0);
    if (!lex_css< alternatives< exact_match, class_match, dash_match,
                                prefix_match, suffix_match, substring_match > >()) {
      error("invalid operator in attribute selector for " + name, pstate);
    }
    string matcher(lexed);

    String* value = 0;
    if (lex_css< identifier >()) {
      value = new (ctx.mem) String_Constant(p, lexed);
    }
    else if (lex_css< quoted_string >()) {
      value = parse_interpolated_chunk(lexed, true); // needed!
    }
    else {
      error("expected a string constant or identifier in attribute selector for " + name, pstate);
    }

    if (!lex_css< exactly<']'> >()) error("unterminated attribute selector for " + name, pstate);
    return new (ctx.mem) Attribute_Selector(p, name, matcher, value);
  }

  /* parse block comment and add to block */
  void Parser::parse_block_comments(Block* block)
  {
    while (lex< block_comment >()) {
      bool is_important = lexed.begin[2] == '!';
      String*  contents = parse_interpolated_chunk(lexed);
      (*block) << new (ctx.mem) Comment(pstate, contents, is_important);
    }
  }

  Block* Parser::parse_block()
  {
    lex< exactly<'{'> >();
    bool semicolon = false;
    Selector_Lookahead lookahead_result;
    Block* block = new (ctx.mem) Block(pstate);
    block_stack.push_back(block);
    lex< zero_plus < alternatives < space, line_comment > > >();
    // JMA - ensure that a block containing only block_comments is parsed
    parse_block_comments(block);

    while (!lex< exactly<'}'> >()) {
      parse_block_comments(block);
      if (semicolon) {
        if (!lex< one_plus< exactly<';'> > >()) {
          error("non-terminal statement or declaration must end with ';'", pstate);
        }
        semicolon = false;
        parse_block_comments(block);
        if (lex< sequence< exactly<'}'>, zero_plus< exactly<';'> > > >()) break;
      }
      else if (peek< kwd_import >(position)) {
        if (stack.back() == mixin_def || stack.back() == function_def) {
          lex< kwd_import >(); // to adjust the before_token number
          error("@import directives are not allowed inside mixins and functions", pstate);
        }
        Import* imp = parse_import();
        if (!imp->urls().empty()) (*block) << imp;
        if (!imp->files().empty()) {
          for (size_t i = 0, S = imp->files().size(); i < S; ++i) {
            (*block) << new (ctx.mem) Import_Stub(pstate, imp->files()[i]);
          }
        }
        semicolon = true;
      }
      else if (lex< variable >()) {
        (*block) << parse_assignment();
        semicolon = true;
      }
      else if (lex< line_comment >()) {
        // throw line comments away
      }
      else if (peek< kwd_if_directive >()) {
        (*block) << parse_if_directive();
      }
      else if (peek< kwd_for_directive >()) {
        (*block) << parse_for_directive();
      }
      else if (peek< kwd_each_directive >()) {
        (*block) << parse_each_directive();
      }
      else if (peek < kwd_while_directive >()) {
        (*block) << parse_while_directive();
      }
      else if (lex < kwd_return_directive >()) {
        (*block) << new (ctx.mem) Return(pstate, parse_list());
        semicolon = true;
      }
      else if (peek< kwd_warn >()) {
        (*block) << parse_warning();
        semicolon = true;
      }
      else if (peek< kwd_err >()) {
        (*block) << parse_error();
        semicolon = true;
      }
      else if (peek< kwd_dbg >()) {
        (*block) << parse_debug();
        semicolon = true;
      }
      else if (stack.back() == function_def) {
        error("only variable declarations and control directives are allowed inside functions", pstate);
      }
      else if (peek< kwd_mixin >() || peek< kwd_function >()) {
        (*block) << parse_definition();
      }
      else if (peek< kwd_include >(position)) {
        Mixin_Call* the_call = parse_mixin_call();
        (*block) << the_call;
        // don't need a semicolon after a content block
        semicolon = (the_call->block()) ? false : true;
      }
      else if (lex< kwd_content >()) {
        if (stack.back() != mixin_def) {
          error("@content may only be used within a mixin", pstate);
        }
        (*block) << new (ctx.mem) Content(pstate);
        semicolon = true;
      }
      /*
      else if (peek< exactly<'+'> >()) {
        (*block) << parse_mixin_call();
        semicolon = true;
      }
      */
      else if (lex< kwd_extend >()) {
        Selector_Lookahead lookahead = lookahead_for_extension_target(position);
        if (!lookahead.found) error("invalid selector for @extend", pstate);
        Selector* target;
        if (lookahead.has_interpolants) target = parse_selector_schema(lookahead.found);
        else                            target = parse_selector_group();
        (*block) << new (ctx.mem) Extension(pstate, target);
        semicolon = true;
      }
      else if (peek< kwd_media >()) {
        (*block) << parse_media_block();
      }
      else if (peek< kwd_supports >()) {
        (*block) << parse_feature_block();
      }
      else if (peek< kwd_at_root >()) {
        (*block) << parse_at_root_block();
      }
      // ignore the @charset directive for now
      else if (lex< exactly< charset_kwd > >()) {
        lex< quoted_string >();
        lex< one_plus< exactly<';'> > >();
      }
      else if (peek< at_keyword >()) {
        At_Rule* at_rule = parse_at_rule();
        (*block) << at_rule;
        if (!at_rule->block()) semicolon = true;
      }
      else if ((lookahead_result = lookahead_for_selector(position)).found) {
        (*block) << parse_ruleset(lookahead_result);
      }/* not used anymore - remove?
      else if (peek< sequence< optional< exactly<'*'> >, alternatives< identifier_schema, identifier >, optional_spaces, exactly<':'>, optional_spaces, exactly<'{'> > >(position)) {
        (*block) << parse_propset();
      }*/
      else if (!peek< exactly<';'> >()) {
        bool indent = ! peek< sequence< optional< exactly<'*'> >, alternatives< identifier_schema, identifier >, optional_spaces, exactly<':'>, optional_spaces, exactly<'{'> > >(position);
        /* not used anymore - remove?
        if (peek< sequence< optional< exactly<'*'> >, identifier_schema, exactly<':'>, exactly<'{'> > >()) {
          (*block) << parse_propset();
        }
        else if (peek< sequence< optional< exactly<'*'> >, identifier, exactly<':'>, exactly<'{'> > >()) {
          (*block) << parse_propset();
        }
        else */ {
          Declaration* decl = parse_declaration();
          decl->tabs(indentation);
          (*block) << decl;
          if (peek< exactly<'{'> >()) {
            // parse a propset that rides on the declaration's property
            if (indent) indentation++;
            Propset* ps = new (ctx.mem) Propset(pstate, decl->property(), parse_block());
            if (indent) indentation--;
            (*block) << ps;
          }
          else {
            // finish and let the semicolon get munched
            semicolon = true;
          }
        }
      }
      else lex< one_plus< exactly<';'> > >();
      parse_block_comments(block);
    }
    block_stack.pop_back();
    return block;
  }

  Declaration* Parser::parse_declaration() {
    String* prop = 0;
    if (peek< sequence< optional< exactly<'*'> >, identifier_schema > >()) {
      prop = parse_identifier_schema();
    }
    else if (lex< sequence< optional< exactly<'*'> >, identifier > >()) {
      prop = new (ctx.mem) String_Quoted(pstate, lexed);
      prop->is_delayed(true);
    }
    else {
      error("invalid property name", pstate);
    }
    const string property(lexed);
    if (!lex_css< one_plus< exactly<':'> > >()) error("property \"" + property + "\" must be followed by a ':'", pstate);
    if (peek_css< exactly<';'> >()) error("style declaration must contain a value", pstate);
    if (peek_css< static_value >()) {
      return new (ctx.mem) Declaration(prop->pstate(), prop, parse_static_value()/*, lex<important>()*/);
    }
    else {
      Expression* value;
      Selector_Lookahead lookahead = lookahead_for_value(position);
      if (lookahead.found) {
        if (lookahead.has_interpolants) {
          value = parse_value_schema(lookahead.found);
        } else {
          value = parse_list();
        }
      }
      else {
        value = parse_list();
        if (List* list = dynamic_cast<List*>(value)) {
          if (list->length() == 0 && !peek< exactly <'{'> >()) {
            css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
          }
        }
      }

      return new (ctx.mem) Declaration(prop->pstate(), prop, value/*, lex<important>()*/);
    }
  }

  // parse +/- and return false if negative
  bool Parser::parse_number_prefix()
  {
    bool positive = true;
    while(true) {
      if (lex < block_comment >()) continue;
      if (lex < number_prefix >()) continue;
      if (lex < exactly < '-' > >()) {
        positive = !positive;
        continue;
      }
      break;
    }
    return positive;
  }

  Expression* Parser::parse_map()
  {
    ParserState opstate = pstate;
    Expression* key = parse_list();
    if (String_Quoted* str = dynamic_cast<String_Quoted*>(key)) {
      if (!str->quote_mark() && !str->is_delayed()) {
        if (ctx.names_to_colors.count(str->value())) {
          Color* c = new (ctx.mem) Color(*ctx.names_to_colors[str->value()]);
          c->pstate(str->pstate());
          c->disp(str->value());
          key = c;
        }
      }
    }

    // it's not a map so return the lexed value as a list value
    if (!peek< exactly<':'> >())
    { return key; }

    lex< exactly<':'> >();

    Expression* value = parse_space_list();

    Map* map = new (ctx.mem) Map(opstate, 1);
    (*map) << make_pair(key, value);

    while (lex_css< exactly<','> >())
    {
      // allow trailing commas - #495
      if (peek_css< exactly<')'> >(position))
      { break; }

      Expression* key = parse_list();
      if (String_Quoted* str = dynamic_cast<String_Quoted*>(key)) {
        if (!str->quote_mark() && !str->is_delayed()) {
          if (ctx.names_to_colors.count(str->value())) {
            Color* c = new (ctx.mem) Color(*ctx.names_to_colors[str->value()]);
            c->pstate(str->pstate());
            c->disp(str->value());
            key = c;
          }
        }
      }

      if (!(lex< exactly<':'> >()))
      { error("invalid syntax", pstate); }

      Expression* value = parse_space_list();

      (*map) << make_pair(key, value);
    }

    ParserState ps = map->pstate();
    ps.offset = pstate - ps + pstate.offset;
    map->pstate(ps);

    return map;
  }

  Expression* Parser::parse_list()
  {
    return parse_comma_list();
  }

  Expression* Parser::parse_comma_list()
  {
    if (peek_css< alternatives <
          // exactly<'!'>,
          // exactly<':'>,
          exactly<';'>,
          exactly<'}'>,
          exactly<'{'>,
          exactly<')'>,
          exactly<ellipsis>
        > >(position))
    { return new (ctx.mem) List(pstate, 0); }
    Expression* list1 = parse_space_list();
    // if it's a singleton, return it directly; don't wrap it
    if (!peek_css< exactly<','> >(position)) return list1;

    List* comma_list = new (ctx.mem) List(pstate, 2, List::COMMA);
    (*comma_list) << list1;

    while (lex_css< exactly<','> >())
    {
      if (peek_css< alternatives <
            // exactly<'!'>,
            exactly<';'>,
            exactly<'}'>,
            exactly<'{'>,
            exactly<')'>,
            exactly<':'>,
            exactly<ellipsis>
          > >(position)
      ) { break; }
      Expression* list = parse_space_list();
      (*comma_list) << list;
    }

    return comma_list;
  }

  Expression* Parser::parse_space_list()
  {
    Expression* disj1 = parse_disjunction();
    // if it's a singleton, return it directly; don't wrap it
    if (peek_css< alternatives <
          // exactly<'!'>,
          exactly<';'>,
          exactly<'}'>,
          exactly<'{'>,
          exactly<')'>,
          exactly<','>,
          exactly<':'>,
          exactly<ellipsis>,
          default_flag,
          global_flag
        > >(position)
    ) { return disj1; }

    List* space_list = new (ctx.mem) List(pstate, 2, List::SPACE);
    (*space_list) << disj1;

    while (!(peek_css< alternatives <
               // exactly<'!'>,
               exactly<';'>,
               exactly<'}'>,
               exactly<'{'>,
               exactly<')'>,
               exactly<','>,
               exactly<':'>,
               exactly<ellipsis>,
               default_flag,
               global_flag
           > >(position)) && peek_css< optional_css_whitespace >() != end
    ) {
      (*space_list) << parse_disjunction();
    }

    return space_list;
  }

  Expression* Parser::parse_disjunction()
  {
    Expression* conj1 = parse_conjunction();
    // if it's a singleton, return it directly; don't wrap it
    if (!peek_css< kwd_or >()) return conj1;

    vector<Expression*> operands;
    while (lex_css< kwd_or >())
      operands.push_back(parse_conjunction());

    return fold_operands(conj1, operands, Binary_Expression::OR);
  }

  Expression* Parser::parse_conjunction()
  {
    Expression* rel1 = parse_relation();
    // if it's a singleton, return it directly; don't wrap it
    if (!peek_css< kwd_and >()) return rel1;

    vector<Expression*> operands;
    while (lex_css< kwd_and >())
      operands.push_back(parse_relation());

    return fold_operands(rel1, operands, Binary_Expression::AND);
  }

  Expression* Parser::parse_relation()
  {
    Expression* expr1 = parse_expression();
    // if it's a singleton, return it directly; don't wrap it
    if (!(peek< alternatives <
            kwd_eq,
            kwd_neq,
            kwd_gte,
            kwd_gt,
            kwd_lte,
            kwd_lt
          > >(position)))
    { return expr1; }

    Binary_Expression::Type op
    = lex<kwd_eq>()  ? Binary_Expression::EQ
    : lex<kwd_neq>() ? Binary_Expression::NEQ
    : lex<kwd_gte>() ? Binary_Expression::GTE
    : lex<kwd_lte>() ? Binary_Expression::LTE
    : lex<kwd_gt>()  ? Binary_Expression::GT
    : lex<kwd_lt>()  ? Binary_Expression::LT
    :                 Binary_Expression::LT; // whatever

    Expression* expr2 = parse_expression();

    return new (ctx.mem) Binary_Expression(expr1->pstate(), op, expr1, expr2);
  }

  Expression* Parser::parse_expression()
  {
    Expression* term1 = parse_term();
    // if it's a singleton, return it directly; don't wrap it
    if (!(peek< exactly<'+'> >(position) ||
          (peek< no_spaces >(position) && peek< sequence< negate< unsigned_number >, exactly<'-'>, negate< space > > >(position)) ||
          (peek< sequence< negate< unsigned_number >, exactly<'-'>, negate< unsigned_number > > >(position))) ||
          peek< identifier >(position))
    { return term1; }

    vector<Expression*> operands;
    vector<Binary_Expression::Type> operators;
    while (lex< exactly<'+'> >() || lex< sequence< negate< digit >, exactly<'-'> > >()) {
      operators.push_back(lexed.to_string() == "+" ? Binary_Expression::ADD : Binary_Expression::SUB);
      operands.push_back(parse_term());
    }

    return fold_operands(term1, operands, operators);
  }

  Expression* Parser::parse_term()
  {
    Expression* factor = parse_factor();
    // Special case: Ruby sass never tries to modulo if the lhs contains an interpolant
    if (peek_css< exactly<'%'> >(position) && factor->concrete_type() == Expression::STRING) {
      String_Schema* ss = dynamic_cast<String_Schema*>(factor);
      if (ss && ss->has_interpolants()) return factor;
    }
    // if it's a singleton, return it directly; don't wrap it
    if (!peek< class_char< static_ops > >(position)) return factor;
    return parse_operators(factor);
  }

  Expression* Parser::parse_operators(Expression* factor)
  {
    // parse more factors and operators
    vector<Expression*> operands; // factors
    vector<Binary_Expression::Type> operators; // ops
    while (lex_css< class_char< static_ops > >()) {
      switch(*lexed.begin) {
        case '*': operators.push_back(Binary_Expression::MUL); break;
        case '/': operators.push_back(Binary_Expression::DIV); break;
        case '%': operators.push_back(Binary_Expression::MOD); break;
        default: throw runtime_error("unknown static op parsed"); break;
      }
      operands.push_back(parse_factor());
    }
    // operands and operators to binary expression
    return fold_operands(factor, operands, operators);
  }

  Expression* Parser::parse_factor()
  {
    if (lex_css< exactly<'('> >()) {
      Expression* value = parse_map();
      if (!lex_css< exactly<')'> >()) error("unclosed parenthesis", pstate);
      value->is_delayed(false);
      // make sure wrapped lists and division expressions are non-delayed within parentheses
      if (value->concrete_type() == Expression::LIST) {
        List* l = static_cast<List*>(value);
        if (!l->empty()) (*l)[0]->is_delayed(false);
      } else if (typeid(*value) == typeid(Binary_Expression)) {
        Binary_Expression* b = static_cast<Binary_Expression*>(value);
        Binary_Expression* lhs = static_cast<Binary_Expression*>(b->left());
        if (lhs && lhs->type() == Binary_Expression::DIV) lhs->is_delayed(false);
      }
      return value;
    }
    else if (peek< ie_property >()) {
      return parse_ie_property();
    }
    else if (peek< ie_keyword_arg >()) {
      return parse_ie_keyword_arg();
    }
    else if (peek< exactly< calc_kwd > >() ||
             peek< exactly< moz_calc_kwd > >() ||
             peek< exactly< ms_calc_kwd > >() ||
             peek< exactly< webkit_calc_kwd > >()) {
      return parse_calc_function();
    }
    else if (peek< functional_schema >()) {
      return parse_function_call_schema();
    }
    else if (peek< sequence< identifier_schema, negate< exactly<'%'> > > >()) {
      return parse_identifier_schema();
    }
    else if (peek< functional >()) {
      return parse_function_call();
    }
    else if (lex< sequence< exactly<'+'>, optional_css_whitespace, negate< number > > >()) {
      return new (ctx.mem) Unary_Expression(pstate, Unary_Expression::PLUS, parse_factor());
    }
    else if (lex< sequence< exactly<'-'>, optional_css_whitespace, negate< number> > >()) {
      return new (ctx.mem) Unary_Expression(pstate, Unary_Expression::MINUS, parse_factor());
    }
    else if (lex< sequence< kwd_not, css_whitespace > >()) {
      return new (ctx.mem) Unary_Expression(pstate, Unary_Expression::NOT, parse_factor());
    }
    else if (peek < sequence < one_plus < alternatives < css_whitespace, exactly<'-'>, exactly<'+'> > >, number > >()) {
      if (parse_number_prefix()) return parse_value(); // prefix is positive
      return new (ctx.mem) Unary_Expression(pstate, Unary_Expression::MINUS, parse_value());
    }
    else {
      return parse_value();
    }
  }

  Expression* Parser::parse_value()
  {
    lex< css_comments >();
    if (lex< ampersand >())
    {
      return new (ctx.mem) Parent_Selector(pstate, parse_selector_group()); }

    if (lex< important >())
    { return new (ctx.mem) String_Constant(pstate, "!important"); }

    const char* stop;
    if ((stop = peek< value_schema >()))
    { return parse_value_schema(stop); }

    if (lex< kwd_true >())
    { return new (ctx.mem) Boolean(pstate, true); }

    if (lex< kwd_false >())
    { return new (ctx.mem) Boolean(pstate, false); }

    if (lex< kwd_null >())
    { return new (ctx.mem) Null(pstate); }

    if (lex< identifier >()) {
      String_Constant* str = new (ctx.mem) String_Quoted(pstate, lexed);
      // Dont' delay this string if it is a name color. Fixes #652.
      str->is_delayed(ctx.names_to_colors.count(unquote(lexed)) == 0);
      return str;
    }

    if (lex< percentage >())
    { return new (ctx.mem) Textual(pstate, Textual::PERCENTAGE, lexed); }

    // match hex number first because 0x000 looks like a number followed by an indentifier
    if (lex< alternatives< hex, hex0 > >())
    { return new (ctx.mem) Textual(pstate, Textual::HEX, lexed); }

    // also handle the 10em- foo special case
    if (lex< sequence< dimension, optional< sequence< exactly<'-'>, negate< digit > > > > >())
    { return new (ctx.mem) Textual(pstate, Textual::DIMENSION, lexed); }

    if (lex< number >())
    { return new (ctx.mem) Textual(pstate, Textual::NUMBER, lexed); }

    if (peek< quoted_string >())
    { return parse_string(); }

    if (lex< variable >())
    { return new (ctx.mem) Variable(pstate, Util::normalize_underscores(lexed)); }

    // Special case handling for `%` proceeding an interpolant.
    if (lex< sequence< exactly<'%'>, optional< percentage > > >())
    { return new (ctx.mem) String_Quoted(pstate, lexed); }

    error("error reading values after " + lexed.to_string(), pstate);

    // unreachable statement
    return 0;
  }

  // this parses interpolation inside other strings
  // means the result should later be quoted again
  String* Parser::parse_interpolated_chunk(Token chunk, bool constant)
  {
    const char* i = chunk.begin;
    // see if there any interpolants
    const char* p = find_first_in_interval< exactly<hash_lbrace> >(i, chunk.end);
    if (!p) {
      String_Quoted* str_quoted = new (ctx.mem) String_Quoted(pstate, string(i, chunk.end));
      if (!constant && str_quoted->quote_mark()) str_quoted->quote_mark('*');
      str_quoted->is_delayed(true);
      return str_quoted;
    }

    String_Schema* schema = new (ctx.mem) String_Schema(pstate);
    while (i < chunk.end) {
      p = find_first_in_interval< exactly<hash_lbrace> >(i, chunk.end);
      if (p) {
        if (i < p) {
          // accumulate the preceding segment if it's nonempty
          (*schema) << new (ctx.mem) String_Quoted(pstate, string(i, p));
        }
        // we need to skip anything inside strings
        // create a new target in parser/prelexer
        if (peek < sequence < optional_spaces, exactly<rbrace> > >(p+2)) { position = p+2;
          css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
        }
        const char* j = skip_over_scopes< exactly<hash_lbrace>, exactly<rbrace> >(p + 2, chunk.end); // find the closing brace
        if (j) { --j;
          // parse the interpolant and accumulate it
          Expression* interp_node = Parser::from_token(Token(p+2, j), ctx, pstate).parse_list();
          interp_node->is_interpolant(true);
          (*schema) << interp_node;
          i = j;
        }
        else {
          // throw an error if the interpolant is unterminated
          error("unterminated interpolant inside string constant " + chunk.to_string(), pstate);
        }
      }
      else { // no interpolants left; add the last segment if nonempty
        // check if we need quotes here (was not sure after merge)
        if (i < chunk.end) (*schema) << new (ctx.mem) String_Quoted(pstate, string(i, chunk.end));
        break;
      }
      ++ i;
    }
    return schema;
  }

  String_Constant* Parser::parse_static_expression()
  {
    if (peek< sequence< number, optional_spaces, exactly<'/'>, optional_spaces, number > >()) {
      return parse_static_value();
    }
    return 0;
  }

  String_Constant* Parser::parse_static_value()
  {
    lex< static_value >();
    Token str(lexed);
    --str.end;
    --position;

    String_Constant* str_node = new (ctx.mem) String_Constant(pstate, str.time_wspace());
    str_node->is_delayed(true);
    return str_node;
  }

  String* Parser::parse_string()
  {
    lex< quoted_string >();
    Token token(lexed);
    return parse_interpolated_chunk(token);
  }

  String* Parser::parse_ie_property()
  {
    lex< ie_property >();
    Token str(lexed);
    const char* i = str.begin;
    // see if there any interpolants
    const char* p = find_first_in_interval< exactly<hash_lbrace> >(str.begin, str.end);
    if (!p) {
      String_Constant* str_node = new (ctx.mem) String_Constant(pstate, normalize_wspace(string(str.begin, str.end)));
      str_node->is_delayed(true);
      str_node->quote_mark('*');
      return str_node;
    }

    String_Schema* schema = new (ctx.mem) String_Schema(pstate);
    while (i < str.end) {
      p = find_first_in_interval< exactly<hash_lbrace> >(i, str.end);
      if (p) {
        if (i < p) {
          String_Constant* part = new (ctx.mem) String_Constant(pstate, normalize_wspace(string(i, p))); // accumulate the preceding segment if it's nonempty
          part->is_delayed(true);
          part->quote_mark('*'); // avoid unquote in interpolation
          (*schema) << part;
        }
        if (peek < sequence < optional_spaces, exactly<rbrace> > >(p+2)) { position = p+2;
          css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
        }
        const char* j = skip_over_scopes< exactly<hash_lbrace>, exactly<rbrace> >(p+2, str.end); // find the closing brace
        if (j) {
          // parse the interpolant and accumulate it
          Expression* interp_node = Parser::from_token(Token(p+2, j), ctx, pstate).parse_list();
          interp_node->is_interpolant(true);
          (*schema) << interp_node;
          i = j;
        }
        else {
          // throw an error if the interpolant is unterminated
          error("unterminated interpolant inside IE function " + str.to_string(), pstate);
        }
      }
      else { // no interpolants left; add the last segment if nonempty
        if (i < str.end) {
          String_Constant* part = new (ctx.mem) String_Constant(pstate, normalize_wspace(string(i, str.end)));
          part->is_delayed(true);
          part->quote_mark('*'); // avoid unquote in interpolation
          (*schema) << part;
        }
        break;
      }
    }
    return schema;
  }

  String* Parser::parse_ie_keyword_arg()
  {
    String_Schema* kwd_arg = new (ctx.mem) String_Schema(pstate, 3);
    if (lex< variable >()) {
      *kwd_arg << new (ctx.mem) Variable(pstate, Util::normalize_underscores(lexed));
    } else {
      lex< alternatives< identifier_schema, identifier > >();
      *kwd_arg << new (ctx.mem) String_Constant(pstate, lexed);
    }
    lex< exactly<'='> >();
    *kwd_arg << new (ctx.mem) String_Constant(pstate, lexed);
    if (peek< variable >()) *kwd_arg << parse_list();
    else if (lex< number >()) *kwd_arg << new (ctx.mem) Textual(pstate, Textual::NUMBER, Util::normalize_decimals(lexed));
    else if (peek < ie_keyword_arg_value >()) { *kwd_arg << parse_list(); }
    return kwd_arg;
  }

  String_Schema* Parser::parse_value_schema(const char* stop)
  {
    String_Schema* schema = new (ctx.mem) String_Schema(pstate);
    size_t num_items = 0;
    if (peek<exactly<'}'>>()) {
      css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
    }
    while (position < stop) {
      if (lex< spaces >() && num_items) {
        (*schema) << new (ctx.mem) String_Constant(pstate, " ");
      }
      else if (lex< interpolant >()) {
        Token insides(Token(lexed.begin + 2, lexed.end - 1));
        Expression* interp_node;
        Parser p = Parser::from_token(insides, ctx, pstate);
        if (!(interp_node = p.parse_static_expression())) {
          interp_node = p.parse_list();
          interp_node->is_interpolant(true);
        }
        (*schema) << interp_node;
      }
      else if (lex< exactly<'%'> >()) {
        (*schema) << new (ctx.mem) String_Constant(pstate, lexed);
      }
      else if (lex< identifier >()) {
        (*schema) << new (ctx.mem) String_Quoted(pstate, lexed);
      }
      else if (lex< percentage >()) {
        (*schema) << new (ctx.mem) Textual(pstate, Textual::PERCENTAGE, lexed);
      }
      else if (lex< dimension >()) {
        (*schema) << new (ctx.mem) Textual(pstate, Textual::DIMENSION, lexed);
      }
      else if (lex< number >()) {
        Expression* factor = new (ctx.mem) Textual(pstate, Textual::NUMBER, lexed);
        if (peek< class_char< static_ops > >()) {
          (*schema) << parse_operators(factor);
        } else {
          (*schema) << factor;
        }
      }
      else if (lex< hex >()) {
        (*schema) << new (ctx.mem) Textual(pstate, Textual::HEX, unquote(lexed));
      }
      else if (lex < exactly < '-' > >()) {
        (*schema) << new (ctx.mem) String_Constant(pstate, lexed);
      }
      else if (lex< quoted_string >()) {
        (*schema) << new (ctx.mem) String_Quoted(pstate, lexed);
      }
      else if (lex< variable >()) {
        (*schema) << new (ctx.mem) Variable(pstate, Util::normalize_underscores(lexed));
      }
      else if (peek< parenthese_scope >()) {
        (*schema) << parse_factor();
      }
      else {
        error("error parsing interpolated value", pstate);
      }
      ++num_items;
    }
    return schema;
  }

  /* not used anymore - remove?
  String_Schema* Parser::parse_url_schema()
  {
    String_Schema* schema = new (ctx.mem) String_Schema(pstate);

    while (position < end) {
      if (position[0] == '/') {
        lexed = Token(position, position+1, before_token);
        (*schema) << new (ctx.mem) String_Quoted(pstate, lexed);
        ++position;
      }
      else if (lex< interpolant >()) {
        Token insides(Token(lexed.begin + 2, lexed.end - 1, before_token));
        Expression* interp_node = Parser::from_token(insides, ctx, pstate).parse_list();
        interp_node->is_interpolant(true);
        (*schema) << interp_node;
      }
      else if (lex< sequence< identifier, exactly<':'> > >()) {
        (*schema) << new (ctx.mem) String_Quoted(pstate, lexed);
      }
      else if (lex< filename >()) {
        (*schema) << new (ctx.mem) String_Quoted(pstate, lexed);
      }
      else {
        error("error parsing interpolated url", pstate);
      }
    }
    return schema;
  } */

  // this parses interpolation outside other strings
  // means the result must not be quoted again later
  String* Parser::parse_identifier_schema()
  {
    // first lex away whatever we have found
    lex< sequence< optional< exactly<'*'> >, identifier_schema > >();
    Token id(lexed);
    const char* i = id.begin;
    // see if there any interpolants
    const char* p = find_first_in_interval< exactly<hash_lbrace> >(id.begin, id.end);
    if (!p) {
      return new (ctx.mem) String_Quoted(pstate, string(id.begin, id.end));
    }

    String_Schema* schema = new (ctx.mem) String_Schema(pstate);
    while (i < id.end) {
      p = find_first_in_interval< exactly<hash_lbrace> >(i, id.end);
      if (p) {
        if (i < p) {
          // accumulate the preceding segment if it's nonempty
          (*schema) << new (ctx.mem) String_Constant(pstate, string(i, p));
        }
        // we need to skip anything inside strings
        // create a new target in parser/prelexer
        if (peek < sequence < optional_spaces, exactly<rbrace> > >(p+2)) { position = p+2;
          css_error("Invalid CSS", " after ", ": expected expression (e.g. 1px, bold), was ");
        }
        const char* j = skip_over_scopes< exactly<hash_lbrace>, exactly<rbrace> >(p+2, id.end); // find the closing brace
        if (j) {
          // parse the interpolant and accumulate it
          Expression* interp_node = Parser::from_token(Token(p+2, j), ctx, pstate).parse_list();
          interp_node->is_interpolant(true);
          (*schema) << interp_node;
          schema->has_interpolants(true);
          i = j;
        }
        else {
          // throw an error if the interpolant is unterminated
          error("unterminated interpolant inside interpolated identifier " + id.to_string(), pstate);
        }
      }
      else { // no interpolants left; add the last segment if nonempty
        if (i < end) (*schema) << new (ctx.mem) String_Quoted(pstate, string(i, id.end));
        break;
      }
    }
    return schema;
  }

  Function_Call* Parser::parse_calc_function()
  {
    lex< identifier >();
    string name(lexed);
    ParserState call_pos = pstate;
    lex< exactly<'('> >();
    ParserState arg_pos = pstate;
    const char* arg_beg = position;
    parse_list();
    const char* arg_end = position;
    lex< exactly<')'> >();

    Argument* arg = new (ctx.mem) Argument(arg_pos, parse_interpolated_chunk(Token(arg_beg, arg_end)));
    Arguments* args = new (ctx.mem) Arguments(arg_pos);
    *args << arg;
    return new (ctx.mem) Function_Call(call_pos, name, args);
  }

  Function_Call* Parser::parse_function_call()
  {
    lex< identifier >();
    string name(lexed);

    ParserState call_pos = pstate;
    Arguments* args = parse_arguments(name == "url");
    return new (ctx.mem) Function_Call(call_pos, name, args);
  }

  Function_Call_Schema* Parser::parse_function_call_schema()
  {
    String* name = parse_identifier_schema();
    ParserState source_position_of_call = pstate;

    Function_Call_Schema* the_call = new (ctx.mem) Function_Call_Schema(source_position_of_call, name, parse_arguments());
    return the_call;
  }

  If* Parser::parse_if_directive(bool else_if)
  {
    lex< kwd_if_directive >() || (else_if && lex< exactly<if_after_else_kwd> >());
    ParserState if_source_position = pstate;
    Expression* predicate = parse_list();
    predicate->is_delayed(false);
    if (!peek< exactly<'{'> >()) error("expected '{' after the predicate for @if", pstate);
    Block* consequent = parse_block();
    Block* alternative = 0;

    if (lex< elseif_directive >()) {
      alternative = new (ctx.mem) Block(pstate);
      (*alternative) << parse_if_directive(true);
    }
    else if (lex< kwd_else_directive >()) {
      if (!peek< exactly<'{'> >()) {
        error("expected '{' after @else", pstate);
      }
      else {
        alternative = parse_block();
      }
    }
    return new (ctx.mem) If(if_source_position, predicate, consequent, alternative);
  }

  For* Parser::parse_for_directive()
  {
    lex< kwd_for_directive >();
    ParserState for_source_position = pstate;
    if (!lex< variable >()) error("@for directive requires an iteration variable", pstate);
    string var(Util::normalize_underscores(lexed));
    if (!lex< kwd_from >()) error("expected 'from' keyword in @for directive", pstate);
    Expression* lower_bound = parse_expression();
    lower_bound->is_delayed(false);
    bool inclusive = false;
    if (lex< kwd_through >()) inclusive = true;
    else if (lex< kwd_to >()) inclusive = false;
    else                  error("expected 'through' or 'to' keyword in @for directive", pstate);
    Expression* upper_bound = parse_expression();
    upper_bound->is_delayed(false);
    if (!peek< exactly<'{'> >()) error("expected '{' after the upper bound in @for directive", pstate);
    Block* body = parse_block();
    return new (ctx.mem) For(for_source_position, var, lower_bound, upper_bound, body, inclusive);
  }

  Each* Parser::parse_each_directive()
  {
    lex < kwd_each_directive >();
    ParserState each_source_position = pstate;
    if (!lex< variable >()) error("@each directive requires an iteration variable", pstate);
    vector<string> vars;
    vars.push_back(Util::normalize_underscores(lexed));
    while (lex< exactly<','> >()) {
      if (!lex< variable >()) error("@each directive requires an iteration variable", pstate);
      vars.push_back(Util::normalize_underscores(lexed));
    }
    if (!lex< kwd_in >()) error("expected 'in' keyword in @each directive", pstate);
    Expression* list = parse_list();
    list->is_delayed(false);
    if (list->concrete_type() == Expression::LIST) {
      List* l = static_cast<List*>(list);
      for (size_t i = 0, L = l->length(); i < L; ++i) {
        (*l)[i]->is_delayed(false);
      }
    }
    if (!peek< exactly<'{'> >()) error("expected '{' after the upper bound in @each directive", pstate);
    Block* body = parse_block();
    return new (ctx.mem) Each(each_source_position, vars, list, body);
  }

  While* Parser::parse_while_directive()
  {
    lex< kwd_while_directive >();
    ParserState while_source_position = pstate;
    Expression* predicate = parse_list();
    predicate->is_delayed(false);
    Block* body = parse_block();
    return new (ctx.mem) While(while_source_position, predicate, body);
  }

  Media_Block* Parser::parse_media_block()
  {
    lex< kwd_media >();
    ParserState media_source_position = pstate;

    List* media_queries = parse_media_queries();

    if (!peek< exactly<'{'> >()) {
      error("expected '{' in media query", pstate);
    }
    Media_Block* media_block = new (ctx.mem) Media_Block(media_source_position, media_queries, 0);
    Media_Block* prev_media_block = last_media_block;
    last_media_block = media_block;
    media_block->block(parse_block());
    last_media_block = prev_media_block;

    return media_block;
  }

  List* Parser::parse_media_queries()
  {
    List* media_queries = new (ctx.mem) List(pstate, 0, List::COMMA);
    if (!peek< exactly<'{'> >()) (*media_queries) << parse_media_query();
    while (lex< exactly<','> >()) (*media_queries) << parse_media_query();
    return media_queries;
  }

  // Expression* Parser::parse_media_query()
  Media_Query* Parser::parse_media_query()
  {
    Media_Query* media_query = new (ctx.mem) Media_Query(pstate);

    if (lex< exactly< not_kwd > >()) media_query->is_negated(true);
    else if (lex< exactly< only_kwd > >()) media_query->is_restricted(true);

    if (peek< identifier_schema >()) media_query->media_type(parse_identifier_schema());
    else if (lex< identifier >())    media_query->media_type(parse_interpolated_chunk(lexed));
    else                             (*media_query) << parse_media_expression();

    while (lex< exactly< and_kwd > >()) (*media_query) << parse_media_expression();
    if (peek< identifier_schema >()) {
      String_Schema* schema = new (ctx.mem) String_Schema(pstate);
      *schema << media_query->media_type();
      *schema << new (ctx.mem) String_Constant(pstate, " ");
      *schema << parse_identifier_schema();
      media_query->media_type(schema);
    }
    while (lex< exactly< and_kwd > >()) (*media_query) << parse_media_expression();
    return media_query;
  }

  Media_Query_Expression* Parser::parse_media_expression()
  {
    if (peek< identifier_schema >()) {
      String* ss = parse_identifier_schema();
      return new (ctx.mem) Media_Query_Expression(pstate, ss, 0, true);
    }
    if (!lex< exactly<'('> >()) {
      error("media query expression must begin with '('", pstate);
    }
    Expression* feature = 0;
    if (peek< exactly<')'> >()) {
      error("media feature required in media query expression", pstate);
    }
    feature = parse_expression();
    Expression* expression = 0;
    if (lex< exactly<':'> >()) {
      expression = parse_list();
    }
    if (!lex< exactly<')'> >()) {
      error("unclosed parenthesis in media query expression", pstate);
    }
    return new (ctx.mem) Media_Query_Expression(feature->pstate(), feature, expression);
  }

  Feature_Block* Parser::parse_feature_block()
  {
    lex< kwd_supports >();
    ParserState supports_source_position = pstate;

    Feature_Query* feature_queries = parse_feature_queries();

    if (!peek< exactly<'{'> >()) {
      error("expected '{' in feature query", pstate);
    }
    Block* block = parse_block();

    return new (ctx.mem) Feature_Block(supports_source_position, feature_queries, block);
  }

  Feature_Query* Parser::parse_feature_queries()
  {
    Feature_Query* fq = new (ctx.mem) Feature_Query(pstate);
    Feature_Query_Condition* cond = new (ctx.mem) Feature_Query_Condition(pstate);
    cond->is_root(true);
    while (!peek< exactly<')'> >(position) && !peek< exactly<'{'> >(position))
      (*cond) << parse_feature_query();
    (*fq) << cond;

    if (fq->empty()) error("expected @supports condition (e.g. (display: flexbox))", pstate);

    return fq;
  }

  Feature_Query_Condition* Parser::parse_feature_query()
  {
    if (peek< kwd_not >(position)) return parse_supports_negation();
    else if (peek< kwd_and >(position)) return parse_supports_conjunction();
    else if (peek< kwd_or >(position)) return parse_supports_disjunction();
    else if (peek< exactly<'('> >(position)) return parse_feature_query_in_parens();
    else return parse_supports_declaration();
  }

  Feature_Query_Condition* Parser::parse_feature_query_in_parens()
  {
    Feature_Query_Condition* cond = new (ctx.mem) Feature_Query_Condition(pstate);

    if (!lex< exactly<'('> >()) error("@supports declaration expected '('", pstate);
    while (!peek< exactly<')'> >(position) && !peek< exactly<'{'> >(position))
      (*cond) << parse_feature_query();
    if (!lex< exactly<')'> >()) error("unclosed parenthesis in @supports declaration", pstate);

    return (cond->length() == 1) ? (*cond)[0] : cond;
  }

  Feature_Query_Condition* Parser::parse_supports_negation()
  {
    lex< kwd_not >();

    Feature_Query_Condition* cond = parse_feature_query();
    cond->operand(Feature_Query_Condition::NOT);

    return cond;
  }

  Feature_Query_Condition* Parser::parse_supports_conjunction()
  {
    lex< kwd_and >();

    Feature_Query_Condition* cond = parse_feature_query();
    cond->operand(Feature_Query_Condition::AND);

    return cond;
  }

  Feature_Query_Condition* Parser::parse_supports_disjunction()
  {
    lex< kwd_or >();

    Feature_Query_Condition* cond = parse_feature_query();
    cond->operand(Feature_Query_Condition::OR);

    return cond;
  }

  Feature_Query_Condition* Parser::parse_supports_declaration()
  {
    Declaration* declaration = parse_declaration();
    Feature_Query_Condition* cond = new (ctx.mem) Feature_Query_Condition(declaration->pstate(),
                                                                          1,
                                                                          declaration->property(),
                                                                          declaration->value());
    return cond;
  }

  At_Root_Block* Parser::parse_at_root_block()
  {
    lex<kwd_at_root>();
    ParserState at_source_position = pstate;
    Block* body = 0;
    At_Root_Expression* expr = 0;
    Selector_Lookahead lookahead_result;
    in_at_root = true;
    if (peek< exactly<'('> >()) {
      expr = parse_at_root_expression();
      body = parse_block();
    }
    else if (peek< exactly<'{'> >()) {
      body = parse_block();
    }
    else if ((lookahead_result = lookahead_for_selector(position)).found) {
      Ruleset* r = parse_ruleset(lookahead_result);
      body = new (ctx.mem) Block(r->pstate(), 1);
      *body << r;
    }
    in_at_root = false;
    At_Root_Block* at_root = new (ctx.mem) At_Root_Block(at_source_position, body);
    if (expr) at_root->expression(expr);
    return at_root;
  }

  At_Root_Expression* Parser::parse_at_root_expression()
  {
    lex< exactly<'('> >();
    if (peek< exactly<')'> >()) error("at-root feature required in at-root expression", pstate);

    if (!peek< alternatives< kwd_with_directive, kwd_without_directive > >()) {
      css_error("Invalid CSS", " after ", ": expected \"with\" or \"without\", was ");
    }

    Declaration* declaration = parse_declaration();
    List* value = new (ctx.mem) List(declaration->value()->pstate(), 1);

    if (declaration->value()->concrete_type() == Expression::LIST) {
        value = static_cast<List*>(declaration->value());
    }
    else *value << declaration->value();

    At_Root_Expression* cond = new (ctx.mem) At_Root_Expression(declaration->pstate(),
                                                                declaration->property(),
                                                                value);
    if (!lex< exactly<')'> >()) error("unclosed parenthesis in @at-root expression", pstate);
    return cond;
  }

  At_Rule* Parser::parse_at_rule()
  {
    lex<at_keyword>();
    string kwd(lexed);
    ParserState at_source_position = pstate;
    Selector* sel = 0;
    Expression* val = 0;
    Selector_Lookahead lookahead = lookahead_for_extension_target(position);
    if (lookahead.found) {
      if (lookahead.has_interpolants) {
        sel = parse_selector_schema(lookahead.found);
      }
      else {
        sel = parse_selector_group();
      }
    }
    else if (!(peek<exactly<'{'> >() || peek<exactly<'}'> >() || peek<exactly<';'> >())) {
      val = parse_list();
    }
    Block* body = 0;
    if (peek< exactly<'{'> >()) body = parse_block();
    At_Rule* rule = new (ctx.mem) At_Rule(at_source_position, kwd, sel, body);
    if (!sel) rule->value(val);
    return rule;
  }

  Warning* Parser::parse_warning()
  {
    lex< kwd_warn >();
    return new (ctx.mem) Warning(pstate, parse_list());
  }

  Error* Parser::parse_error()
  {
    lex< kwd_err >();
    return new (ctx.mem) Error(pstate, parse_list());
  }

  Debug* Parser::parse_debug()
  {
    lex< kwd_dbg >();
    return new (ctx.mem) Debug(pstate, parse_list());
  }

  Selector_Lookahead Parser::lookahead_for_selector(const char* start)
  {
    const char* p = start ? start : position;
    const char* q;
    bool saw_stuff = false;
    bool saw_interpolant = false;

    while ((q = peek< identifier >(p))                             ||
           (q = peek< hyphens_and_identifier >(p))                 ||
           (q = peek< hyphens_and_name >(p))                       ||
           (q = peek< type_selector >(p))                          ||
           (q = peek< id_name >(p))                                ||
           (q = peek< class_name >(p))                             ||
           (q = peek< sequence< pseudo_prefix, identifier > >(p))  ||
           (q = peek< percentage >(p))                             ||
           (q = peek< variable >(p))                            ||
           (q = peek< dimension >(p))                              ||
           (q = peek< quoted_string >(p))                          ||
           (q = peek< exactly<'*'> >(p))                           ||
           (q = peek< exactly<sel_deep_kwd> >(p))                           ||
           (q = peek< exactly<'('> >(p))                           ||
           (q = peek< exactly<')'> >(p))                           ||
           (q = peek< exactly<'['> >(p))                           ||
           (q = peek< exactly<']'> >(p))                           ||
           (q = peek< exactly<'+'> >(p))                           ||
           (q = peek< exactly<'~'> >(p))                           ||
           (q = peek< exactly<'>'> >(p))                           ||
           (q = peek< exactly<','> >(p))                           ||
           (saw_stuff && (q = peek< exactly<'-'> >(p)))            ||
           (q = peek< binomial >(p))                               ||
           (q = peek< block_comment >(p))                          ||
           (q = peek< sequence< optional<sign>,
                                zero_plus<digit>,
                                exactly<'n'> > >(p))               ||
           (q = peek< sequence< optional<sign>,
                                one_plus<digit> > >(p))                     ||
           (q = peek< number >(p))                                 ||
           (q = peek< sequence< exactly<'&'>,
                                identifier_alnums > >(p))        ||
           (q = peek< exactly<'&'> >(p))                           ||
           (q = peek< exactly<'%'> >(p))                           ||
           (q = peek< alternatives<exact_match,
                                   class_match,
                                   dash_match,
                                   prefix_match,
                                   suffix_match,
                                   substring_match> >(p))          ||
           (q = peek< sequence< exactly<'.'>, interpolant > >(p))  ||
           (q = peek< sequence< exactly<'#'>, interpolant > >(p))  ||
           (q = peek< sequence< one_plus< exactly<'-'> >, interpolant > >(p))  ||
           (q = peek< sequence< pseudo_prefix, interpolant > >(p)) ||
           (q = peek< interpolant >(p))) {
      saw_stuff = true;
      p = q;
      if (*(p - 1) == '}') saw_interpolant = true;
    }

    Selector_Lookahead result;
    result.found            = saw_stuff && peek< exactly<'{'> >(p) ? p : 0;
    result.has_interpolants = saw_interpolant;

    return result;
  }

  Selector_Lookahead Parser::lookahead_for_extension_target(const char* start)
  {
    const char* p = start ? start : position;
    const char* q;
    bool saw_interpolant = false;
    bool saw_stuff = false;

    while ((q = peek< identifier >(p))                             ||
           (q = peek< type_selector >(p))                          ||
           (q = peek< id_name >(p))                                ||
           (q = peek< class_name >(p))                             ||
           (q = peek< sequence< pseudo_prefix, identifier > >(p))  ||
           (q = peek< percentage >(p))                             ||
           (q = peek< dimension >(p))                              ||
           (q = peek< quoted_string >(p))                          ||
           (q = peek< exactly<'*'> >(p))                           ||
           (q = peek< exactly<'('> >(p))                           ||
           (q = peek< exactly<')'> >(p))                           ||
           (q = peek< exactly<'['> >(p))                           ||
           (q = peek< exactly<']'> >(p))                           ||
           (q = peek< exactly<'+'> >(p))                           ||
           (q = peek< exactly<'~'> >(p))                           ||
           (q = peek< exactly<'>'> >(p))                           ||
           (q = peek< exactly<','> >(p))                           ||
           (saw_stuff && (q = peek< exactly<'-'> >(p)))            ||
           (q = peek< binomial >(p))                               ||
           (q = peek< block_comment >(p))                          ||
           (q = peek< sequence< optional<sign>,
                                zero_plus<digit>,
                                exactly<'n'> > >(p))               ||
           (q = peek< sequence< optional<sign>,
                                one_plus<digit> > >(p))                     ||
           (q = peek< number >(p))                                 ||
           (q = peek< sequence< exactly<'&'>,
                                identifier_alnums > >(p))        ||
           (q = peek< exactly<'&'> >(p))                           ||
           (q = peek< exactly<'%'> >(p))                           ||
           (q = peek< alternatives<exact_match,
                                   class_match,
                                   dash_match,
                                   prefix_match,
                                   suffix_match,
                                   substring_match> >(p))          ||
           (q = peek< sequence< exactly<'.'>, interpolant > >(p))  ||
           (q = peek< sequence< exactly<'#'>, interpolant > >(p))  ||
           (q = peek< sequence< one_plus< exactly<'-'> >, interpolant > >(p))  ||
           (q = peek< sequence< pseudo_prefix, interpolant > >(p)) ||
           (q = peek< interpolant >(p))                            ||
           (q = peek< optional >(p))) {
      p = q;
      if (*(p - 1) == '}') saw_interpolant = true;
      saw_stuff = true;
    }

    Selector_Lookahead result;
    result.found            = peek< alternatives< exactly<';'>, exactly<'}'>, exactly<'{'> > >(p) && saw_stuff ? p : 0;
    result.has_interpolants = saw_interpolant;

    return result;
  }


  Selector_Lookahead Parser::lookahead_for_value(const char* start)
  {
    const char* p = start ? start : position;
    const char* q;
    bool saw_interpolant = false;
    bool saw_stuff = false;

    while ((q = peek< identifier >(p))                             ||
           (q = peek< percentage >(p))                             ||
           (q = peek< dimension >(p))                              ||
           (q = peek< quoted_string >(p))                          ||
           (q = peek< variable >(p))                               ||
           (q = peek< exactly<'*'> >(p))                           ||
           (q = peek< exactly<'+'> >(p))                           ||
           (q = peek< exactly<'~'> >(p))                           ||
           (q = peek< exactly<'>'> >(p))                           ||
           (q = peek< exactly<','> >(p))                           ||
           (q = peek< sequence<parenthese_scope, interpolant>>(p)) ||
           (saw_stuff && (q = peek< exactly<'-'> >(p)))            ||
           (q = peek< binomial >(p))                               ||
           (q = peek< block_comment >(p))                          ||
           (q = peek< sequence< optional<sign>,
                                zero_plus<digit>,
                                exactly<'n'> > >(p))               ||
           (q = peek< sequence< optional<sign>,
                                one_plus<digit> > >(p))            ||
           (q = peek< number >(p))                                 ||
           (q = peek< sequence< exactly<'&'>,
                                identifier_alnums > >(p))          ||
           (q = peek< exactly<'&'> >(p))                           ||
           (q = peek< exactly<'%'> >(p))                           ||
           (q = peek< sequence< exactly<'.'>, interpolant > >(p))  ||
           (q = peek< sequence< exactly<'#'>, interpolant > >(p))  ||
           (q = peek< sequence< one_plus< exactly<'-'> >, interpolant > >(p))  ||
           (q = peek< sequence< pseudo_prefix, interpolant > >(p)) ||
           (q = peek< interpolant >(p))                            ||
           (q = peek< optional >(p))) {
      p = q;
      if (*(p - 1) == '}') saw_interpolant = true;
      saw_stuff = true;
    }

    Selector_Lookahead result;
    result.found            = peek< alternatives< exactly<';'>, exactly<'}'>, exactly<'{'> > >(p) && saw_stuff ? p : 0;
    result.has_interpolants = saw_interpolant;

    return result;
  }

  void Parser::read_bom()
  {
    size_t skip = 0;
    string encoding;
    bool utf_8 = false;
    switch ((unsigned char) source[0]) {
    case 0xEF:
      skip = check_bom_chars(source, end, utf_8_bom, 3);
      encoding = "UTF-8";
      utf_8 = true;
      break;
    case 0xFE:
      skip = check_bom_chars(source, end, utf_16_bom_be, 2);
      encoding = "UTF-16 (big endian)";
      break;
    case 0xFF:
      skip = check_bom_chars(source, end, utf_16_bom_le, 2);
      skip += (skip ? check_bom_chars(source, end, utf_32_bom_le, 4) : 0);
      encoding = (skip == 2 ? "UTF-16 (little endian)" : "UTF-32 (little endian)");
      break;
    case 0x00:
      skip = check_bom_chars(source, end, utf_32_bom_be, 4);
      encoding = "UTF-32 (big endian)";
      break;
    case 0x2B:
      skip = check_bom_chars(source, end, utf_7_bom_1, 4)
           | check_bom_chars(source, end, utf_7_bom_2, 4)
           | check_bom_chars(source, end, utf_7_bom_3, 4)
           | check_bom_chars(source, end, utf_7_bom_4, 4)
           | check_bom_chars(source, end, utf_7_bom_5, 5);
      encoding = "UTF-7";
      break;
    case 0xF7:
      skip = check_bom_chars(source, end, utf_1_bom, 3);
      encoding = "UTF-1";
      break;
    case 0xDD:
      skip = check_bom_chars(source, end, utf_ebcdic_bom, 4);
      encoding = "UTF-EBCDIC";
      break;
    case 0x0E:
      skip = check_bom_chars(source, end, scsu_bom, 3);
      encoding = "SCSU";
      break;
    case 0xFB:
      skip = check_bom_chars(source, end, bocu_1_bom, 3);
      encoding = "BOCU-1";
      break;
    case 0x84:
      skip = check_bom_chars(source, end, gb_18030_bom, 4);
      encoding = "GB-18030";
      break;
    }
    if (skip > 0 && !utf_8) error("only UTF-8 documents are currently supported; your document appears to be " + encoding, pstate);
    position += skip;
  }

  size_t check_bom_chars(const char* src, const char *end, const unsigned char* bom, size_t len)
  {
    size_t skip = 0;
    if (src + len > end) return 0;
    for (size_t i = 0; i < len; ++i, ++skip) {
      if ((unsigned char) src[i] != bom[i]) return 0;
    }
    return skip;
  }


  Expression* Parser::fold_operands(Expression* base, vector<Expression*>& operands, Binary_Expression::Type op)
  {
    for (size_t i = 0, S = operands.size(); i < S; ++i) {
      base = new (ctx.mem) Binary_Expression(pstate, op, base, operands[i]);
      Binary_Expression* b = static_cast<Binary_Expression*>(base);
      if (op == Binary_Expression::DIV && b->left()->is_delayed() && b->right()->is_delayed()) {
        base->is_delayed(true);
      }
      else {
        b->left()->is_delayed(false);
        b->right()->is_delayed(false);
      }
    }
    return base;
  }

  Expression* Parser::fold_operands(Expression* base, vector<Expression*>& operands, vector<Binary_Expression::Type>& ops)
  {
    for (size_t i = 0, S = operands.size(); i < S; ++i) {
      base = new (ctx.mem) Binary_Expression(base->pstate(), ops[i], base, operands[i]);
      Binary_Expression* b = static_cast<Binary_Expression*>(base);
      if (ops[i] == Binary_Expression::DIV && b->left()->is_delayed() && b->right()->is_delayed()) {
        base->is_delayed(true);
      }
      else {
        b->left()->is_delayed(false);
        b->right()->is_delayed(false);
      }
    }
    return base;
  }

  void Parser::error(string msg, Position pos)
  {
    throw Sass_Error(Sass_Error::syntax, ParserState(path, source, pos.line ? pos : before_token, Offset(0, 0)), msg);
  }

  // print a css parsing error with actual context information from parsed source
  void Parser::css_error(const string& msg, const string& prefix, const string& middle)
  {
    int max_len = 14;
    const char* pos = peek < optional_spaces >();
    bool ellipsis_left = false;
    const char* pos_left(pos);
    while (*pos_left && pos_left > source) {
      if (pos - pos_left > max_len) {
        ellipsis_left = true;
        break;
      }
      const char* prev = pos_left - 1;
      if (*prev == '\r') break;
      if (*prev == '\n') break;
      if (*prev == 10) break;
      pos_left = prev;
    }
    bool ellipsis_right = false;
    const char* pos_right(pos);
    while (*pos_right && pos_right <= end) {
      if (pos_right - pos > max_len) {
        ellipsis_right = true;
        break;
      }
      if (*pos_right == '\r') break;
      if (*pos_right == '\n') break;
      if (*pos_left == 10) break;
      ++ pos_right;
    }
    string left(pos_left, pos);
    string right(pos, pos_right);
    if (ellipsis_left) left = ellipsis + left;
    if (ellipsis_right) right = right + ellipsis;
    // now pass new message to the more generic error function
    error(msg + prefix + quote(left) + middle + quote(right), pstate);
  }

}
