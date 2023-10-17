import { SearchPatternType } from '../../graphql-operations'

import { validateFilter } from './filters'
import {
    type PatternOf,
    each,
    matchesValue,
    eachOf,
    allOf,
    some,
    oneOf,
    not,
    type DataMapper,
    type MatchContext,
} from './patternMatcher'
import type { CharacterRange, Filter, Token } from './token'

interface ChangeSpec {
    from: number
    to?: number
    insert?: string
}

interface Action {
    label: string
    change?: ChangeSpec
    selection?: { head?: number; anchor: number }
}

export interface Diagnostic {
    range: CharacterRange
    message: string
    severity: 'info' | 'warning' | 'error'
    actions?: Action[]
}

type FilterCheck = (f: Filter) => Diagnostic[]

function createDiagnostic(
    message: string,
    token: Token,
    severity: Diagnostic['severity'] = 'error',
    actions?: Action[]
): Diagnostic {
    return {
        severity,
        message,
        range: token.range,
        actions,
    }
}

function validFilterValue(filter: Filter): Diagnostic[] {
    if (!filter.value) {
        return []
    }

    const validationResult = validateFilter(filter.field.value, filter.value)
    if (validationResult.valid) {
        return []
    }
    return [createDiagnostic(validationResult.reason, filter)]
}

function emptyFilterValue(filter: Filter): Diagnostic[] {
    if (filter.value?.value !== '') {
        return []
    }
    return [
        createDiagnostic(
            `This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g. \`${filter.field.value}:" a term"\`.`,
            filter,
            'warning'
        ),
    ]
}

// Returns the first nonempty diagnostic for a filter, or nothing otherwise. We return
// the only the first so that we don't overwhelm the the user with multiple diagnostics
// for a single filter.
function checkFilter(filter: Filter): Diagnostic[] {
    const checks: FilterCheck[] = [validFilterValue, emptyFilterValue]
    return checks.map(check => check(filter)).find(value => value.length !== 0) || []
}

// TODO(fkling): Improve how we pass additional data to patterns #25070
interface PatternData {
    // Used to make the search pattern type available to patterns
    searchPatternType: SearchPatternType
    // Used by patterns to "export" markers
    diagnostics: Diagnostic[]
    // Used by the rev+repo patterns to keep track of the rev filter
    revFilter?: Filter
    // Used by structural search check to add language
    patterns: Token[]
}

function filterDiagnosticCreator(
    message: string,
    severity: Diagnostic['severity'] = 'error'
): (token: Token) => Diagnostic[] {
    return token => [createDiagnostic(message, token as Filter, severity)]
}

function addDiagnostic<T>(
    provider: (value: T, context: MatchContext<PatternData>) => Diagnostic[]
): DataMapper<T, PatternData> {
    return (value, context) => {
        context.data.diagnostics.push(...provider(value, context))
    }
}

const repoFilterPattern: PatternOf<Token, any> = {
    field: { value: oneOf('repo', 'r') },
}

const rules: PatternOf<Token[], PatternData>[] = [
    // Validate the value of each filter
    each({
        type: 'filter',
        $data: (token: Token, context: MatchContext<PatternData>) => {
            context.data.diagnostics.push(...checkFilter(token as Filter))
        },
    }),

    // Validates that author/before/after/message fields are only valid
    // type:diff/commit queries
    allOf(
        not(some({ field: { value: 'type' }, value: { value: oneOf('diff', 'commit') } })),
        each({
            field: { value: oneOf('author', 'before', 'until', 'after', 'since', 'message', 'msg', 'm') },
            $data: addDiagnostic(token => [
                createDiagnostic('This filter requires `type:commit` or `type:diff` in the query', token, 'error', [
                    {
                        label: 'Add "type:commit"',
                        change: { from: token.range.start, insert: 'type:commit ' },
                    },
                    {
                        label: 'Add "type:diff"',
                        change: { from: token.range.start, insert: 'type:diff ' },
                    },
                ]),
            ]),
        })
    ),

    // Validates that query contains a valid repo: filter if it contains a rev:
    // filter.
    // NOTE: We don't have to explicitly verify that the repo: filters are
    // negated because they cannot be matched by these patterns anyway (the
    // field value would be `-repo`, not `repo`).
    allOf(
        some({ field: { value: oneOf('rev', 'revision') }, $data: { revFilter: token => token as Filter } }), // "remember" rev filter token
        oneOf(
            // No repo: filter
            {
                $pattern: not(some(repoFilterPattern)),
                $data: addDiagnostic((_tokens, context) => {
                    const revisionFilter = context.data.revFilter!

                    return [
                        createDiagnostic(
                            'Query contains `rev:` without `repo:`. Add a `repo:` filter.',
                            revisionFilter,
                            'error',
                            [
                                {
                                    label: 'Add "repo:"',
                                    change: { from: revisionFilter.range.end, insert: ' repo:' },
                                    selection: { anchor: revisionFilter.range.end + 6 },
                                },
                            ]
                        ),
                    ]
                }),
            },
            // Only empty repo: filter
            allOf(
                not(some({ ...repoFilterPattern, value: { value: value => value !== '' } })),
                some({
                    ...repoFilterPattern,
                    value: oneOf(undefined, { value: '' }),
                    $data: addDiagnostic((token, context) => {
                        const errorMessage =
                            'Query contains `rev:` with an empty `repo:` filter. Add a non-empty `repo:` filter.'
                        const actions: Action[] = [
                            {
                                label: 'Add "repo:" value',
                                selection: { anchor: token.range.end },
                            },
                        ]
                        return [
                            createDiagnostic(errorMessage, token as Filter, 'error', actions),
                            createDiagnostic(errorMessage, context.data.revFilter!, 'error', actions),
                        ]
                    }),
                })
            ),
            // Repo filter that contains a revision "tag"
            some({
                ...repoFilterPattern,
                value: { value: value => value.includes('@') },
                $data: addDiagnostic((token, context) => {
                    const repoFilter = token as Filter
                    const errorMessage =
                        "You have specified both `@<rev>` and `rev:` for a repo filter and I don't know how to interpret this. Remove either `@<rev>` or `rev:`"
                    const start = repoFilter.value!.range.start
                    const actions: Action[] = [
                        {
                            label: 'Remove @<rev>',
                            change: {
                                from: start + repoFilter.value!.value.indexOf('@'),
                                to: start + repoFilter.value!.value.length,
                            },
                        },
                        {
                            label: 'Remove "rev:" filter',
                            change: {
                                from: context.data.revFilter!.range.start,
                                to: context.data.revFilter!.range.end,
                            },
                        },
                    ]
                    return [
                        createDiagnostic(errorMessage, token as Filter, 'error', actions),
                        createDiagnostic(errorMessage, context.data.revFilter!, 'error', actions),
                    ]
                }),
            })
        )
    ),

    // Validates that query can be evaluated as structural search query
    allOf(
        oneOf(
            some({ field: { value: 'patterntype' }, value: { value: 'structural' } }),
            allOf(
                (_tokens, context) => context.data.searchPatternType === SearchPatternType.structural,
                not(some({ field: { value: 'patterntype' }, value: { value: not('structural') } }))
            )
        ),
        some({
            field: { value: 'type' },
            $data: addDiagnostic(
                filterDiagnosticCreator(
                    'Structural search syntax only applies to searching file contents and is not compatible with `type:`. Remove this filter or switch to a different search type.'
                )
            ),
        })
    ),

    // Warn if structural search runs without `lang:` filter
    allOf(
        oneOf(
            some({ field: { value: 'patterntype' }, value: { value: 'structural' } }),
            allOf(
                (_tokens, context) => context.data.searchPatternType === SearchPatternType.structural,
                not(some({ field: { value: 'patterntype' }, value: { value: not('structural') } }))
            )
        ),
        oneOf({
            $pattern: not(some({ field: { value: oneOf('language', 'lang', 'l') } })),
            $data: addDiagnostic((_tokens, context) =>
                context.data.patterns.length > 0
                    ? [
                          createDiagnostic(
                              'Add a `lang` filter when using structural search. Structural search may miss results without a `lang` filter because it only guesses the language of files searched.',
                              context.data.patterns[0],
                              'warning'
                          ),
                      ]
                    : []
            ),
        })
    ),
]

/**
 * Returns the diagnostics for a scanned search query to be displayed in the query input.
 */
export function getDiagnostics(tokens: Token[], searchPatternType: SearchPatternType): Diagnostic[] {
    const patterns = tokens.filter(token => token.type === 'pattern')
    const result = matchesValue<Token[], PatternData>(tokens, eachOf(...rules), {
        searchPatternType,
        diagnostics: [],
        patterns,
    })
    if (result.success) {
        return result.data.diagnostics
    }
    return []
}
