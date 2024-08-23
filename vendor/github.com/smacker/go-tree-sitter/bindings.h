#ifndef TREE_SITTER_BINDINGS_H_
#define TREE_SITTER_BINDINGS_H_

#include "api.h"

TSLogger stderr_logger_new(bool include_lexing);

typedef struct
{
    int read_function_id;
    char *previous_content;
} ParsePayload;

extern char *callReadFunc(int id, uint32_t byteIndex, TSPoint position, uint32_t *bytesRead);
TSTree *call_ts_parser_parse(TSParser *self, const TSTree *old_tree, int read_function_id, TSInputEncoding encoding);

#endif
