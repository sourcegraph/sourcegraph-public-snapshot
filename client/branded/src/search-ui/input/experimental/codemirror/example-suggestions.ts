import type { Extension } from '@codemirror/state'
import { sampleSize } from 'lodash'

import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'
import type { Token } from '@sourcegraph/shared/src/search/query/token'
import { resolveFilterMemoized } from '@sourcegraph/shared/src/search/query/utils'

import { getQueryInformation } from '../../codemirror/parsedQuery'
import { queryRenderer } from '../optionRenderer'
import {
    type Action,
    type Option,
    startCompletion,
    suggestionSources,
    selectionListener,
} from '../suggestionsExtension'

interface Example {
    label: string
    description: string
    snippet?: string
    /**
     * Whether or not to show the example.
     */
    valid(tokens: Token[]): boolean
}

const REFRESH_TIMEOUT = 60000 // 1 minute
const EXAMPLES_TO_SHOW = 3

const examples: Example[] = [
    {
        label: 'file:has.owner()',
        snippet: 'file:has.owner(${}) ${}',
        description: 'Search code ownership',
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('has.owner(')),
    },
    {
        label: 'repo:has.path()',
        snippet: 'repo:has.path(${}) ${}',
        description: 'Search in repositories containing a path',
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('has.path(')),
    },
    {
        label: 'repo:has.content()',
        snippet: 'repo:has.content(${}) ${}',
        description: 'Search in repositories with files having specific contents',
        valid: tokens =>
            !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('has.content(')),
    },
    {
        label: '-file:',
        description: FILTERS[FilterType.file].description(true),
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.field.value === '-file'),
    },
    {
        label: '-repo:',
        description: FILTERS[FilterType.repo].description(true),
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.field.value === '-repo'),
    },
    {
        label: 'repo:my-org.*/.*-cli$',
        // eslint-disable-next-line no-template-curly-in-string
        snippet: 'repo:${my-org.*/.*-cli$} ${}',
        description: 'Search in repositories matching a pattern',
        valid: tokens =>
            !tokens.some(
                token => token.type === 'filter' && resolveFilterMemoized(token.field.value)?.type === FilterType.repo
            ),
    },
    {
        label: 'type:diff select:commit.diff.removed TODO',
        // eslint-disable-next-line no-template-curly-in-string
        snippet: 'type:diff select:commit.diff.removed repo:${my-repo} TODO ${}',
        description: 'Find commits that removed "TODO"',
        valid: tokens => !tokens.some(token => token.type === 'filter' && token.value?.value.startsWith('commit.diff')),
    },
]

const choosenExamples = {
    examples: sampleSize(examples, EXAMPLES_TO_SHOW),
    timestamp: Date.now(),
}

export function exampleSuggestions(config: {
    getUsedExamples: () => Set<string>
    markExampleUsed: (example: string) => void
}): Extension {
    const hideAction: Action = {
        type: 'command',
        name: 'Hide',
        apply(_option, view) {
            // Trigger a new completion to refresh the list of examples after hiding one.
            // The example is marked as "used" via the selection listener below.
            window.setTimeout(() => {
                startCompletion(view)
            }, 0)
        },
    }

    return [
        suggestionSources.of({
            query: (state, position) => {
                const { token, tokens } = getQueryInformation(state, position)
                if (token && token.type !== 'whitespace') {
                    return null
                }

                const usedExamples = config.getUsedExamples()
                const validExamples = examples.filter(
                    example => !usedExamples.has(example.label) && example.valid(tokens)
                )

                if (Date.now() - choosenExamples.timestamp > REFRESH_TIMEOUT) {
                    choosenExamples.examples = sampleSize(validExamples, EXAMPLES_TO_SHOW)
                    choosenExamples.timestamp = Date.now()
                }

                let validChosenExamples = choosenExamples.examples.filter(example => validExamples.includes(example))

                if (validChosenExamples.length < EXAMPLES_TO_SHOW) {
                    validChosenExamples = validChosenExamples.concat(
                        sampleSize(
                            validExamples.filter(example => !validChosenExamples.includes(example)),
                            EXAMPLES_TO_SHOW - validChosenExamples.length
                        )
                    )
                }

                return {
                    result: [
                        {
                            title: 'Learn',
                            options: validChosenExamples.map(
                                (example): Option => ({
                                    label: example.label,
                                    description: example.description,
                                    render: queryRenderer,
                                    kind: 'example',
                                    action: {
                                        type: 'completion',
                                        from: position,
                                        insertValue: example.snippet,
                                        asSnippet: !!example.snippet,
                                    },
                                    alternativeAction: hideAction,
                                })
                            ),
                        },
                    ],
                }
            },
        }),
        selectionListener.of(({ option }) => {
            if (option.kind === 'example') {
                config.markExampleUsed(option.label)
            }
        }),
    ]
}
