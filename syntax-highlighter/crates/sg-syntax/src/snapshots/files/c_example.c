#include "runtime/parser.h"
#include <assert.h>
#include <stdio.h>
#include <limits.h>
#include <stdbool.h>
#include "tree_sitter/runtime.h"
#include "runtime/tree.h"
#include "runtime/lexer.h"
#include "runtime/length.h"
#include "runtime/array.h"
#include "runtime/language.h"
#include "runtime/alloc.h"
#include "runtime/reduce_action.h"
#include "runtime/error_costs.h"

#define LOG(...)                                                           \
  if (self->lexer.logger.log) {                                            \
    snprintf(self->lexer.debug_buffer, TS_DEBUG_BUFFER_SIZE, __VA_ARGS__); \
    self->lexer.logger.log(self->lexer.logger.payload, TSLogTypeParse,     \
                           self->lexer.debug_buffer);                      \
  }                                                                        \
  if (self->print_debugging_graphs) {                                      \
    fprintf(stderr, "graph {\nlabel=\"");                                  \
    fprintf(stderr, __VA_ARGS__);                                          \
    fprintf(stderr, "\"\n}\n\n");                                          \
  }


#define SYM_NAME(symbol) ts_language_symbol_name(self->language, symbol)

typedef struct {
  Parser *parser;
  TSSymbol lookahead_symbol;
  TreeArray *trees_above_error;
  uint32_t tree_count_above_error;
  bool found_repair;
  ReduceAction best_repair;
  TSStateId best_repair_next_state;
  uint32_t best_repair_skip_count;
} ErrorRepairSession;

static void parser__push(Parser *self, StackVersion version, Tree *tree,
                         TSStateId state) {
  ts_stack_push(self->stack, version, tree, false, state);
  ts_tree_release(tree);
}

static bool parser__breakdown_top_of_stack(Parser *self, StackVersion version) {
  bool did_break_down = false;
  bool pending = false;

  do {
    StackPopResult pop = ts_stack_pop_pending(self->stack, version);
    if (!pop.slices.size)
      break;

    did_break_down = true;
    pending = false;
    for (uint32_t i = 0; i < pop.slices.size; i++) {
      StackSlice slice = pop.slices.contents[i];
      TSStateId state = ts_stack_top_state(self->stack, slice.version);
      Tree *parent = *array_front(&slice.trees);

      for (uint32_t j = 0; j < parent->child_count; j++) {
        Tree *child = parent->children[j];
        pending = child->child_count > 0;

        if (child->symbol == ts_builtin_sym_error) {
          state = ERROR_STATE;
        } else if (!child->extra) {
          state = ts_language_next_state(self->language, state, child->symbol);
        }

        ts_stack_push(self->stack, slice.version, child, pending, state);
      }

      for (uint32_t j = 1; j < slice.trees.size; j++) {
        Tree *tree = slice.trees.contents[j];
        parser__push(self, slice.version, tree, state);
      }

      LOG("breakdown_top_of_stack tree:%s", SYM_NAME(parent->symbol));
      LOG_STACK();

      ts_stack_decrease_push_count(self->stack, slice.version,
                                   parent->child_count + 1);
      ts_tree_release(parent);
      array_delete(&slice.trees);
    }
  } while (pending);

  return did_break_down;
}

static bool parser__breakdown_lookahead(Parser *self, Tree **lookahead,
                                        TSStateId state,
                                        ReusableNode *reusable_node) {
  bool result = false;
  while (reusable_node->tree->child_count > 0 &&
         (self->is_split || reusable_node->tree->parse_state != state ||
          reusable_node->tree->fragile_left ||
          reusable_node->tree->fragile_right)) {
    LOG("state_mismatch sym:%s", SYM_NAME(reusable_node->tree->symbol));
    reusable_node_breakdown(reusable_node);
    result = true;
  }

  if (result) {
    ts_tree_release(*lookahead);
    ts_tree_retain(*lookahead = reusable_node->tree);
  }

  return result;
}

static inline bool ts_lex_mode_eq(TSLexMode self, TSLexMode other) {
  return self.lex_state == other.lex_state &&
    self.external_lex_state == other.external_lex_state;
}

static bool parser__can_reuse(Parser *self, TSStateId state, Tree *tree,
                              TableEntry *table_entry) {
  TSLexMode current_lex_mode = self->language->lex_modes[state];
  if (ts_lex_mode_eq(tree->first_leaf.lex_mode, current_lex_mode))
    return true;
  if (current_lex_mode.external_lex_state != 0)
    return false;
  if (tree->size.bytes == 0)
    return false;
  if (!table_entry->is_reusable)
    return false;
  if (!table_entry->depends_on_lookahead)
    return true;
  return tree->child_count > 1 && tree->error_cost == 0;
}

typedef int CondenseResult;
static int CondenseResultMadeChange = 1;
static int CondenseResultAllVersionsHadError = 2;

static CondenseResult parser__condense_stack(Parser *self) {
  CondenseResult result = 0;
  bool has_version_without_errors = false;

  for (StackVersion i = 0; i < ts_stack_version_count(self->stack); i++) {
    if (ts_stack_is_halted(self->stack, i)) {
      ts_stack_remove_version(self->stack, i);
      result |= CondenseResultMadeChange;
      i--;
      continue;
    }

    ErrorStatus error_status = ts_stack_error_status(self->stack, i);
    if (error_status.count == 0) has_version_without_errors = true;

    for (StackVersion j = 0; j < i; j++) {
      if (ts_stack_merge(self->stack, j, i)) {
        result |= CondenseResultMadeChange;
        i--;
        break;
      }

      switch (error_status_compare(error_status,
                                   ts_stack_error_status(self->stack, j))) {
        case -1:
          ts_stack_remove_version(self->stack, j);
          result |= CondenseResultMadeChange;
          i--;
          j--;
          break;
        case 1:
          ts_stack_remove_version(self->stack, i);
          result |= CondenseResultMadeChange;
          i--;
          break;
      }
    }
  }

  if (!has_version_without_errors && ts_stack_version_count(self->stack) > 0) {
    result |= CondenseResultAllVersionsHadError;
  }

  return result;
}

static void parser__restore_external_scanner(Parser *self, StackVersion version) {
  const TSExternalTokenState *state = ts_stack_external_token_state(self->stack, version);
  if (self->lexer.last_external_token_state != state) {
    LOG("restore_external_scanner");
    self->lexer.last_external_token_state = state;
    if (state) {
      self->language->external_scanner.deserialize(
        self->external_scanner_payload,
        *state
      );
    } else {
      self->language->external_scanner.reset(self->external_scanner_payload);
    }
  }
}

