import * as Monaco from 'monaco-editor'

import { SearchPatternType } from '../../graphql-operations'

import { validateFilter } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import {
    PatternOf,
    each,
    matchesValue,
    eachOf,
    allOf,
    some,
    oneOf,
    not,
    DataMapper,
    MatchContext,
} from './patternMatcher'
import { Filter, Token } from './token'

type FilterCheck = (f: Filter) => Monaco.editor.IMarkerData[]

function createMarker(
    message: string,
    filter: Filter,
    severity = Monaco.MarkerSeverity.Error
): Monaco.editor.IMarkerData {
    return {
        severity,
        message,
        ...toMonacoSingleLineRange(filter.field.range),
    }
}

export function validFilterValue(filter: Filter): Monaco.editor.IMarkerData[] {
    const validationResult = validateFilter(filter.field.value, filter.value)
    if (validationResult.valid) {
        return []
    }
    return [createMarker(`Error: ${validationResult.reason}`, filter)]
}

export function emptyFilterValue(filter: Filter): Monaco.editor.IMarkerData[] {
    if (filter.value?.value !== '') {
        return []
    }
    return [
        createMarker(
            `Warning: This filter is empty. Remove the space between the filter and value or quote the value to include the space. E.g., ${filter.field.value}:" a term".`,
            filter,
            Monaco.MarkerSeverity.Warning
        ),
    ]
}

// Returns the first nonempty diagnostic for a filter, or nothing otherwise. We return
// the only the first so that we don't overwhelm the the user with multiple diagnostics
// for a single filter.
export function checkFilter(filter: Filter): Monaco.editor.IMarkerData[] {
    const checks: FilterCheck[] = [validFilterValue, emptyFilterValue]
    return checks.map(check => check(filter)).find(value => value.length !== 0) || []
}

// There should be a better way to pass additional data to patterns, or for
// patterns to keep internal state than piggybacking on `data`
// (tracked in https://github.com/sourcegraph/sourcegraph/issues/25070)
interface PatternData {
    // Used to make the search pattern type available to patterns
    searchPatternType: SearchPatternType
    // Used by patterns to "export" markers
    marker: Monaco.editor.IMarkerData[]
    // Used by the rev+repo patterns to keep track of the rev filter
    revFilter?: Filter
}

function addFilterDiagnostic(message: string, severity = Monaco.MarkerSeverity.Error): DataMapper<Token, PatternData> {
    return (token, context) => {
        context.data.marker.push(createMarker(message, token as Filter, severity))
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
            context.data.marker.push(...checkFilter(token as Filter))
        },
    }),

    // Validates that author/before/after/message fields are only valid
    // type:diff/commit queries
    allOf(
        not(some({ field: { value: 'type' }, value: { value: oneOf('diff', 'commit') } })),
        each({
            field: { value: oneOf('author', 'before', 'until', 'after', 'since', 'message', 'msg', 'm') },
            $data: addFilterDiagnostic('Error: this filter requires "type:commit" or "type:diff" in the query'),
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
                $data: (_tokens, context) => {
                    context.data.marker.push(
                        createMarker(
                            'Error: query contains "rev:" without "repo:". Add a "repo:" filter.',
                            context.data.revFilter!
                        )
                    )
                },
            },
            // Only empty repo: filter
            allOf(
                not(some({ ...repoFilterPattern, value: { value: value => value !== '' } })),
                some({
                    ...repoFilterPattern,
                    value: { value: '' },
                    $data: (token: Token, context: MatchContext<PatternData>) => {
                        const errorMessage =
                            'Error: query contains "rev:" with an empty "repo:" filter. Add a non-empty "repo:" filter.'
                        context.data.marker.push(
                            createMarker(errorMessage, token as Filter),
                            createMarker(errorMessage, context.data.revFilter!)
                        )
                    },
                })
            ),
            // Repo filter that contains a revision "tag"
            some({
                ...repoFilterPattern,
                value: { value: value => value.includes('@') },
                $data: (token: Token, context: MatchContext<PatternData>) => {
                    const errorMessage =
                        'Error: You have specified both "@" and "rev:" for a repo filter and I don"t know how to interpret this. Remove either "@" or "rev:"'
                    context.data.marker.push(
                        createMarker(errorMessage, token as Filter),
                        createMarker(errorMessage, context.data.revFilter!)
                    )
                },
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
            $data: addFilterDiagnostic(
                'Error: Structural search syntax only applies to searching file contents and is not compatible with "type:". Remove this filter or switch to a different search type.'
            ),
        })
    ),
]

/**
 * Returns the diagnostics for a scanned search query to be displayed in the Monaco query input.
 */
export function getDiagnostics(tokens: Token[], searchPatternType: SearchPatternType): Monaco.editor.IMarkerData[] {
    const result = matchesValue<Token[], PatternData>(tokens, eachOf(...rules), { searchPatternType, marker: [] })
    if (result.success) {
        return result.data.marker
    }
    return []
}
