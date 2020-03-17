import { count } from '../../../shared/src/util/strings'

export function queryTelemetryData(query: string, caseSensitive: boolean): { [key: string]: any } {
    return {
        // 🚨 PRIVACY: never provide any private data in { code_search: { query_data: { query } } }.
        query: query ? queryStringTelemetryData(query, caseSensitive) : undefined,
        combined: query,
        empty: !query,
    }
}

function queryStringTelemetryData(q: string, caseSensitive: boolean): { [key: string]: any } {
    // 🚨 PRIVACY: never provide any private data in this function's return value.
    // This only takes ~1.7ms per call, so it does not need to be optimized.
    return {
        field_archived: {
            count: count(q, /(^|\s)archived:/g),
        },
        field_after: q.includes('after:')
            ? {
                  count: count(q, /(^|\s)after:/g),
                  value_sec: count(q, /(^|\s)after:"[^"]*sec/g),
                  value_min: count(q, /(^|\s)after:"[^"]*min/g),
                  value_hour: count(q, /(^|\s)after:"[^"]*h(ou)?r/g),
                  value_day: count(q, /(^|\s)after:"[^"]*day/g),
                  value_week: count(q, /(^|\s)after:"[^"]*w(ee)?k/g),
                  value_month: count(q, /(^|\s)after:"[^"]*m(on|th)/g),
                  value_year: count(q, /(^|\s)after:"[^"]*y(ea)?r/g),
              }
            : undefined,
        field_author: q.includes('author:')
            ? {
                  count: count(q, /(^|\s)author:/g),
                  count_negated: count(q, /(^|\s)-author:/g),
                  value_at_sign: count(q, /(^|\s)-?author:[^s]*@/g),
              }
            : undefined,
        field_before: q.includes('before:')
            ? {
                  count: count(q, /(^|\s)before:/g),
                  value_sec: count(q, /(^|\s)before:"[^"]*sec/g),
                  value_min: count(q, /(^|\s)before:"[^"]*min/g),
                  value_hour: count(q, /(^|\s)before:"[^"]*h(ou)?r/g),
                  value_day: count(q, /(^|\s)before:"[^"]*day/g),
                  value_week: count(q, /(^|\s)before:"[^"]*w(ee)?k/g),
                  value_month: count(q, /(^|\s)before:"[^"]*m(on|th)/g),
                  value_year: count(q, /(^|\s)before:"[^"]*y(ea)?r/g),
              }
            : undefined,
        field_case: {
            count: caseSensitive ? 1 : 0,
        },
        field_committer: q.includes('committer:')
            ? {
                  count: count(q, /(^|\s)committer:/g),
                  count_negated: count(q, /(^|\s)-committer:/g),
                  value_at_sign: count(q, /(^|\s)-?committer:[^s]*@/g),
              }
            : undefined,
        field_content: {
            count: count(q, /(^|\s)content:/g),
        },
        field_count: {
            count: count(q, /(^|\s)count:/g),
        },
        field_file:
            q.includes('file:') || q.includes('f:')
                ? {
                      count: count(q, /(^|\s)f(ile)?:/g),
                      count_negated: count(q, /(^|\s)-f(ile?):/g),
                      count_alias: count(q, /(^|\s)-?f:/g),

                      // likely regexp char
                      value_regexp: count(q, /(^|\s)-?f(ile)?:[^\s]*[$?[\]|()*^]/g),

                      // likely regexp matching a file ext
                      value_regexp_file_ext: count(q, /(^|\s)-?f(ile)?:[^\s]*\.\w+\$\b/g),

                      // oops! user tried to use a (likely) glob
                      value_glob: count(q, /(^|\s)-?f(ile)?:[^\s]*(\*\.|\.\{[a-zA-Z]|\*\*|\/\*)/g),
                  }
                : undefined,
        field_fork: {
            count: count(q, /(^|\s)fork:/g),
        },
        field_index: {
            count: count(q, /(^|\s)index:/g),
        },
        field_lang:
            q.includes('lang:') || q.includes('l:')
                ? {
                      count: count(q, /(^|\s)l(ang)?:/g),
                      count_negated: count(q, /(^|\s)-l(ang)?:/g),
                      count_alias: count(q, /(^|\s)-?l:/g),
                  }
                : undefined,
        field_message:
            q.includes('message:') || q.includes('m:')
                ? {
                      count: count(q, /(^|\s)m(essage)?:/g),
                      count_negated: count(q, /(^|\s)-m(essage)?:/g),
                      count_alias: count(q, /(^|\s)-?m:/g),
                  }
                : undefined,
        field_patterntype: {
            count: count(q, /(^|\s)patterntype:/gi),
        },
        field_repo:
            q.includes('repo:') || q.includes('r:')
                ? {
                      count: count(q, /(^|\s)r(epo)?:/g),
                      count_negated: count(q, /(^|\s)-r(epo)?:/g),
                      count_alias: count(q, /(^|\s)-?r:/g),
                      value_at_sign: count(q, /(^|\s)-?r(epo)?:[^s]*@/g),
                      value_pipe: count(q, /(^|\s)-?r(epo)?:[^s]*\|/g),
                      value_rev_star: count(q, /(^|\s)-?r(epo)?:[^s]*@[^s]*\*/g),
                      value_rev_colon: count(q, /(^|\s)-?r(epo)?:[^s]*@[^s]*:/g),
                      value_rev_caret: count(q, /(^|\s)-?r(epo)?:[^s]*@[^s]*\^/g),

                      // likely regexp char
                      value_regexp: count(q, /(^|\s)-?r(epo)?:[^\s]*[$?[\]|()*^]/g),

                      // oops! user tried to use a (likely) glob
                      value_glob: count(q, /(^|\s)-?r(epo)?:[^\s]*(\*\.|\.\{[a-zA-Z]|\*\*|\/\*)/g),
                  }
                : undefined,
        field_repogroup:
            q.includes('repogroup:') || q.includes('g:')
                ? {
                      count: count(q, /(^|\s)(repogroup|g):/g),
                      count_negated: count(q, /(^|\s)-(repogroup|g):/g),
                      count_alias: count(q, /(^|\s)-?g:/g),
                      value_active: count(q, /(^|\s)-?(repogroup|g):active/g),
                      value_inactive: count(q, /(^|\s)-?(repogroup|g):active/g),
                  }
                : undefined,
        field_repohascommitafter: q.includes('repohascommitafter:')
            ? {
                  count: count(q, /(^|\s)(repohascommitafter):/g),
              }
            : undefined,
        field_repohasfile: q.includes('repohascommitafter:')
            ? {
                  count: count(q, /(^|\s)(repohasfile):/g),
              }
            : undefined,
        field_timeout: {
            count: count(q, /(^|\s)timeout:/g),
        },
        field_type: q.includes('type:')
            ? {
                  count: count(q, /(^|\s)type:/g),
                  value_file: count(q, /(^|\s)type:file(\s|$)/g),
                  value_diff: count(q, /(^|\s)type:diff(\s|$)/g),
                  value_commit: count(q, /(^|\s)type:commit(\s|$)/g),
                  value_symbol: count(q, /(^|\s)type:symbol(\s|$)/g),
              }
            : undefined,
        field_default: defaultQueryFieldTelemetryData(q),
        fields: {
            count: count(q, /(^|\s)(\w+:)?([^:"'/\s]+|"[^"]*"|'[^']*'|\/[^/]*\/)/g),
            count_non_default: count(q, /(^|\s)\w+:([^:"'/\s]+|"[^"]*"|'[^']*')/g),
        },
        chars: {
            count: q.length,
            non_ascii: count(q, /[^ -~\t\n\r]/g),
            space: count(q, /\s/g),
            double_quote: count(q, /"/g),
            single_quote: count(q, /'/g),
        },
    }
}

function defaultQueryFieldTelemetryData(q: string): { [key: string]: any } {
    // 🚨 PRIVACY: never provide any private data in this function's return value.

    // Strip non-default fields. Does not account for backslashes.
    q = q.replace(/(^|\s)\w+:([^:"'/\s]+|"[^"]*"|'[^']*')/g, ' ')

    return {
        count: count(q, /(^|\s)([^:"'/\s]+|"[^"]*"|'[^']*'|\/[^/]*\/)/g),
        count_literal: count(q, /(^|\s)[^:"'/\s]+/g),
        count_double_quote: count(q, /(^|\s)"[^"]*"/g),
        count_single_quote: count(q, /(^|\s)'[^']*'/g),
        count_pattern: count(q, /(^|\s)\/[^/]*\//g),
        count_regexp: count(q, /(^|\s)[$?[\]|()*^]/g),
    }
}
