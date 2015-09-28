#include <cctype>
#include <cstddef>
#include <iostream>
#include <iomanip>
#include "util.hpp"
#include "position.hpp"
#include "prelexer.hpp"
#include "constants.hpp"


namespace Sass {
  // using namespace Lexer;
  using namespace Constants;

  namespace Prelexer {


    // Match a line comment (/.*?(?=\n|\r\n?|\Z)/.
    const char* line_comment(const char* src)
    {
      return sequence<
               exactly <
                 slash_slash
               >,
               non_greedy<
                 any_char,
                 end_of_line
               >
             >(src);
    }

    // Match a block comment.
    const char* block_comment(const char* src)
    {
      return sequence<
               zero_plus < space >,
               delimited_by<
                 slash_star,
                 star_slash,
                 false
               >
             >(src);
    }
    /* not use anymore - remove?
    const char* block_comment_prefix(const char* src) {
      return exactly<slash_star>(src);
    }
    // Match either comment.
    const char* comment(const char* src) {
      return line_comment(src);
    }
    */

    // Match zero plus white-space or line_comments
    const char* optional_css_whitespace(const char* src) {
      return zero_plus< alternatives<spaces, line_comment> >(src);
    }
    const char* css_whitespace(const char* src) {
      return one_plus< alternatives<spaces, line_comment> >(src);
    }
    // Match optional_css_whitepace plus block_comments
    const char* optional_css_comments(const char* src) {
      return zero_plus< alternatives<spaces, line_comment, block_comment> >(src);
    }
    const char* css_comments(const char* src) {
      return one_plus< alternatives<spaces, line_comment, block_comment> >(src);
    }

    // Match one backslash escaped char /\\./
    const char* escape_seq(const char* src)
    {
      return sequence<
        exactly<'\\'>,
        any_char
      >(src);
    }

    // Match identifier start
    const char* identifier_alpha(const char* src)
    {
      return alternatives<
               alpha,
               unicode,
               exactly<'-'>,
               exactly<'_'>,
               escape_seq
             >(src);
    }

    // Match identifier after start
    const char* identifier_alnum(const char* src)
    {
      return alternatives<
               alnum,
               unicode,
               exactly<'-'>,
               exactly<'_'>,
               escape_seq
             >(src);
    }

    // Match CSS identifiers.
    const char* identifier(const char* src)
    {
      return sequence<
               zero_plus< exactly<'-'> >,
               one_plus < identifier_alpha >,
               zero_plus < identifier_alnum >
               // word_boundary not needed
             >(src);
    }

    const char* identifier_alnums(const char* src)
    {
      return one_plus< identifier_alnum >(src);
    }

    // Match number prefix ([\+\-]+)
    const char* number_prefix(const char* src) {
      return alternatives <
        exactly < '+' >,
        sequence <
          exactly < '-' >,
          optional_css_whitespace,
          exactly< '-' >
        >
      >(src);
    }

    // Match interpolant schemas
    const char* identifier_schema(const char* src) {
      // follows this pattern: (x*ix*)+ ... well, not quite
      return sequence< one_plus< sequence< zero_plus< alternatives< identifier, exactly<'-'> > >,
                                 interpolant,
                                 zero_plus< alternatives< identifier, number, exactly<'-'> > > > >,
                       negate< exactly<'%'> > >(src);
    }

    // interpolants can be recursive/nested
    const char* interpolant(const char* src) {
      return recursive_scopes< exactly<hash_lbrace>, exactly<rbrace> >(src);
    }

    // $re_squote = /'(?:$re_itplnt|\\.|[^'])*'/
    const char* single_quoted_string(const char* src) {
      // match a single quoted string, while skipping interpolants
      return sequence <
        exactly <'\''>,
        zero_plus <
          alternatives <
            // skip escapes
            sequence <
              exactly < '\\' >,
              exactly < '\r' >,
              exactly < '\n' >
            >,
            escape_seq,
            // skip interpolants
            interpolant,
            // skip non delimiters
            any_char_but < '\'' >
          >
        >,
        exactly <'\''>
      >(src);
    }

    // $re_dquote = /"(?:$re_itp|\\.|[^"])*"/
    const char* double_quoted_string(const char* src) {
      // match a single quoted string, while skipping interpolants
      return sequence <
        exactly <'"'>,
        zero_plus <
          alternatives <
            // skip escapes
            sequence <
              exactly < '\\' >,
              exactly < '\r' >,
              exactly < '\n' >
            >,
            escape_seq,
            // skip interpolants
            interpolant,
            // skip non delimiters
            any_char_but < '"' >
          >
        >,
        exactly <'"'>
      >(src);
    }

    // $re_quoted = /(?:$re_squote|$re_dquote)/
    const char* quoted_string(const char* src) {
      // match a quoted string, while skipping interpolants
      return alternatives<
        single_quoted_string,
        double_quoted_string
      >(src);
    }

    const char* value_schema(const char* src) {
      // follows this pattern: ([xyz]*i[xyz]*)+
      return one_plus< sequence< zero_plus< alternatives< identifier, percentage, dimension, hex, number, quoted_string > >,
                                 interpolant,
                                 zero_plus< alternatives< identifier, percentage, dimension, hex, number, quoted_string, exactly<'%'> > > > >(src);
    }

    /* not used anymore - remove?
    const char* filename(const char* src) {
      return one_plus< alternatives< identifier, number, exactly<'.'> > >(src);
    }
    */

    // Match CSS '@' keywords.
    const char* at_keyword(const char* src) {
      return sequence<exactly<'@'>, identifier>(src);
    }

    const char* kwd_sel_deep(const char* src) {
      return word<sel_deep_kwd>(src);
    }

    const char* kwd_import(const char* src) {
      return word<import_kwd>(src);
    }

    const char* kwd_at_root(const char* src) {
      return word<at_root_kwd>(src);
    }

    const char* kwd_with_directive(const char* src) {
      return word<with_kwd>(src);
    }

    const char* kwd_without_directive(const char* src) {
      return word<without_kwd>(src);
    }

    const char* kwd_media(const char* src) {
      return word<media_kwd>(src);
    }

    const char* kwd_supports(const char* src) {
      return word<supports_kwd>(src);
    }

    const char* kwd_mixin(const char* src) {
      return word<mixin_kwd>(src);
    }

    const char* kwd_function(const char* src) {
      return word<function_kwd>(src);
    }

    const char* kwd_return_directive(const char* src) {
      return word<return_kwd>(src);
    }

    const char* kwd_include(const char* src) {
      return word<include_kwd>(src);
    }

    const char* kwd_content(const char* src) {
      return word<content_kwd>(src);
    }

    const char* kwd_extend(const char* src) {
      return word<extend_kwd>(src);
    }


    const char* kwd_if_directive(const char* src) {
      return word<if_kwd>(src);
    }

    const char* kwd_else_directive(const char* src) {
      return word<else_kwd>(src);
    }
    const char* elseif_directive(const char* src) {
      return sequence< exactly< else_kwd >,
                                optional_css_whitespace,
                                word< if_after_else_kwd > >(src);
    }

    const char* kwd_for_directive(const char* src) {
      return word<for_kwd>(src);
    }

    const char* kwd_from(const char* src) {
      return word<from_kwd>(src);
    }

    const char* kwd_to(const char* src) {
      return word<to_kwd>(src);
    }

    const char* kwd_through(const char* src) {
      return word<through_kwd>(src);
    }

    const char* kwd_each_directive(const char* src) {
      return word<each_kwd>(src);
    }

    const char* kwd_in(const char* src) {
      return word<in_kwd>(src);
    }

    const char* kwd_while_directive(const char* src) {
      return word<while_kwd>(src);
    }

    const char* name(const char* src) {
      return one_plus< alternatives< alnum,
                                     exactly<'-'>,
                                     exactly<'_'>,
                                     exactly<'\\'> > >(src);
    }

    const char* kwd_warn(const char* src) {
      return word<warn_kwd>(src);
    }

    const char* kwd_err(const char* src) {
      return word<error_kwd>(src);
    }

    const char* kwd_dbg(const char* src) {
      return word<debug_kwd>(src);
    }

    /* not used anymore - remove?
    const char* directive(const char* src) {
      return sequence< exactly<'@'>, identifier >(src);
    } */

    const char* kwd_null(const char* src) {
      return word<null_kwd>(src);
    }

    // Match CSS type selectors
    const char* namespace_prefix(const char* src) {
      return sequence< optional< alternatives< identifier, exactly<'*'> > >,
                       exactly<'|'> >(src);
    }
    const char* type_selector(const char* src) {
      return sequence< optional<namespace_prefix>, identifier>(src);
    }
    const char* hyphens_and_identifier(const char* src) {
      return sequence< zero_plus< exactly< '-' > >, identifier >(src);
    }
    const char* hyphens_and_name(const char* src) {
      return sequence< zero_plus< exactly< '-' > >, name >(src);
    }
    const char* universal(const char* src) {
      return sequence< optional<namespace_prefix>, exactly<'*'> >(src);
    }
    // Match CSS id names.
    const char* id_name(const char* src) {
      return sequence<exactly<'#'>, name>(src);
    }
    // Match CSS class names.
    const char* class_name(const char* src) {
      return sequence<exactly<'.'>, identifier>(src);
    }
    // Attribute name in an attribute selector.
    const char* attribute_name(const char* src) {
      return alternatives< sequence< optional<namespace_prefix>, identifier>,
                           identifier >(src);
    }
    // match placeholder selectors
    const char* placeholder(const char* src) {
      return sequence<exactly<'%'>, identifier>(src);
    }
    // Match CSS numeric constants.

    const char* sign(const char* src) {
      return class_char<sign_chars>(src);
    }
    const char* unsigned_number(const char* src) {
      return alternatives<sequence< zero_plus<digits>,
                                    exactly<'.'>,
                                    one_plus<digits> >,
                          digits>(src);
    }
    const char* number(const char* src) {
      return sequence< optional<sign>, unsigned_number>(src);
    }
    const char* coefficient(const char* src) {
      return alternatives< sequence< optional<sign>, digits >,
                           sign >(src);
    }
    const char* binomial(const char* src) {
      return sequence< optional<sign>,
                       optional<digits>,
                       exactly<'n'>,
                       zero_plus < space >,
                       sign,
                       zero_plus < space >,
                       digits >(src);
    }
    const char* percentage(const char* src) {
      return sequence< number, exactly<'%'> >(src);
    }
    const char* ampersand(const char* src) {
      return exactly<'&'>(src);
    }

    /* not used anymore - remove?
    const char* em(const char* src) {
      return sequence< number, exactly<em_kwd> >(src);
    } */
    const char* dimension(const char* src) {
      return sequence<number, one_plus< alpha > >(src);
    }
    const char* hex(const char* src) {
      const char* p = sequence< exactly<'#'>, one_plus<xdigit> >(src);
      ptrdiff_t len = p - src;
      return (len != 4 && len != 7) ? 0 : p;
    }
    const char* hexa(const char* src) {
      const char* p = sequence< exactly<'#'>, one_plus<xdigit> >(src);
      ptrdiff_t len = p - src;
      return (len != 4 && len != 7 && len != 9) ? 0 : p;
    }
    const char* hex0(const char* src) {
      const char* p = sequence< exactly<'0'>, exactly<'x'>, one_plus<xdigit> >(src);
      ptrdiff_t len = p - src;
      return (len != 5 && len != 8) ? 0 : p;
    }

    /* no longer used - remove?
    const char* rgb_prefix(const char* src) {
      return word<rgb_kwd>(src);
    }*/
    // Match CSS uri specifiers.

    const char* uri_prefix(const char* src) {
      return exactly<url_kwd>(src);
    }
    const char* uri_value(const char* src)
    {
      return
      sequence <
        negate <
          exactly < '$' >
        >,
        zero_plus <
          alternatives <
            alnum,
            exactly <'/'>,
            class_char < uri_chars >
          >
        >
      >(src);
    }

    // TODO: rename the following two functions
    /* no longer used - remove?
    const char* uri(const char* src) {
      return sequence< exactly<url_kwd>,
                       optional<spaces>,
                       quoted_string,
                       optional<spaces>,
                       exactly<')'> >(src);
    }*/
    /* no longer used - remove?
    const char* url_value(const char* src) {
      return sequence< optional< sequence< identifier, exactly<':'> > >, // optional protocol
                       one_plus< sequence< zero_plus< exactly<'/'> >, filename > >, // one or more folders and/or trailing filename
                       optional< exactly<'/'> > >(src);
    }*/
    /* no longer used - remove?
    const char* url_schema(const char* src) {
      return sequence< optional< sequence< identifier, exactly<':'> > >, // optional protocol
                       filename_schema >(src); // optional trailing slash
    }*/
    // Match CSS "!important" keyword.
    const char* important(const char* src) {
      return sequence< exactly<'!'>,
                       optional_css_whitespace,
                       word<important_kwd> >(src);
    }
    // Match CSS "!optional" keyword.
    const char* optional(const char* src) {
      return sequence< exactly<'!'>,
      optional_css_whitespace,
      word<optional_kwd> >(src);
    }
    // Match Sass "!default" keyword.
    const char* default_flag(const char* src) {
      return sequence< exactly<'!'>,
                       optional_css_whitespace,
                       word<default_kwd> >(src);
    }
    // Match Sass "!global" keyword.
    const char* global_flag(const char* src) {
      return sequence< exactly<'!'>,
                       optional_css_whitespace,
                       word<global_kwd> >(src);
    }
    // Match CSS pseudo-class/element prefixes.
    const char* pseudo_prefix(const char* src) {
      return sequence< exactly<':'>, optional< exactly<':'> > >(src);
    }
    // Match CSS function call openers.
    const char* functional_schema(const char* src) {
      return sequence< identifier_schema, exactly<'('> >(src);
    }
    const char* functional(const char* src) {
      return sequence< identifier, exactly<'('> >(src);
    }
    // Match the CSS negation pseudo-class.
    const char* pseudo_not(const char* src) {
      return word< pseudo_not_kwd >(src);
    }
    // Match CSS 'odd' and 'even' keywords for functional pseudo-classes.
    const char* even(const char* src) {
      return word<even_kwd>(src);
    }
    const char* odd(const char* src) {
      return word<odd_kwd>(src);
    }
    // Match CSS attribute-matching operators.
    const char* exact_match(const char* src) { return exactly<'='>(src); }
    const char* class_match(const char* src) { return exactly<tilde_equal>(src); }
    const char* dash_match(const char* src) { return exactly<pipe_equal>(src); }
    const char* prefix_match(const char* src) { return exactly<caret_equal>(src); }
    const char* suffix_match(const char* src) { return exactly<dollar_equal>(src); }
    const char* substring_match(const char* src) { return exactly<star_equal>(src); }
    // Match CSS combinators.
    /* not used anymore - remove?
    const char* adjacent_to(const char* src) {
      return sequence< optional_spaces, exactly<'+'> >(src);
    }
    const char* precedes(const char* src) {
      return sequence< optional_spaces, exactly<'~'> >(src);
    }
    const char* parent_of(const char* src) {
      return sequence< optional_spaces, exactly<'>'> >(src);
    }
    const char* ancestor_of(const char* src) {
      return sequence< spaces, negate< exactly<'{'> > >(src);
    }*/

    // Match SCSS variable names.
    const char* variable(const char* src) {
      return sequence<exactly<'$'>, identifier>(src);
    }

    // Match Sass boolean keywords.
    const char* kwd_true(const char* src) {
      return word<true_kwd>(src);
    }
    const char* kwd_false(const char* src) {
      return word<false_kwd>(src);
    }
    const char* kwd_and(const char* src) {
      return word<and_kwd>(src);
    }
    const char* kwd_or(const char* src) {
      return word<or_kwd>(src);
    }
    const char* kwd_not(const char* src) {
      return word<not_kwd>(src);
    }
    const char* kwd_eq(const char* src) {
      return exactly<eq>(src);
    }
    const char* kwd_neq(const char* src) {
      return exactly<neq>(src);
    }
    const char* kwd_gt(const char* src) {
      return exactly<gt>(src);
    }
    const char* kwd_gte(const char* src) {
      return exactly<gte>(src);
    }
    const char* kwd_lt(const char* src) {
      return exactly<lt>(src);
    }
    const char* kwd_lte(const char* src) {
      return exactly<lte>(src);
    }

    // match specific IE syntax
    const char* ie_progid(const char* src) {
      return sequence <
        word<progid_kwd>,
        exactly<':'>,
        alternatives< identifier_schema, identifier >,
        zero_plus< sequence<
          exactly<'.'>,
          alternatives< identifier_schema, identifier >
        > >,
        zero_plus < sequence<
          exactly<'('>,
          optional_css_whitespace,
          optional < sequence<
            alternatives< variable, identifier_schema, identifier >,
            optional_css_whitespace,
            exactly<'='>,
            optional_css_whitespace,
            alternatives< variable, identifier_schema, identifier, quoted_string, number, hexa >,
            zero_plus< sequence<
              optional_css_whitespace,
              exactly<','>,
              optional_css_whitespace,
              sequence<
                alternatives< variable, identifier_schema, identifier >,
                optional_css_whitespace,
                exactly<'='>,
                optional_css_whitespace,
                alternatives< variable, identifier_schema, identifier, quoted_string, number, hexa >
              >
            > >
          > >,
          optional_css_whitespace,
          exactly<')'>
        > >
      >(src);
    }
    const char* ie_expression(const char* src) {
      return sequence < word<expression_kwd>, exactly<'('>, skip_over_scopes< exactly<'('>, exactly<')'> > >(src);
    }
    const char* ie_property(const char* src) {
      return alternatives < ie_expression, ie_progid >(src);
    }

    // const char* ie_args(const char* src) {
    //   return sequence< alternatives< ie_keyword_arg, value_schema, quoted_string, interpolant, number, identifier, delimited_by< '(', ')', true> >,
    //                    zero_plus< sequence< optional_css_whitespace, exactly<','>, optional_css_whitespace, alternatives< ie_keyword_arg, value_schema, quoted_string, interpolant, number, identifier, delimited_by<'(', ')', true> > > > >(src);
    // }

    const char* ie_keyword_arg_property(const char* src) {
      return alternatives <
          variable,
          identifier_schema,
          identifier
        >(src);
    }
    const char* ie_keyword_arg_value(const char* src) {
      return alternatives <
          variable,
          identifier_schema,
          identifier,
          quoted_string,
          number,
          hexa,
          sequence <
            exactly < '(' >,
            skip_over_scopes <
              exactly < '(' >,
              exactly < ')' >
            >
          >
        >(src);
    }

    const char* ie_keyword_arg(const char* src) {
      return sequence <
        ie_keyword_arg_property,
        optional_css_whitespace,
        exactly<'='>,
        optional_css_whitespace,
        ie_keyword_arg_value
      >(src);
    }

    // Path matching functions.
    /* not used anymore - remove?
    const char* folder(const char* src) {
      return sequence< zero_plus< any_char_except<'/'> >,
                       exactly<'/'> >(src);
    }
    const char* folders(const char* src) {
      return zero_plus< folder >(src);
    }*/
    /* not used anymore - remove?
    const char* chunk(const char* src) {
      char inside_str = 0;
      const char* p = src;
      size_t depth = 0;
      while (true) {
        if (!*p) {
          return 0;
        }
        else if (!inside_str && (*p == '"' || *p == '\'')) {
          inside_str = *p;
        }
        else if (*p == inside_str && *(p-1) != '\\') {
          inside_str = 0;
        }
        else if (*p == '(' && !inside_str) {
          ++depth;
        }
        else if (*p == ')' && !inside_str) {
          if (depth == 0) return p;
          else            --depth;
        }
        ++p;
      }
      // unreachable
      return 0;
    }
    */

    // follow the CSS spec more closely and see if this helps us scan URLs correctly
    /* not used anymore - remove?
    const char* NL(const char* src) {
      return alternatives< exactly<'\n'>,
                           sequence< exactly<'\r'>, exactly<'\n'> >,
                           exactly<'\r'>,
                           exactly<'\f'> >(src);
    }*/

    /* not used anymore - remove?
    const char* H(const char* src) {
      return std::isxdigit(*src) ? src+1 : 0;
    }*/

    /* not used anymore - remove?
    const char* unicode(const char* src) {
      return sequence< exactly<'\\'>,
                       between<H, 1, 6>,
                       optional< class_char<url_space_chars> > >(src);
    }*/

    /* not used anymore - remove?
    const char* ESCAPE(const char* src) {
      return alternatives< unicode, class_char<escape_chars> >(src);
    }*/

    const char* static_string(const char* src) {
      const char* pos = src;
      const char * s = quoted_string(pos);
      Token t(pos, s);
      const unsigned int p = count_interval< interpolant >(t.begin, t.end);
      return (p == 0) ? t.end : 0;
    }

    const char* static_component(const char* src) {
      return alternatives< identifier,
                           static_string,
                           percentage,
                           hex,
                           number,
                           sequence< exactly<'!'>, word<important_kwd> >
                          >(src);
    }

    const char* static_value(const char* src) {
      return sequence< sequence<
                         static_component,
                         zero_plus< identifier >
                       >,
                       zero_plus < sequence<
                                   alternatives<
                                     sequence< optional_spaces, alternatives<
                                       exactly < '/' >,
                                       exactly < ',' >,
                                       exactly < ' ' >
                                     >, optional_spaces >,
                                     spaces
                                   >,
                                   static_component
                       > >,
                       optional_css_whitespace,
                       alternatives< exactly<';'>, exactly<'}'> >
                      >(src);
    }

    const char* parenthese_scope(const char* src) {
      return sequence <
        exactly < '(' >,
        skip_over_scopes <
          exactly < '(' >,
          exactly < ')' >
        >
      >(src);
    }

  }
}
