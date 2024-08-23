#include "api.h"
#include "./array.h"
#include "./get_changed_ranges.h"
#include "./length.h"
#include "./subtree.h"
#include "./tree_cursor.h"
#include "./tree.h"

TSTree *ts_tree_new(
  Subtree root, const TSLanguage *language,
  const TSRange *included_ranges, unsigned included_range_count
) {
  TSTree *result = ts_malloc(sizeof(TSTree));
  result->root = root;
  result->language = language;
  result->included_ranges = ts_calloc(included_range_count, sizeof(TSRange));
  memcpy(result->included_ranges, included_ranges, included_range_count * sizeof(TSRange));
  result->included_range_count = included_range_count;
  return result;
}

TSTree *ts_tree_copy(const TSTree *self) {
  ts_subtree_retain(self->root);
  return ts_tree_new(self->root, self->language, self->included_ranges, self->included_range_count);
}

void ts_tree_delete(TSTree *self) {
  if (!self) return;

  SubtreePool pool = ts_subtree_pool_new(0);
  ts_subtree_release(&pool, self->root);
  ts_subtree_pool_delete(&pool);
  ts_free(self->included_ranges);
  ts_free(self);
}

TSNode ts_tree_root_node(const TSTree *self) {
  return ts_node_new(self, &self->root, ts_subtree_padding(self->root), 0);
}

TSNode ts_tree_root_node_with_offset(
  const TSTree *self,
  uint32_t offset_bytes,
  TSPoint offset_extent
) {
  Length offset = {offset_bytes, offset_extent};
  return ts_node_new(self, &self->root, length_add(offset, ts_subtree_padding(self->root)), 0);
}

const TSLanguage *ts_tree_language(const TSTree *self) {
  return self->language;
}

void ts_tree_edit(TSTree *self, const TSInputEdit *edit) {
  for (unsigned i = 0; i < self->included_range_count; i++) {
    TSRange *range = &self->included_ranges[i];
    if (range->end_byte >= edit->old_end_byte) {
      if (range->end_byte != UINT32_MAX) {
        range->end_byte = edit->new_end_byte + (range->end_byte - edit->old_end_byte);
        range->end_point = point_add(
          edit->new_end_point,
          point_sub(range->end_point, edit->old_end_point)
        );
        if (range->end_byte < edit->new_end_byte) {
          range->end_byte = UINT32_MAX;
          range->end_point = POINT_MAX;
        }
      }
    } else if (range->end_byte > edit->start_byte) {
      range->end_byte = edit->start_byte;
      range->end_point = edit->start_point;
    }
    if (range->start_byte >= edit->old_end_byte) {
      range->start_byte = edit->new_end_byte + (range->start_byte - edit->old_end_byte);
      range->start_point = point_add(
        edit->new_end_point,
        point_sub(range->start_point, edit->old_end_point)
      );
      if (range->start_byte < edit->new_end_byte) {
        range->start_byte = UINT32_MAX;
        range->start_point = POINT_MAX;
      }
    } else if (range->start_byte > edit->start_byte) {
      range->start_byte = edit->start_byte;
      range->start_point = edit->start_point;
    }
  }

  SubtreePool pool = ts_subtree_pool_new(0);
  self->root = ts_subtree_edit(self->root, edit, &pool);
  ts_subtree_pool_delete(&pool);
}

TSRange *ts_tree_included_ranges(const TSTree *self, uint32_t *length) {
  *length = self->included_range_count;
  TSRange *ranges = ts_calloc(self->included_range_count, sizeof(TSRange));
  memcpy(ranges, self->included_ranges, self->included_range_count * sizeof(TSRange));
  return ranges;
}

TSRange *ts_tree_get_changed_ranges(const TSTree *self, const TSTree *other, uint32_t *count) {
  TreeCursor cursor1 = {NULL, array_new()};
  TreeCursor cursor2 = {NULL, array_new()};
  ts_tree_cursor_init(&cursor1, ts_tree_root_node(self));
  ts_tree_cursor_init(&cursor2, ts_tree_root_node(other));

  TSRangeArray included_range_differences = array_new();
  ts_range_array_get_changed_ranges(
    self->included_ranges, self->included_range_count,
    other->included_ranges, other->included_range_count,
    &included_range_differences
  );

  TSRange *result;
  *count = ts_subtree_get_changed_ranges(
    &self->root, &other->root, &cursor1, &cursor2,
    self->language, &included_range_differences, &result
  );

  array_delete(&included_range_differences);
  array_delete(&cursor1.stack);
  array_delete(&cursor2.stack);
  return result;
}

#ifdef _WIN32

void ts_tree_print_dot_graph(const TSTree *self, int fd) {
  (void)self;
  (void)fd;
}

#else

#include <unistd.h>

void ts_tree_print_dot_graph(const TSTree *self, int fd) {
  FILE *file = fdopen(dup(fd), "a");
  ts_subtree_print_dot_graph(self->root, self->language, file);
  fclose(file);
}

#endif
