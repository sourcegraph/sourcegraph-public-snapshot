#include "api.h"
#include "bindings.h"
#include <string.h>
#include <stdio.h>

static void stderr_log(void *payload, TSLogType type, const char *msg)
{
    bool include_lexing = (bool)payload;
    switch (type)
    {
    case TSLogTypeParse:
        fprintf(stderr, "* %s\n", msg);
        break;
    case TSLogTypeLex:
        if (include_lexing)
            fprintf(stderr, "  %s\n", msg);
        break;
    }
}

TSLogger stderr_logger_new(bool include_lexing)
{
    TSLogger result;
    result.payload = (void *)include_lexing;
    result.log = stderr_log;
    return result;
}

const char *call_callReadFunc(void *payload, uint32_t byte_index, TSPoint position, uint32_t *bytes_read)
{
    ParsePayload *p = payload;
    if (p->previous_content != NULL)
    {
        free(p->previous_content);
    }
    p->previous_content = callReadFunc(p->read_function_id, byte_index, position, bytes_read);
    return p->previous_content;
}

TSTree *call_ts_parser_parse(TSParser *self, const TSTree *old_tree, int read_function_id, TSInputEncoding encoding)
{
    ParsePayload payload = {read_function_id, NULL};
    TSInput input = {&payload, call_callReadFunc, encoding};
    TSTree *tree = ts_parser_parse(self, old_tree, input);
    if (payload.previous_content != NULL)
    {
        free(payload.previous_content);
    }
    return tree;
}