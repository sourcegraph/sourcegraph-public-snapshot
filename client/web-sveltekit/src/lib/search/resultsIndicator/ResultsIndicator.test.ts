// @vitest-environment jsdom

import { render } from '@testing-library/svelte'
import type { ComponentProps } from 'svelte'
import { describe, test, expect } from 'vitest'

import ResultsIndicator from '$lib/search/resultsIndicator/ResultsIndicator.svelte'

describe('ResultsIndicator.svelte', () => {
    function renderResultsIndicator(options?: Partial<ComponentProps<ResultsIndicator>>): void {
        render(ResultsIndicator, { ...options })
    }

    test('renders correct format for indicator', async () => {
        renderResultsIndicator({
            hasSkippedItems: true,
            sortedItems: [
                {
                    reason: 'display',
                    title: 'Display limit hit',
                    message: 'you hit the display limit',
                    severity: 'info',
                },
            ],
            hasSuggestedItems: true,
            progress: {
                matchCount: 600,
                durationMs: 5260,
                skipped: [],
            },
            state: 'complete',
        })

        const indicator = document.getElementsByClassName('indicator')
        expect(indicator).toHaveLength(1)

        const icon = document.getElementsByClassName('icon')
        expect(icon).toHaveLength(1)

        const messages = document.getElementsByClassName('messages')
        expect(messages).toHaveLength(1)

        const dropdownIcon = document.getElementsByClassName('dropdown-icon')
        expect(dropdownIcon).toHaveLength(1)
    })
})
