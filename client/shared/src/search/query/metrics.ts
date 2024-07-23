import { resolveFieldAlias } from './filters'
import { scanPredicate } from './predicates'
import { scanSearchQuery } from './scanner'
import { KeywordKind } from './token'

const nonzero = (value: number): number | undefined => (value === 0 ? undefined : value)

interface Metrics {
    count_or?: number
    count_and?: number
    count_not?: number
    count_select_repo?: number
    count_select_file?: number
    count_select_content?: number
    count_select_symbol?: number
    count_select_commit_diff_added?: number
    count_select_commit_diff_removed?: number
    count_repo_contains_path?: number
    count_repo_contains_file?: number
    count_repo_contains_content?: number
    count_repo_contains_commit_after?: number
    count_repo_dependencies?: number
    count_count_all?: number
    count_non_global_context?: number
    count_only_patterns?: number
    count_only_patterns_three_or_more?: number
}

export const collectMetrics = (query: string): Metrics | undefined => {
    const tokens = scanSearchQuery(query)
    if (tokens.type !== 'success') {
        return undefined
    }

    let count_or = 0
    let count_and = 0
    let count_not = 0
    let count_select_repo = 0
    let count_select_file = 0
    let count_select_content = 0
    let count_select_symbol = 0
    let count_select_commit_diff_added = 0
    let count_select_commit_diff_removed = 0
    let count_repo_contains_path = 0
    let count_repo_contains_file = 0
    let count_repo_contains_content = 0
    let count_repo_contains_commit_after = 0
    const count_repo_dependencies = 0
    let count_count_all = 0
    let count_non_global_context = 0
    let count_only_patterns = 0
    let count_only_patterns_three_or_more = 0

    const onlyPatterns = tokens.term.reduce(
        (result, token) =>
            result &&
            (token.type === 'pattern' ||
                token.type === 'whitespace' ||
                (token.type === 'filter' && token.field.value === 'context' && token.value?.value === 'global')),
        true
    )

    if (onlyPatterns) {
        count_only_patterns += 1
        if (tokens.term.filter(token => token.type === 'pattern').length >= 3) {
            count_only_patterns_three_or_more += 1
        }
        // We're done, we can't learn more about this query.
        return {
            // RFC 384: unqualified queries
            count_only_patterns,
            count_only_patterns_three_or_more: nonzero(count_only_patterns_three_or_more),
        }
    }

    for (const token of tokens.term) {
        switch (token.type) {
            case 'keyword': {
                switch (token.kind) {
                    case KeywordKind.Or: {
                        count_or += 1
                        break
                    }
                    case KeywordKind.And: {
                        count_and += 1
                        break
                    }
                    case KeywordKind.Not: {
                        count_not += 1
                        break
                    }
                }
                break
            }
            case 'filter': {
                if (!token.value) {
                    continue
                }
                switch (resolveFieldAlias(token.field.value)) {
                    case 'select': {
                        switch (token.value.value) {
                            case 'repo': {
                                count_select_repo += 1
                                break
                            }
                            case 'file': {
                                count_select_file += 1
                                break
                            }
                            case 'content': {
                                count_select_content += 1
                                break
                            }
                            case 'symbol': {
                                count_select_symbol += 1
                                break
                            }
                            case 'commit.diff.added': {
                                count_select_commit_diff_added += 1
                                break
                            }
                            case 'commit.diff.removed': {
                                count_select_commit_diff_removed += 1
                                break
                            }
                        }
                    }
                    case 'repo': {
                        const predicate = scanPredicate('repo', token.value.value)
                        if (!predicate) {
                            continue
                        }
                        switch (predicate.name) {
                            case 'contains.path': {
                                count_repo_contains_path += 1
                                break
                            }
                            case 'contains.file': {
                                count_repo_contains_file += 1
                                break
                            }
                            case 'contains.content': {
                                count_repo_contains_content += 1
                                break
                            }
                            case 'contains.commit.after': {
                                count_repo_contains_commit_after += 1
                                break
                            }
                        }
                    }
                    case 'count': {
                        if (token.value.value === 'all') {
                            count_count_all += 1
                        }
                        break
                    }
                    case 'context': {
                        if (token.value.value !== 'global') {
                            count_non_global_context += 1
                        }
                        break
                    }
                }
            }
        }
    }
    return {
        // RFC 384: operator presence
        count_or: nonzero(count_or),
        count_and: nonzero(count_and),
        count_not: nonzero(count_not),
        // RFC 384: select usage
        count_select_repo: nonzero(count_select_repo),
        count_select_file: nonzero(count_select_file),
        count_select_content: nonzero(count_select_content),
        count_select_symbol: nonzero(count_select_symbol),
        count_select_commit_diff_added: nonzero(count_select_commit_diff_added),
        count_select_commit_diff_removed: nonzero(count_select_commit_diff_removed),
        // RFC 384: predicate usage
        count_repo_contains_path: nonzero(count_repo_contains_path),
        count_repo_contains_file: nonzero(count_repo_contains_file),
        count_repo_contains_content: nonzero(count_repo_contains_content),
        count_repo_contains_commit_after: nonzero(count_repo_contains_commit_after),
        count_repo_dependencies: nonzero(count_repo_dependencies),
        // RFC 384: exhaustive search frequency
        count_count_all: nonzero(count_count_all),
        // RFC 384: context usage
        count_non_global_context: nonzero(count_non_global_context),
    }
}
