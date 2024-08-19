#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 13
#define STATE_COUNT 41
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 33
#define ALIAS_COUNT 0
#define TOKEN_COUNT 15
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 8
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 9

enum ts_symbol_identifiers {
  sym_workspace_identifier = 1,
  anon_sym_LF = 2,
  anon_sym_definition = 3,
  anon_sym_reference = 4,
  anon_sym_forward_definition = 5,
  anon_sym_implements = 6,
  anon_sym_type_defines = 7,
  anon_sym_references = 8,
  anon_sym_relationships = 9,
  anon_sym_defined_by = 10,
  anon_sym_POUND = 11,
  aux_sym_comment_token1 = 12,
  anon_sym_POUNDdocstring_COLON = 13,
  anon_sym_global = 14,
  sym_source_file = 15,
  sym__statement = 16,
  sym_definition_statement = 17,
  sym_reference_statement = 18,
  sym__definition_relations = 19,
  sym_implementation_relation = 20,
  sym_type_definition_relation = 21,
  sym_references_relation = 22,
  sym_relationships_statement = 23,
  sym__all_relations = 24,
  sym_defined_by_relation = 25,
  sym_comment = 26,
  sym_docstring = 27,
  sym_identifier = 28,
  sym_global_identifier = 29,
  aux_sym_source_file_repeat1 = 30,
  aux_sym_definition_statement_repeat1 = 31,
  aux_sym_relationships_statement_repeat1 = 32,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_workspace_identifier] = "workspace_identifier",
  [anon_sym_LF] = "\n",
  [anon_sym_definition] = "definition",
  [anon_sym_reference] = "reference",
  [anon_sym_forward_definition] = "forward_definition",
  [anon_sym_implements] = "implements",
  [anon_sym_type_defines] = "type_defines",
  [anon_sym_references] = "references",
  [anon_sym_relationships] = "relationships",
  [anon_sym_defined_by] = "defined_by",
  [anon_sym_POUND] = "#",
  [aux_sym_comment_token1] = "comment_token1",
  [anon_sym_POUNDdocstring_COLON] = "# docstring:",
  [anon_sym_global] = "global",
  [sym_source_file] = "source_file",
  [sym__statement] = "_statement",
  [sym_definition_statement] = "definition_statement",
  [sym_reference_statement] = "reference_statement",
  [sym__definition_relations] = "_definition_relations",
  [sym_implementation_relation] = "implementation_relation",
  [sym_type_definition_relation] = "type_definition_relation",
  [sym_references_relation] = "references_relation",
  [sym_relationships_statement] = "relationships_statement",
  [sym__all_relations] = "_all_relations",
  [sym_defined_by_relation] = "defined_by_relation",
  [sym_comment] = "comment",
  [sym_docstring] = "docstring",
  [sym_identifier] = "identifier",
  [sym_global_identifier] = "global_identifier",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_definition_statement_repeat1] = "definition_statement_repeat1",
  [aux_sym_relationships_statement_repeat1] = "relationships_statement_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_workspace_identifier] = sym_workspace_identifier,
  [anon_sym_LF] = anon_sym_LF,
  [anon_sym_definition] = anon_sym_definition,
  [anon_sym_reference] = anon_sym_reference,
  [anon_sym_forward_definition] = anon_sym_forward_definition,
  [anon_sym_implements] = anon_sym_implements,
  [anon_sym_type_defines] = anon_sym_type_defines,
  [anon_sym_references] = anon_sym_references,
  [anon_sym_relationships] = anon_sym_relationships,
  [anon_sym_defined_by] = anon_sym_defined_by,
  [anon_sym_POUND] = anon_sym_POUND,
  [aux_sym_comment_token1] = aux_sym_comment_token1,
  [anon_sym_POUNDdocstring_COLON] = anon_sym_POUNDdocstring_COLON,
  [anon_sym_global] = anon_sym_global,
  [sym_source_file] = sym_source_file,
  [sym__statement] = sym__statement,
  [sym_definition_statement] = sym_definition_statement,
  [sym_reference_statement] = sym_reference_statement,
  [sym__definition_relations] = sym__definition_relations,
  [sym_implementation_relation] = sym_implementation_relation,
  [sym_type_definition_relation] = sym_type_definition_relation,
  [sym_references_relation] = sym_references_relation,
  [sym_relationships_statement] = sym_relationships_statement,
  [sym__all_relations] = sym__all_relations,
  [sym_defined_by_relation] = sym_defined_by_relation,
  [sym_comment] = sym_comment,
  [sym_docstring] = sym_docstring,
  [sym_identifier] = sym_identifier,
  [sym_global_identifier] = sym_global_identifier,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_definition_statement_repeat1] = aux_sym_definition_statement_repeat1,
  [aux_sym_relationships_statement_repeat1] = aux_sym_relationships_statement_repeat1,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [sym_workspace_identifier] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_LF] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_definition] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_reference] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_forward_definition] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_implements] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_type_defines] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_references] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_relationships] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_defined_by] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_POUND] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_comment_token1] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_POUNDdocstring_COLON] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_global] = {
    .visible = true,
    .named = false,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__statement] = {
    .visible = false,
    .named = true,
  },
  [sym_definition_statement] = {
    .visible = true,
    .named = true,
  },
  [sym_reference_statement] = {
    .visible = true,
    .named = true,
  },
  [sym__definition_relations] = {
    .visible = false,
    .named = true,
  },
  [sym_implementation_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_type_definition_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_references_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_relationships_statement] = {
    .visible = true,
    .named = true,
  },
  [sym__all_relations] = {
    .visible = false,
    .named = true,
  },
  [sym_defined_by_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_comment] = {
    .visible = true,
    .named = true,
  },
  [sym_docstring] = {
    .visible = true,
    .named = true,
  },
  [sym_identifier] = {
    .visible = true,
    .named = true,
  },
  [sym_global_identifier] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_definition_statement_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_relationships_statement_repeat1] = {
    .visible = false,
    .named = false,
  },
};

enum ts_field_identifiers {
  field_descriptors = 1,
  field_docstring = 2,
  field_forward_definition = 3,
  field_global = 4,
  field_name = 5,
  field_project_name = 6,
  field_roles = 7,
  field_workspace = 8,
};

static const char * const ts_field_names[] = {
  [0] = NULL,
  [field_descriptors] = "descriptors",
  [field_docstring] = "docstring",
  [field_forward_definition] = "forward_definition",
  [field_global] = "global",
  [field_name] = "name",
  [field_project_name] = "project_name",
  [field_roles] = "roles",
  [field_workspace] = "workspace",
};

static const TSFieldMapSlice ts_field_map_slices[PRODUCTION_ID_COUNT] = {
  [1] = {.index = 0, .length = 1},
  [2] = {.index = 1, .length = 1},
  [3] = {.index = 2, .length = 1},
  [4] = {.index = 3, .length = 2},
  [5] = {.index = 5, .length = 2},
  [6] = {.index = 7, .length = 2},
  [7] = {.index = 9, .length = 3},
  [8] = {.index = 12, .length = 4},
};

static const TSFieldMapEntry ts_field_map_entries[] = {
  [0] =
    {field_workspace, 0},
  [1] =
    {field_name, 1},
  [2] =
    {field_global, 0},
  [3] =
    {field_name, 1},
    {field_roles, 2},
  [5] =
    {field_forward_definition, 1},
    {field_name, 2},
  [7] =
    {field_descriptors, 2},
    {field_project_name, 1},
  [9] =
    {field_docstring, 0},
    {field_docstring, 1},
    {field_name, 3},
  [12] =
    {field_docstring, 0},
    {field_docstring, 1},
    {field_name, 3},
    {field_roles, 4},
};

static const TSSymbol ts_alias_sequences[PRODUCTION_ID_COUNT][MAX_ALIAS_SEQUENCE_LENGTH] = {
  [0] = {0},
};

static const uint16_t ts_non_terminal_alias_map[] = {
  0,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(42);
      if (lookahead == '#') ADVANCE(51);
      if (lookahead == 'd') ADVANCE(57);
      if (lookahead == 'r') ADVANCE(58);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(0)
      if (lookahead != 0) ADVANCE(82);
      END_STATE();
    case 1:
      if (lookahead == '\n') ADVANCE(43);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(1)
      if (lookahead != 0) ADVANCE(82);
      END_STATE();
    case 2:
      if (lookahead == '\n') ADVANCE(43);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(2)
      END_STATE();
    case 3:
      if (lookahead == ':') ADVANCE(54);
      END_STATE();
    case 4:
      if (lookahead == 'a') ADVANCE(38);
      END_STATE();
    case 5:
      if (lookahead == 'c') ADVANCE(34);
      END_STATE();
    case 6:
      if (lookahead == 'c') ADVANCE(10);
      END_STATE();
    case 7:
      if (lookahead == 'd') ADVANCE(28);
      END_STATE();
    case 8:
      if (lookahead == 'e') ADVANCE(14);
      END_STATE();
    case 9:
      if (lookahead == 'e') ADVANCE(13);
      END_STATE();
    case 10:
      if (lookahead == 'e') ADVANCE(46);
      END_STATE();
    case 11:
      if (lookahead == 'e') ADVANCE(33);
      END_STATE();
    case 12:
      if (lookahead == 'e') ADVANCE(25);
      END_STATE();
    case 13:
      if (lookahead == 'f') ADVANCE(11);
      if (lookahead == 'l') ADVANCE(4);
      END_STATE();
    case 14:
      if (lookahead == 'f') ADVANCE(20);
      END_STATE();
    case 15:
      if (lookahead == 'g') ADVANCE(3);
      END_STATE();
    case 16:
      if (lookahead == 'h') ADVANCE(18);
      END_STATE();
    case 17:
      if (lookahead == 'i') ADVANCE(23);
      END_STATE();
    case 18:
      if (lookahead == 'i') ADVANCE(31);
      END_STATE();
    case 19:
      if (lookahead == 'i') ADVANCE(29);
      END_STATE();
    case 20:
      if (lookahead == 'i') ADVANCE(27);
      END_STATE();
    case 21:
      if (lookahead == 'i') ADVANCE(30);
      END_STATE();
    case 22:
      if (lookahead == 'i') ADVANCE(39);
      END_STATE();
    case 23:
      if (lookahead == 'n') ADVANCE(15);
      END_STATE();
    case 24:
      if (lookahead == 'n') ADVANCE(44);
      END_STATE();
    case 25:
      if (lookahead == 'n') ADVANCE(6);
      END_STATE();
    case 26:
      if (lookahead == 'n') ADVANCE(35);
      END_STATE();
    case 27:
      if (lookahead == 'n') ADVANCE(22);
      END_STATE();
    case 28:
      if (lookahead == 'o') ADVANCE(5);
      END_STATE();
    case 29:
      if (lookahead == 'o') ADVANCE(26);
      END_STATE();
    case 30:
      if (lookahead == 'o') ADVANCE(24);
      END_STATE();
    case 31:
      if (lookahead == 'p') ADVANCE(36);
      END_STATE();
    case 32:
      if (lookahead == 'r') ADVANCE(17);
      END_STATE();
    case 33:
      if (lookahead == 'r') ADVANCE(12);
      END_STATE();
    case 34:
      if (lookahead == 's') ADVANCE(37);
      END_STATE();
    case 35:
      if (lookahead == 's') ADVANCE(16);
      END_STATE();
    case 36:
      if (lookahead == 's') ADVANCE(48);
      END_STATE();
    case 37:
      if (lookahead == 't') ADVANCE(32);
      END_STATE();
    case 38:
      if (lookahead == 't') ADVANCE(19);
      END_STATE();
    case 39:
      if (lookahead == 't') ADVANCE(21);
      END_STATE();
    case 40:
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(40)
      if (lookahead != 0) ADVANCE(82);
      END_STATE();
    case 41:
      if (eof) ADVANCE(42);
      if (lookahead == '#') ADVANCE(50);
      if (lookahead == 'd') ADVANCE(8);
      if (lookahead == 'r') ADVANCE(9);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(41)
      END_STATE();
    case 42:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(anon_sym_LF);
      if (lookahead == '\n') ADVANCE(43);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(anon_sym_definition);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(anon_sym_definition);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(anon_sym_reference);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(anon_sym_reference);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(anon_sym_relationships);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(anon_sym_relationships);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(7);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(7);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead)) ADVANCE(82);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead == '\t' ||
          (11 <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') ADVANCE(52);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(53);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(53);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(anon_sym_POUNDdocstring_COLON);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'a') ADVANCE(80);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'c') ADVANCE(60);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(62);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(63);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(77);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(47);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(70);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 62:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'f') ADVANCE(65);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'f') ADVANCE(59);
      if (lookahead == 'l') ADVANCE(55);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 64:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'h') ADVANCE(67);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 65:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(73);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 66:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(74);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 67:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(76);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 68:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(75);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 69:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(81);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 70:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(56);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 71:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(78);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 72:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(45);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 73:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(69);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 74:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'o') ADVANCE(71);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 75:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'o') ADVANCE(72);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 76:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'p') ADVANCE(79);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 77:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'r') ADVANCE(61);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 78:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 's') ADVANCE(64);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 79:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 's') ADVANCE(49);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 80:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 't') ADVANCE(66);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 81:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 't') ADVANCE(68);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    case 82:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead != 0 &&
          (lookahead < '\t' || '\r' < lookahead) &&
          lookahead != ' ') ADVANCE(82);
      END_STATE();
    default:
      return false;
  }
}

static bool ts_lex_keywords(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (lookahead == 'd') ADVANCE(1);
      if (lookahead == 'f') ADVANCE(2);
      if (lookahead == 'g') ADVANCE(3);
      if (lookahead == 'i') ADVANCE(4);
      if (lookahead == 'r') ADVANCE(5);
      if (lookahead == 't') ADVANCE(6);
      if (('\t' <= lookahead && lookahead <= '\r') ||
          lookahead == ' ') SKIP(0)
      END_STATE();
    case 1:
      if (lookahead == 'e') ADVANCE(7);
      END_STATE();
    case 2:
      if (lookahead == 'o') ADVANCE(8);
      END_STATE();
    case 3:
      if (lookahead == 'l') ADVANCE(9);
      END_STATE();
    case 4:
      if (lookahead == 'm') ADVANCE(10);
      END_STATE();
    case 5:
      if (lookahead == 'e') ADVANCE(11);
      END_STATE();
    case 6:
      if (lookahead == 'y') ADVANCE(12);
      END_STATE();
    case 7:
      if (lookahead == 'f') ADVANCE(13);
      END_STATE();
    case 8:
      if (lookahead == 'r') ADVANCE(14);
      END_STATE();
    case 9:
      if (lookahead == 'o') ADVANCE(15);
      END_STATE();
    case 10:
      if (lookahead == 'p') ADVANCE(16);
      END_STATE();
    case 11:
      if (lookahead == 'f') ADVANCE(17);
      END_STATE();
    case 12:
      if (lookahead == 'p') ADVANCE(18);
      END_STATE();
    case 13:
      if (lookahead == 'i') ADVANCE(19);
      END_STATE();
    case 14:
      if (lookahead == 'w') ADVANCE(20);
      END_STATE();
    case 15:
      if (lookahead == 'b') ADVANCE(21);
      END_STATE();
    case 16:
      if (lookahead == 'l') ADVANCE(22);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(23);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(24);
      END_STATE();
    case 19:
      if (lookahead == 'n') ADVANCE(25);
      END_STATE();
    case 20:
      if (lookahead == 'a') ADVANCE(26);
      END_STATE();
    case 21:
      if (lookahead == 'a') ADVANCE(27);
      END_STATE();
    case 22:
      if (lookahead == 'e') ADVANCE(28);
      END_STATE();
    case 23:
      if (lookahead == 'r') ADVANCE(29);
      END_STATE();
    case 24:
      if (lookahead == '_') ADVANCE(30);
      END_STATE();
    case 25:
      if (lookahead == 'e') ADVANCE(31);
      END_STATE();
    case 26:
      if (lookahead == 'r') ADVANCE(32);
      END_STATE();
    case 27:
      if (lookahead == 'l') ADVANCE(33);
      END_STATE();
    case 28:
      if (lookahead == 'm') ADVANCE(34);
      END_STATE();
    case 29:
      if (lookahead == 'e') ADVANCE(35);
      END_STATE();
    case 30:
      if (lookahead == 'd') ADVANCE(36);
      END_STATE();
    case 31:
      if (lookahead == 'd') ADVANCE(37);
      END_STATE();
    case 32:
      if (lookahead == 'd') ADVANCE(38);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_global);
      END_STATE();
    case 34:
      if (lookahead == 'e') ADVANCE(39);
      END_STATE();
    case 35:
      if (lookahead == 'n') ADVANCE(40);
      END_STATE();
    case 36:
      if (lookahead == 'e') ADVANCE(41);
      END_STATE();
    case 37:
      if (lookahead == '_') ADVANCE(42);
      END_STATE();
    case 38:
      if (lookahead == '_') ADVANCE(43);
      END_STATE();
    case 39:
      if (lookahead == 'n') ADVANCE(44);
      END_STATE();
    case 40:
      if (lookahead == 'c') ADVANCE(45);
      END_STATE();
    case 41:
      if (lookahead == 'f') ADVANCE(46);
      END_STATE();
    case 42:
      if (lookahead == 'b') ADVANCE(47);
      END_STATE();
    case 43:
      if (lookahead == 'd') ADVANCE(48);
      END_STATE();
    case 44:
      if (lookahead == 't') ADVANCE(49);
      END_STATE();
    case 45:
      if (lookahead == 'e') ADVANCE(50);
      END_STATE();
    case 46:
      if (lookahead == 'i') ADVANCE(51);
      END_STATE();
    case 47:
      if (lookahead == 'y') ADVANCE(52);
      END_STATE();
    case 48:
      if (lookahead == 'e') ADVANCE(53);
      END_STATE();
    case 49:
      if (lookahead == 's') ADVANCE(54);
      END_STATE();
    case 50:
      if (lookahead == 's') ADVANCE(55);
      END_STATE();
    case 51:
      if (lookahead == 'n') ADVANCE(56);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(anon_sym_defined_by);
      END_STATE();
    case 53:
      if (lookahead == 'f') ADVANCE(57);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(anon_sym_implements);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(anon_sym_references);
      END_STATE();
    case 56:
      if (lookahead == 'e') ADVANCE(58);
      END_STATE();
    case 57:
      if (lookahead == 'i') ADVANCE(59);
      END_STATE();
    case 58:
      if (lookahead == 's') ADVANCE(60);
      END_STATE();
    case 59:
      if (lookahead == 'n') ADVANCE(61);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(anon_sym_type_defines);
      END_STATE();
    case 61:
      if (lookahead == 'i') ADVANCE(62);
      END_STATE();
    case 62:
      if (lookahead == 't') ADVANCE(63);
      END_STATE();
    case 63:
      if (lookahead == 'i') ADVANCE(64);
      END_STATE();
    case 64:
      if (lookahead == 'o') ADVANCE(65);
      END_STATE();
    case 65:
      if (lookahead == 'n') ADVANCE(66);
      END_STATE();
    case 66:
      ACCEPT_TOKEN(anon_sym_forward_definition);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 41},
  [2] = {.lex_state = 41},
  [3] = {.lex_state = 41},
  [4] = {.lex_state = 1},
  [5] = {.lex_state = 1},
  [6] = {.lex_state = 1},
  [7] = {.lex_state = 1},
  [8] = {.lex_state = 1},
  [9] = {.lex_state = 1},
  [10] = {.lex_state = 1},
  [11] = {.lex_state = 1},
  [12] = {.lex_state = 41},
  [13] = {.lex_state = 1},
  [14] = {.lex_state = 40},
  [15] = {.lex_state = 1},
  [16] = {.lex_state = 1},
  [17] = {.lex_state = 1},
  [18] = {.lex_state = 1},
  [19] = {.lex_state = 1},
  [20] = {.lex_state = 1},
  [21] = {.lex_state = 40},
  [22] = {.lex_state = 40},
  [23] = {.lex_state = 40},
  [24] = {.lex_state = 40},
  [25] = {.lex_state = 40},
  [26] = {.lex_state = 40},
  [27] = {.lex_state = 40},
  [28] = {.lex_state = 40},
  [29] = {.lex_state = 40},
  [30] = {.lex_state = 2},
  [31] = {.lex_state = 52},
  [32] = {.lex_state = 2},
  [33] = {.lex_state = 40},
  [34] = {.lex_state = 41},
  [35] = {.lex_state = 2},
  [36] = {.lex_state = 2},
  [37] = {.lex_state = 2},
  [38] = {.lex_state = 2},
  [39] = {.lex_state = 0},
  [40] = {.lex_state = 52},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_workspace_identifier] = ACTIONS(1),
    [anon_sym_definition] = ACTIONS(1),
    [anon_sym_reference] = ACTIONS(1),
    [anon_sym_forward_definition] = ACTIONS(1),
    [anon_sym_implements] = ACTIONS(1),
    [anon_sym_type_defines] = ACTIONS(1),
    [anon_sym_references] = ACTIONS(1),
    [anon_sym_relationships] = ACTIONS(1),
    [anon_sym_defined_by] = ACTIONS(1),
    [anon_sym_POUND] = ACTIONS(1),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(1),
    [anon_sym_global] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(39),
    [sym__statement] = STATE(2),
    [sym_definition_statement] = STATE(30),
    [sym_reference_statement] = STATE(30),
    [sym_relationships_statement] = STATE(30),
    [sym_comment] = STATE(30),
    [sym_docstring] = STATE(37),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(3),
    [anon_sym_definition] = ACTIONS(5),
    [anon_sym_reference] = ACTIONS(7),
    [anon_sym_relationships] = ACTIONS(9),
    [anon_sym_POUND] = ACTIONS(11),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(13),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 9,
    ACTIONS(5), 1,
      anon_sym_definition,
    ACTIONS(7), 1,
      anon_sym_reference,
    ACTIONS(9), 1,
      anon_sym_relationships,
    ACTIONS(11), 1,
      anon_sym_POUND,
    ACTIONS(13), 1,
      anon_sym_POUNDdocstring_COLON,
    ACTIONS(15), 1,
      ts_builtin_sym_end,
    STATE(37), 1,
      sym_docstring,
    STATE(3), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(30), 4,
      sym_definition_statement,
      sym_reference_statement,
      sym_relationships_statement,
      sym_comment,
  [32] = 9,
    ACTIONS(17), 1,
      ts_builtin_sym_end,
    ACTIONS(19), 1,
      anon_sym_definition,
    ACTIONS(22), 1,
      anon_sym_reference,
    ACTIONS(25), 1,
      anon_sym_relationships,
    ACTIONS(28), 1,
      anon_sym_POUND,
    ACTIONS(31), 1,
      anon_sym_POUNDdocstring_COLON,
    STATE(37), 1,
      sym_docstring,
    STATE(3), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(30), 4,
      sym_definition_statement,
      sym_reference_statement,
      sym_relationships_statement,
      sym_comment,
  [64] = 6,
    ACTIONS(34), 1,
      anon_sym_LF,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(42), 1,
      anon_sym_defined_by,
    STATE(5), 7,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      sym__all_relations,
      sym_defined_by_relation,
      aux_sym_relationships_statement_repeat1,
  [89] = 6,
    ACTIONS(44), 1,
      anon_sym_LF,
    ACTIONS(46), 1,
      anon_sym_implements,
    ACTIONS(49), 1,
      anon_sym_type_defines,
    ACTIONS(52), 1,
      anon_sym_references,
    ACTIONS(55), 1,
      anon_sym_defined_by,
    STATE(5), 7,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      sym__all_relations,
      sym_defined_by_relation,
      aux_sym_relationships_statement_repeat1,
  [114] = 6,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(42), 1,
      anon_sym_defined_by,
    ACTIONS(58), 1,
      anon_sym_LF,
    STATE(4), 7,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      sym__all_relations,
      sym_defined_by_relation,
      aux_sym_relationships_statement_repeat1,
  [139] = 5,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(60), 1,
      anon_sym_LF,
    STATE(11), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [159] = 5,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(62), 1,
      anon_sym_LF,
    STATE(10), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [179] = 5,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(64), 1,
      anon_sym_LF,
    STATE(8), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [199] = 5,
    ACTIONS(66), 1,
      anon_sym_LF,
    ACTIONS(68), 1,
      anon_sym_implements,
    ACTIONS(71), 1,
      anon_sym_type_defines,
    ACTIONS(74), 1,
      anon_sym_references,
    STATE(10), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [219] = 5,
    ACTIONS(36), 1,
      anon_sym_implements,
    ACTIONS(38), 1,
      anon_sym_type_defines,
    ACTIONS(40), 1,
      anon_sym_references,
    ACTIONS(77), 1,
      anon_sym_LF,
    STATE(10), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [239] = 2,
    ACTIONS(81), 1,
      anon_sym_POUND,
    ACTIONS(79), 5,
      ts_builtin_sym_end,
      anon_sym_definition,
      anon_sym_reference,
      anon_sym_relationships,
      anon_sym_POUNDdocstring_COLON,
  [250] = 2,
    ACTIONS(83), 1,
      anon_sym_LF,
    ACTIONS(85), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [260] = 5,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(89), 1,
      anon_sym_forward_definition,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(38), 1,
      sym_identifier,
  [276] = 2,
    ACTIONS(93), 1,
      anon_sym_LF,
    ACTIONS(95), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [286] = 2,
    ACTIONS(97), 1,
      anon_sym_LF,
    ACTIONS(99), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [296] = 2,
    ACTIONS(101), 1,
      anon_sym_LF,
    ACTIONS(103), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [306] = 2,
    ACTIONS(105), 1,
      anon_sym_LF,
    ACTIONS(107), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [316] = 2,
    ACTIONS(109), 1,
      anon_sym_LF,
    ACTIONS(111), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [326] = 2,
    ACTIONS(113), 1,
      anon_sym_LF,
    ACTIONS(115), 4,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
      anon_sym_defined_by,
  [336] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(7), 1,
      sym_identifier,
    STATE(16), 1,
      sym_global_identifier,
  [349] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(6), 1,
      sym_identifier,
    STATE(16), 1,
      sym_global_identifier,
  [362] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(19), 1,
      sym_identifier,
  [375] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(18), 1,
      sym_identifier,
  [388] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(17), 1,
      sym_identifier,
  [401] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(32), 1,
      sym_identifier,
  [414] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(9), 1,
      sym_identifier,
    STATE(16), 1,
      sym_global_identifier,
  [427] = 4,
    ACTIONS(87), 1,
      sym_workspace_identifier,
    ACTIONS(91), 1,
      anon_sym_global,
    STATE(16), 1,
      sym_global_identifier,
    STATE(20), 1,
      sym_identifier,
  [440] = 1,
    ACTIONS(117), 1,
      sym_workspace_identifier,
  [444] = 1,
    ACTIONS(119), 1,
      anon_sym_LF,
  [448] = 1,
    ACTIONS(121), 1,
      aux_sym_comment_token1,
  [452] = 1,
    ACTIONS(123), 1,
      anon_sym_LF,
  [456] = 1,
    ACTIONS(125), 1,
      sym_workspace_identifier,
  [460] = 1,
    ACTIONS(127), 1,
      anon_sym_definition,
  [464] = 1,
    ACTIONS(129), 1,
      anon_sym_LF,
  [468] = 1,
    ACTIONS(131), 1,
      anon_sym_LF,
  [472] = 1,
    ACTIONS(133), 1,
      anon_sym_LF,
  [476] = 1,
    ACTIONS(135), 1,
      anon_sym_LF,
  [480] = 1,
    ACTIONS(137), 1,
      ts_builtin_sym_end,
  [484] = 1,
    ACTIONS(139), 1,
      aux_sym_comment_token1,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 32,
  [SMALL_STATE(4)] = 64,
  [SMALL_STATE(5)] = 89,
  [SMALL_STATE(6)] = 114,
  [SMALL_STATE(7)] = 139,
  [SMALL_STATE(8)] = 159,
  [SMALL_STATE(9)] = 179,
  [SMALL_STATE(10)] = 199,
  [SMALL_STATE(11)] = 219,
  [SMALL_STATE(12)] = 239,
  [SMALL_STATE(13)] = 250,
  [SMALL_STATE(14)] = 260,
  [SMALL_STATE(15)] = 276,
  [SMALL_STATE(16)] = 286,
  [SMALL_STATE(17)] = 296,
  [SMALL_STATE(18)] = 306,
  [SMALL_STATE(19)] = 316,
  [SMALL_STATE(20)] = 326,
  [SMALL_STATE(21)] = 336,
  [SMALL_STATE(22)] = 349,
  [SMALL_STATE(23)] = 362,
  [SMALL_STATE(24)] = 375,
  [SMALL_STATE(25)] = 388,
  [SMALL_STATE(26)] = 401,
  [SMALL_STATE(27)] = 414,
  [SMALL_STATE(28)] = 427,
  [SMALL_STATE(29)] = 440,
  [SMALL_STATE(30)] = 444,
  [SMALL_STATE(31)] = 448,
  [SMALL_STATE(32)] = 452,
  [SMALL_STATE(33)] = 456,
  [SMALL_STATE(34)] = 460,
  [SMALL_STATE(35)] = 464,
  [SMALL_STATE(36)] = 468,
  [SMALL_STATE(37)] = 472,
  [SMALL_STATE(38)] = 476,
  [SMALL_STATE(39)] = 480,
  [SMALL_STATE(40)] = 484,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [11] = {.entry = {.count = 1, .reusable = false}}, SHIFT(31),
  [13] = {.entry = {.count = 1, .reusable = true}}, SHIFT(40),
  [15] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1),
  [17] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2),
  [19] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(21),
  [22] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(14),
  [25] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(22),
  [28] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(31),
  [31] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(40),
  [34] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_relationships_statement, 3, .production_id = 4),
  [36] = {.entry = {.count = 1, .reusable = false}}, SHIFT(28),
  [38] = {.entry = {.count = 1, .reusable = false}}, SHIFT(23),
  [40] = {.entry = {.count = 1, .reusable = false}}, SHIFT(24),
  [42] = {.entry = {.count = 1, .reusable = false}}, SHIFT(25),
  [44] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_relationships_statement_repeat1, 2),
  [46] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_relationships_statement_repeat1, 2), SHIFT_REPEAT(28),
  [49] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_relationships_statement_repeat1, 2), SHIFT_REPEAT(23),
  [52] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_relationships_statement_repeat1, 2), SHIFT_REPEAT(24),
  [55] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_relationships_statement_repeat1, 2), SHIFT_REPEAT(25),
  [58] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_relationships_statement, 2, .production_id = 2),
  [60] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 2, .production_id = 2),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 5, .production_id = 8),
  [64] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 4, .production_id = 7),
  [66] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_definition_statement_repeat1, 2),
  [68] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(28),
  [71] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(23),
  [74] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(24),
  [77] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 3, .production_id = 4),
  [79] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym__statement, 2),
  [81] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym__statement, 2),
  [83] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_global_identifier, 3, .production_id = 6),
  [85] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_global_identifier, 3, .production_id = 6),
  [87] = {.entry = {.count = 1, .reusable = false}}, SHIFT(15),
  [89] = {.entry = {.count = 1, .reusable = false}}, SHIFT(26),
  [91] = {.entry = {.count = 1, .reusable = false}}, SHIFT(33),
  [93] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier, 1, .production_id = 1),
  [95] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_identifier, 1, .production_id = 1),
  [97] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier, 1, .production_id = 3),
  [99] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_identifier, 1, .production_id = 3),
  [101] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defined_by_relation, 2, .production_id = 2),
  [103] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_defined_by_relation, 2, .production_id = 2),
  [105] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_references_relation, 2, .production_id = 2),
  [107] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_references_relation, 2, .production_id = 2),
  [109] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_type_definition_relation, 2, .production_id = 2),
  [111] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_type_definition_relation, 2, .production_id = 2),
  [113] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_implementation_relation, 2, .production_id = 2),
  [115] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_implementation_relation, 2, .production_id = 2),
  [117] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [119] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [121] = {.entry = {.count = 1, .reusable = true}}, SHIFT(36),
  [123] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_reference_statement, 3, .production_id = 5),
  [125] = {.entry = {.count = 1, .reusable = true}}, SHIFT(29),
  [127] = {.entry = {.count = 1, .reusable = true}}, SHIFT(27),
  [129] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_docstring, 2),
  [131] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_comment, 2),
  [133] = {.entry = {.count = 1, .reusable = true}}, SHIFT(34),
  [135] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_reference_statement, 2, .production_id = 2),
  [137] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [139] = {.entry = {.count = 1, .reusable = true}}, SHIFT(35),
};

#ifdef __cplusplus
extern "C" {
#endif
#ifdef _WIN32
#define extern __declspec(dllexport)
#endif

extern const TSLanguage *tree_sitter_reprolang(void) {
  static const TSLanguage language = {
    .version = LANGUAGE_VERSION,
    .symbol_count = SYMBOL_COUNT,
    .alias_count = ALIAS_COUNT,
    .token_count = TOKEN_COUNT,
    .external_token_count = EXTERNAL_TOKEN_COUNT,
    .state_count = STATE_COUNT,
    .large_state_count = LARGE_STATE_COUNT,
    .production_id_count = PRODUCTION_ID_COUNT,
    .field_count = FIELD_COUNT,
    .max_alias_sequence_length = MAX_ALIAS_SEQUENCE_LENGTH,
    .parse_table = &ts_parse_table[0][0],
    .small_parse_table = ts_small_parse_table,
    .small_parse_table_map = ts_small_parse_table_map,
    .parse_actions = ts_parse_actions,
    .symbol_names = ts_symbol_names,
    .field_names = ts_field_names,
    .field_map_slices = ts_field_map_slices,
    .field_map_entries = ts_field_map_entries,
    .symbol_metadata = ts_symbol_metadata,
    .public_symbol_map = ts_symbol_map,
    .alias_map = ts_non_terminal_alias_map,
    .alias_sequences = &ts_alias_sequences[0][0],
    .lex_modes = ts_lex_modes,
    .lex_fn = ts_lex,
    .keyword_lex_fn = ts_lex_keywords,
    .keyword_capture_token = sym_workspace_identifier,
  };
  return &language;
}
#ifdef __cplusplus
}
#endif
