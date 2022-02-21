#include <tree_sitter/parser.h>

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 13
#define STATE_COUNT 33
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 26
#define ALIAS_COUNT 0
#define TOKEN_COUNT 12
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 7
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 8

enum {
  sym_workspace_identifier = 1,
  anon_sym_LF = 2,
  anon_sym_definition = 3,
  anon_sym_reference = 4,
  anon_sym_implements = 5,
  anon_sym_type_defines = 6,
  anon_sym_references = 7,
  anon_sym_POUND = 8,
  aux_sym_comment_token1 = 9,
  anon_sym_POUNDdocstring_COLON = 10,
  anon_sym_global = 11,
  sym_source_file = 12,
  sym__statement = 13,
  sym_definition_statement = 14,
  sym_reference_statement = 15,
  sym__definition_relations = 16,
  sym_implementation_relation = 17,
  sym_type_definition_relation = 18,
  sym_references_relation = 19,
  sym_comment = 20,
  sym_docstring = 21,
  sym_identifier = 22,
  sym_global_identifier = 23,
  aux_sym_source_file_repeat1 = 24,
  aux_sym_definition_statement_repeat1 = 25,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_workspace_identifier] = "workspace_identifier",
  [anon_sym_LF] = "\n",
  [anon_sym_definition] = "definition",
  [anon_sym_reference] = "reference",
  [anon_sym_implements] = "implements",
  [anon_sym_type_defines] = "type_defines",
  [anon_sym_references] = "references",
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
  [sym_comment] = "comment",
  [sym_docstring] = "docstring",
  [sym_identifier] = "identifier",
  [sym_global_identifier] = "global_identifier",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_definition_statement_repeat1] = "definition_statement_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_workspace_identifier] = sym_workspace_identifier,
  [anon_sym_LF] = anon_sym_LF,
  [anon_sym_definition] = anon_sym_definition,
  [anon_sym_reference] = anon_sym_reference,
  [anon_sym_implements] = anon_sym_implements,
  [anon_sym_type_defines] = anon_sym_type_defines,
  [anon_sym_references] = anon_sym_references,
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
  [sym_comment] = sym_comment,
  [sym_docstring] = sym_docstring,
  [sym_identifier] = sym_identifier,
  [sym_global_identifier] = sym_global_identifier,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_definition_statement_repeat1] = aux_sym_definition_statement_repeat1,
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
};

enum {
  field_descriptors = 1,
  field_docstring = 2,
  field_global = 3,
  field_name = 4,
  field_project_name = 5,
  field_roles = 6,
  field_workspace = 7,
};

static const char * const ts_field_names[] = {
  [0] = NULL,
  [field_descriptors] = "descriptors",
  [field_docstring] = "docstring",
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
  [6] = {.index = 7, .length = 3},
  [7] = {.index = 10, .length = 4},
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
    {field_descriptors, 2},
    {field_project_name, 1},
  [7] =
    {field_docstring, 0},
    {field_docstring, 1},
    {field_name, 3},
  [10] =
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

static const TSStateId ts_primary_state_ids[STATE_COUNT] = {
  [0] = 0,
  [1] = 1,
  [2] = 2,
  [3] = 3,
  [4] = 4,
  [5] = 5,
  [6] = 6,
  [7] = 7,
  [8] = 8,
  [9] = 9,
  [10] = 10,
  [11] = 11,
  [12] = 12,
  [13] = 13,
  [14] = 14,
  [15] = 15,
  [16] = 16,
  [17] = 17,
  [18] = 18,
  [19] = 19,
  [20] = 20,
  [21] = 21,
  [22] = 22,
  [23] = 23,
  [24] = 24,
  [25] = 25,
  [26] = 26,
  [27] = 27,
  [28] = 28,
  [29] = 29,
  [30] = 30,
  [31] = 31,
  [32] = 32,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(32);
      if (lookahead == '#') ADVANCE(39);
      if (lookahead == 'd') ADVANCE(44);
      if (lookahead == 'r') ADVANCE(47);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 1:
      if (lookahead == '\n') ADVANCE(33);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(1)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 2:
      if (lookahead == '\n') ADVANCE(33);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(2)
      END_STATE();
    case 3:
      if (lookahead == ':') ADVANCE(42);
      END_STATE();
    case 4:
      if (lookahead == 'c') ADVANCE(27);
      END_STATE();
    case 5:
      if (lookahead == 'c') ADVANCE(8);
      END_STATE();
    case 6:
      if (lookahead == 'd') ADVANCE(23);
      END_STATE();
    case 7:
      if (lookahead == 'e') ADVANCE(13);
      END_STATE();
    case 8:
      if (lookahead == 'e') ADVANCE(36);
      END_STATE();
    case 9:
      if (lookahead == 'e') ADVANCE(26);
      END_STATE();
    case 10:
      if (lookahead == 'e') ADVANCE(12);
      END_STATE();
    case 11:
      if (lookahead == 'e') ADVANCE(21);
      END_STATE();
    case 12:
      if (lookahead == 'f') ADVANCE(9);
      END_STATE();
    case 13:
      if (lookahead == 'f') ADVANCE(18);
      END_STATE();
    case 14:
      if (lookahead == 'g') ADVANCE(3);
      END_STATE();
    case 15:
      if (lookahead == 'i') ADVANCE(19);
      END_STATE();
    case 16:
      if (lookahead == 'i') ADVANCE(24);
      END_STATE();
    case 17:
      if (lookahead == 'i') ADVANCE(29);
      END_STATE();
    case 18:
      if (lookahead == 'i') ADVANCE(22);
      END_STATE();
    case 19:
      if (lookahead == 'n') ADVANCE(14);
      END_STATE();
    case 20:
      if (lookahead == 'n') ADVANCE(34);
      END_STATE();
    case 21:
      if (lookahead == 'n') ADVANCE(5);
      END_STATE();
    case 22:
      if (lookahead == 'n') ADVANCE(17);
      END_STATE();
    case 23:
      if (lookahead == 'o') ADVANCE(4);
      END_STATE();
    case 24:
      if (lookahead == 'o') ADVANCE(20);
      END_STATE();
    case 25:
      if (lookahead == 'r') ADVANCE(15);
      END_STATE();
    case 26:
      if (lookahead == 'r') ADVANCE(11);
      END_STATE();
    case 27:
      if (lookahead == 's') ADVANCE(28);
      END_STATE();
    case 28:
      if (lookahead == 't') ADVANCE(25);
      END_STATE();
    case 29:
      if (lookahead == 't') ADVANCE(16);
      END_STATE();
    case 30:
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(30)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 31:
      if (eof) ADVANCE(32);
      if (lookahead == '#') ADVANCE(38);
      if (lookahead == 'd') ADVANCE(7);
      if (lookahead == 'r') ADVANCE(10);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(31)
      END_STATE();
    case 32:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_LF);
      if (lookahead == '\n') ADVANCE(33);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(anon_sym_definition);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(anon_sym_definition);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_reference);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(anon_sym_reference);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(6);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(6);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r') ADVANCE(60);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(40);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(41);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(41);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(anon_sym_POUNDdocstring_COLON);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'c') ADVANCE(46);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(49);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(58);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(37);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(50);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'e') ADVANCE(54);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'f') ADVANCE(51);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'f') ADVANCE(45);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(56);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(59);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'i') ADVANCE(57);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(43);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(35);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'n') ADVANCE(52);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'o') ADVANCE(55);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 'r') ADVANCE(48);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead == 't') ADVANCE(53);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym_workspace_identifier);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
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
      if (lookahead == 'g') ADVANCE(1);
      if (lookahead == 'i') ADVANCE(2);
      if (lookahead == 'r') ADVANCE(3);
      if (lookahead == 't') ADVANCE(4);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0)
      END_STATE();
    case 1:
      if (lookahead == 'l') ADVANCE(5);
      END_STATE();
    case 2:
      if (lookahead == 'm') ADVANCE(6);
      END_STATE();
    case 3:
      if (lookahead == 'e') ADVANCE(7);
      END_STATE();
    case 4:
      if (lookahead == 'y') ADVANCE(8);
      END_STATE();
    case 5:
      if (lookahead == 'o') ADVANCE(9);
      END_STATE();
    case 6:
      if (lookahead == 'p') ADVANCE(10);
      END_STATE();
    case 7:
      if (lookahead == 'f') ADVANCE(11);
      END_STATE();
    case 8:
      if (lookahead == 'p') ADVANCE(12);
      END_STATE();
    case 9:
      if (lookahead == 'b') ADVANCE(13);
      END_STATE();
    case 10:
      if (lookahead == 'l') ADVANCE(14);
      END_STATE();
    case 11:
      if (lookahead == 'e') ADVANCE(15);
      END_STATE();
    case 12:
      if (lookahead == 'e') ADVANCE(16);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(17);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(18);
      END_STATE();
    case 15:
      if (lookahead == 'r') ADVANCE(19);
      END_STATE();
    case 16:
      if (lookahead == '_') ADVANCE(20);
      END_STATE();
    case 17:
      if (lookahead == 'l') ADVANCE(21);
      END_STATE();
    case 18:
      if (lookahead == 'm') ADVANCE(22);
      END_STATE();
    case 19:
      if (lookahead == 'e') ADVANCE(23);
      END_STATE();
    case 20:
      if (lookahead == 'd') ADVANCE(24);
      END_STATE();
    case 21:
      ACCEPT_TOKEN(anon_sym_global);
      END_STATE();
    case 22:
      if (lookahead == 'e') ADVANCE(25);
      END_STATE();
    case 23:
      if (lookahead == 'n') ADVANCE(26);
      END_STATE();
    case 24:
      if (lookahead == 'e') ADVANCE(27);
      END_STATE();
    case 25:
      if (lookahead == 'n') ADVANCE(28);
      END_STATE();
    case 26:
      if (lookahead == 'c') ADVANCE(29);
      END_STATE();
    case 27:
      if (lookahead == 'f') ADVANCE(30);
      END_STATE();
    case 28:
      if (lookahead == 't') ADVANCE(31);
      END_STATE();
    case 29:
      if (lookahead == 'e') ADVANCE(32);
      END_STATE();
    case 30:
      if (lookahead == 'i') ADVANCE(33);
      END_STATE();
    case 31:
      if (lookahead == 's') ADVANCE(34);
      END_STATE();
    case 32:
      if (lookahead == 's') ADVANCE(35);
      END_STATE();
    case 33:
      if (lookahead == 'n') ADVANCE(36);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(anon_sym_implements);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(anon_sym_references);
      END_STATE();
    case 36:
      if (lookahead == 'e') ADVANCE(37);
      END_STATE();
    case 37:
      if (lookahead == 's') ADVANCE(38);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(anon_sym_type_defines);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 31},
  [2] = {.lex_state = 31},
  [3] = {.lex_state = 31},
  [4] = {.lex_state = 1},
  [5] = {.lex_state = 1},
  [6] = {.lex_state = 1},
  [7] = {.lex_state = 1},
  [8] = {.lex_state = 1},
  [9] = {.lex_state = 31},
  [10] = {.lex_state = 1},
  [11] = {.lex_state = 1},
  [12] = {.lex_state = 30},
  [13] = {.lex_state = 1},
  [14] = {.lex_state = 1},
  [15] = {.lex_state = 1},
  [16] = {.lex_state = 30},
  [17] = {.lex_state = 1},
  [18] = {.lex_state = 30},
  [19] = {.lex_state = 30},
  [20] = {.lex_state = 30},
  [21] = {.lex_state = 30},
  [22] = {.lex_state = 2},
  [23] = {.lex_state = 31},
  [24] = {.lex_state = 30},
  [25] = {.lex_state = 40},
  [26] = {.lex_state = 2},
  [27] = {.lex_state = 2},
  [28] = {.lex_state = 2},
  [29] = {.lex_state = 30},
  [30] = {.lex_state = 2},
  [31] = {.lex_state = 0},
  [32] = {.lex_state = 40},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_workspace_identifier] = ACTIONS(1),
    [anon_sym_definition] = ACTIONS(1),
    [anon_sym_reference] = ACTIONS(1),
    [anon_sym_implements] = ACTIONS(1),
    [anon_sym_type_defines] = ACTIONS(1),
    [anon_sym_references] = ACTIONS(1),
    [anon_sym_POUND] = ACTIONS(1),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(1),
    [anon_sym_global] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(31),
    [sym__statement] = STATE(3),
    [sym_definition_statement] = STATE(30),
    [sym_reference_statement] = STATE(30),
    [sym_comment] = STATE(30),
    [sym_docstring] = STATE(22),
    [aux_sym_source_file_repeat1] = STATE(3),
    [ts_builtin_sym_end] = ACTIONS(3),
    [anon_sym_definition] = ACTIONS(5),
    [anon_sym_reference] = ACTIONS(7),
    [anon_sym_POUND] = ACTIONS(9),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(11),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 8,
    ACTIONS(13), 1,
      ts_builtin_sym_end,
    ACTIONS(15), 1,
      anon_sym_definition,
    ACTIONS(18), 1,
      anon_sym_reference,
    ACTIONS(21), 1,
      anon_sym_POUND,
    ACTIONS(24), 1,
      anon_sym_POUNDdocstring_COLON,
    STATE(22), 1,
      sym_docstring,
    STATE(2), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(30), 3,
      sym_definition_statement,
      sym_reference_statement,
      sym_comment,
  [28] = 8,
    ACTIONS(5), 1,
      anon_sym_definition,
    ACTIONS(7), 1,
      anon_sym_reference,
    ACTIONS(9), 1,
      anon_sym_POUND,
    ACTIONS(11), 1,
      anon_sym_POUNDdocstring_COLON,
    ACTIONS(27), 1,
      ts_builtin_sym_end,
    STATE(22), 1,
      sym_docstring,
    STATE(2), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(30), 3,
      sym_definition_statement,
      sym_reference_statement,
      sym_comment,
  [56] = 5,
    ACTIONS(29), 1,
      anon_sym_LF,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_type_defines,
    ACTIONS(35), 1,
      anon_sym_references,
    STATE(7), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [76] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_type_defines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(37), 1,
      anon_sym_LF,
    STATE(7), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [96] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_type_defines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(39), 1,
      anon_sym_LF,
    STATE(5), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [116] = 5,
    ACTIONS(41), 1,
      anon_sym_LF,
    ACTIONS(43), 1,
      anon_sym_implements,
    ACTIONS(46), 1,
      anon_sym_type_defines,
    ACTIONS(49), 1,
      anon_sym_references,
    STATE(7), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [136] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_type_defines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(52), 1,
      anon_sym_LF,
    STATE(4), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [156] = 2,
    ACTIONS(56), 1,
      anon_sym_POUND,
    ACTIONS(54), 4,
      ts_builtin_sym_end,
      anon_sym_definition,
      anon_sym_reference,
      anon_sym_POUNDdocstring_COLON,
  [166] = 2,
    ACTIONS(58), 1,
      anon_sym_LF,
    ACTIONS(60), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [175] = 2,
    ACTIONS(62), 1,
      anon_sym_LF,
    ACTIONS(64), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [184] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(8), 1,
      sym_identifier,
    STATE(13), 1,
      sym_global_identifier,
  [197] = 2,
    ACTIONS(70), 1,
      anon_sym_LF,
    ACTIONS(72), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [206] = 2,
    ACTIONS(74), 1,
      anon_sym_LF,
    ACTIONS(76), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [215] = 2,
    ACTIONS(78), 1,
      anon_sym_LF,
    ACTIONS(80), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [224] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(6), 1,
      sym_identifier,
    STATE(13), 1,
      sym_global_identifier,
  [237] = 2,
    ACTIONS(82), 1,
      anon_sym_LF,
    ACTIONS(84), 3,
      anon_sym_implements,
      anon_sym_type_defines,
      anon_sym_references,
  [246] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(13), 1,
      sym_global_identifier,
    STATE(28), 1,
      sym_identifier,
  [259] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(13), 1,
      sym_global_identifier,
    STATE(14), 1,
      sym_identifier,
  [272] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(10), 1,
      sym_identifier,
    STATE(13), 1,
      sym_global_identifier,
  [285] = 4,
    ACTIONS(66), 1,
      sym_workspace_identifier,
    ACTIONS(68), 1,
      anon_sym_global,
    STATE(13), 1,
      sym_global_identifier,
    STATE(17), 1,
      sym_identifier,
  [298] = 1,
    ACTIONS(86), 1,
      anon_sym_LF,
  [302] = 1,
    ACTIONS(88), 1,
      anon_sym_definition,
  [306] = 1,
    ACTIONS(90), 1,
      sym_workspace_identifier,
  [310] = 1,
    ACTIONS(92), 1,
      aux_sym_comment_token1,
  [314] = 1,
    ACTIONS(94), 1,
      anon_sym_LF,
  [318] = 1,
    ACTIONS(96), 1,
      anon_sym_LF,
  [322] = 1,
    ACTIONS(98), 1,
      anon_sym_LF,
  [326] = 1,
    ACTIONS(100), 1,
      sym_workspace_identifier,
  [330] = 1,
    ACTIONS(102), 1,
      anon_sym_LF,
  [334] = 1,
    ACTIONS(104), 1,
      ts_builtin_sym_end,
  [338] = 1,
    ACTIONS(106), 1,
      aux_sym_comment_token1,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 28,
  [SMALL_STATE(4)] = 56,
  [SMALL_STATE(5)] = 76,
  [SMALL_STATE(6)] = 96,
  [SMALL_STATE(7)] = 116,
  [SMALL_STATE(8)] = 136,
  [SMALL_STATE(9)] = 156,
  [SMALL_STATE(10)] = 166,
  [SMALL_STATE(11)] = 175,
  [SMALL_STATE(12)] = 184,
  [SMALL_STATE(13)] = 197,
  [SMALL_STATE(14)] = 206,
  [SMALL_STATE(15)] = 215,
  [SMALL_STATE(16)] = 224,
  [SMALL_STATE(17)] = 237,
  [SMALL_STATE(18)] = 246,
  [SMALL_STATE(19)] = 259,
  [SMALL_STATE(20)] = 272,
  [SMALL_STATE(21)] = 285,
  [SMALL_STATE(22)] = 298,
  [SMALL_STATE(23)] = 302,
  [SMALL_STATE(24)] = 306,
  [SMALL_STATE(25)] = 310,
  [SMALL_STATE(26)] = 314,
  [SMALL_STATE(27)] = 318,
  [SMALL_STATE(28)] = 322,
  [SMALL_STATE(29)] = 326,
  [SMALL_STATE(30)] = 330,
  [SMALL_STATE(31)] = 334,
  [SMALL_STATE(32)] = 338,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [9] = {.entry = {.count = 1, .reusable = false}}, SHIFT(25),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(32),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2),
  [15] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(12),
  [18] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(18),
  [21] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(25),
  [24] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(32),
  [27] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1),
  [29] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 3, .production_id = 4),
  [31] = {.entry = {.count = 1, .reusable = false}}, SHIFT(19),
  [33] = {.entry = {.count = 1, .reusable = false}}, SHIFT(20),
  [35] = {.entry = {.count = 1, .reusable = false}}, SHIFT(21),
  [37] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 5, .production_id = 7),
  [39] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 4, .production_id = 6),
  [41] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_definition_statement_repeat1, 2),
  [43] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(19),
  [46] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(20),
  [49] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(21),
  [52] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 2, .production_id = 2),
  [54] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym__statement, 2),
  [56] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym__statement, 2),
  [58] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_type_definition_relation, 2, .production_id = 2),
  [60] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_type_definition_relation, 2, .production_id = 2),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier, 1, .production_id = 1),
  [64] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_identifier, 1, .production_id = 1),
  [66] = {.entry = {.count = 1, .reusable = false}}, SHIFT(11),
  [68] = {.entry = {.count = 1, .reusable = false}}, SHIFT(29),
  [70] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier, 1, .production_id = 3),
  [72] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_identifier, 1, .production_id = 3),
  [74] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_implementation_relation, 2, .production_id = 2),
  [76] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_implementation_relation, 2, .production_id = 2),
  [78] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_global_identifier, 3, .production_id = 5),
  [80] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_global_identifier, 3, .production_id = 5),
  [82] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_references_relation, 2, .production_id = 2),
  [84] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_references_relation, 2, .production_id = 2),
  [86] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [88] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [90] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [92] = {.entry = {.count = 1, .reusable = true}}, SHIFT(27),
  [94] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_docstring, 2),
  [96] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_comment, 2),
  [98] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_reference_statement, 2, .production_id = 2),
  [100] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [104] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [106] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
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
