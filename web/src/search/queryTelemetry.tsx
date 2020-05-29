import { count } from '../../../shared/src/util/strings'
import { parseSearchQuery, ParserResult, Sequence } from '../../../shared/src/search/parser/parser'
import { resolveFilter } from '../../../shared/src/search/parser/filters'

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
export function queryTelemetryData(query: string, caseSensitive: boolean) {
    return {
        // 🚨 PRIVACY: never provide any private data in { code_search: { query_data: { query } } }.
        query: query ? queryStringTelemetryData(query, caseSensitive) : undefined,
        combined: query,
        empty: !query,
    }
}
function filterExistsInQuery(parsedQuery: ParserResult<Sequence>, filterToMatch: string): boolean {
    if (parsedQuery.type === 'success') {
        const members = parsedQuery.token.members
        for (const member of members) {
            if (member.token.type === 'filter') {
                const resolvedFilter = resolveFilter(member.token.filterType.token.value)
                if (resolvedFilter !== undefined && resolvedFilter.type === filterToMatch) {
                    return true
                }
            }
        }
    }
    return false
}
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
function queryStringTelemetryData(q: string, caseSensitive: boolean) {
    // 🚨 PRIVACY: never provide any private data in this function's return value.
    // This only takes ~1.7ms per call, so it does not need to be optimized.
    const parsedQuery = parseSearchQuery(q)
    return {
        field_archived: filterExistsInQuery(parsedQuery, 'archived')
            ? {
                  count: count(q, /(^|\s)archived:/g),
              }
            : undefined,
        field_after: filterExistsInQuery(parsedQuery, 'after')
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
        field_author: filterExistsInQuery(parsedQuery, 'author')
            ? {
                  count: count(q, /(^|\s)author:/g),
                  count_negated: count(q, /(^|\s)-author:/g),
                  value_at_sign: count(q, /(^|\s)-?author:[^s]*@/g),
              }
            : undefined,
        field_before: filterExistsInQuery(parsedQuery, 'before')
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
        field_case: caseSensitive
            ? {
                  count: 1,
              }
            : undefined,
        field_committer: filterExistsInQuery(parsedQuery, 'committer')
            ? {
                  count: count(q, /(^|\s)committer:/g),
                  count_negated: count(q, /(^|\s)-committer:/g),
                  value_at_sign: count(q, /(^|\s)-?committer:[^s]*@/g),
              }
            : undefined,
        field_content: filterExistsInQuery(parsedQuery, 'content')
            ? {
                  count: count(q, /(^|\s)content:/g),
              }
            : undefined,
        field_count: filterExistsInQuery(parsedQuery, 'count')
            ? {
                  count: count(q, /(^|\s)count:/g),
              }
            : undefined,
        field_file: filterExistsInQuery(parsedQuery, 'file')
            ? {
                  count: count(q, /(^|\s)f(ile)?:/g),
                  count_negated: count(q, /(^|\s)-f(ile?):/g),
                  count_alias: count(q, /(^|\s)-?f:/g),

                  // likely regexp char
                  value_regexp: count(q, /(^|\s)-?f(ile)?:\S*[$()*?[\]^|]/g),

                  // likely regexp matching a file ext
                  value_regexp_file_ext: count(q, /(^|\s)-?f(ile)?:\S*\.\w+\$\b/g),

                  // oops! user tried to use a (likely) glob
                  value_glob: count(q, /(^|\s)-?f(ile)?:\S*(\*\.|\.{[A-Za-z]|\*\*|\/\*)/g),
              }
            : undefined,
        field_fork: filterExistsInQuery(parsedQuery, 'fork')
            ? {
                  count: count(q, /(^|\s)fork:/g),
              }
            : undefined,
        field_index: filterExistsInQuery(parsedQuery, 'index')
            ? {
                  count: count(q, /(^|\s)index:/g),
              }
            : undefined,
        field_lang: filterExistsInQuery(parsedQuery, 'lang')
            ? {
                  count: count(q, /(^|\s)l(ang)?:/g),
                  count_negated: count(q, /(^|\s)-l(ang)?:/g),
                  count_alias: count(q, /(^|\s)-?l:/g),
              }
            : undefined,
        field_message: filterExistsInQuery(parsedQuery, 'message')
            ? {
                  count: count(q, /(^|\s)m(essage)?:/g),
                  count_negated: count(q, /(^|\s)-m(essage)?:/g),
                  count_alias: count(q, /(^|\s)-?m:/g),
              }
            : undefined,
        field_patterntype: filterExistsInQuery(parsedQuery, 'patterntype')
            ? {
                  count: count(q, /(^|\s)patterntype:/gi),
              }
            : undefined,
        field_repo: filterExistsInQuery(parsedQuery, 'repo')
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
                  value_regexp: count(q, /(^|\s)-?r(epo)?:\S*[$()*?[\]^|]/g),

                  // oops! user tried to use a (likely) glob
                  value_glob: count(q, /(^|\s)-?r(epo)?:\S*(\*\.|\.{[A-Za-z]|\*\*|\/\*)/g),
              }
            : undefined,
        field_repogroup: filterExistsInQuery(parsedQuery, 'repogroup')
            ? {
                  count: count(q, /(^|\s)(repogroup|g):/g),
                  count_negated: count(q, /(^|\s)-(repogroup|g):/g),
                  count_alias: count(q, /(^|\s)-?g:/g),
                  value_active: count(q, /(^|\s)-?(repogroup|g):active/g),
                  value_inactive: count(q, /(^|\s)-?(repogroup|g):active/g),
              }
            : undefined,
        field_repohascommitafter: filterExistsInQuery(parsedQuery, 'repohascommitafter')
            ? {
                  count: count(q, /(^|\s)(repohascommitafter):/g),
              }
            : undefined,
        field_repohasfile: filterExistsInQuery(parsedQuery, 'repohasfile')
            ? {
                  count: count(q, /(^|\s)(repohasfile):/g),
              }
            : undefined,
        field_timeout: filterExistsInQuery(parsedQuery, 'timeout')
            ? {
                  count: count(q, /(^|\s)timeout:/g),
              }
            : undefined,
        field_type: filterExistsInQuery(parsedQuery, 'type')
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
            count: count(q, /(^|\s)(\w+:)?([^\s"'/:]+|"[^"]*"|'[^']*'|\/[^/]*\/)/g),
            count_non_default: count(q, /(^|\s)\w+:([^\s"'/:]+|"[^"]*"|'[^']*')/g),
        },
        chars: {
            count: q.length,
            non_ascii: count(q, /[^\t\n\r -~]/g),
            space: count(q, /\s/g),
            double_quote: count(q, /"/g),
            single_quote: count(q, /'/g),
        },
    }
}

function defaultQueryFieldTelemetryData(q: string): { [key: string]: any } {
    // 🚨 PRIVACY: never provide any private data in this function's return value.

    // Strip non-default fields. Does not account for backslashes.
    q = q.replace(/(^|\s)\w+:([^\s"'/:]+|"[^"]*"|'[^']*')/g, ' ')

    return {
        count: count(q, /(^|\s)([^\s"'/:]+|"[^"]*"|'[^']*'|\/[^/]*\/)/g),
        count_literal: count(q, /(^|\s)[^\s"'/:]+/g),
        count_double_quote: count(q, /(^|\s)"[^"]*"/g),
        count_single_quote: count(q, /(^|\s)'[^']*'/g),
        count_pattern: count(q, /(^|\s)\/[^/]*\//g),
        count_regexp: count(q, /(^|\s)[$()*?[\]^|]/g),
    }
}
