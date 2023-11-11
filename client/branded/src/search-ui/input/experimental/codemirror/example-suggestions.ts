import type { Extension } from '@codemirror/state'
import { sampleSize } from 'lodash'

import type { Token } from '@sourcegraph/shared/src/search/query/token'

import { getQueryInformation } from '../../codemirror/parsedQuery'
import {
    type Action,
    type Option,
    startCompletion,
    suggestionSources,
    selectionListener,
    RenderAs,
} from '../suggestionsExtension'

export interface Example {
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

export function exampleSuggestions(config: {
    getUsedExamples: () => Set<string>
    markExampleUsed: (example: string) => void
    examples: Example[]
}): Extension {
    const { getUsedExamples, markExampleUsed, examples } = config

    const choosenExamples = {
        examples: sampleSize(examples, EXAMPLES_TO_SHOW),
        timestamp: Date.now(),
    }

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

                const usedExamples = getUsedExamples()
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
                                    render: RenderAs.QUERY,
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
                markExampleUsed(option.label)
            }
        }),
    ]
}
