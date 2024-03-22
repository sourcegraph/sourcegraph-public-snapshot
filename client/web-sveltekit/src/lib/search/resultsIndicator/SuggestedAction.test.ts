// @vitest-environment jsdom

import { render } from '@testing-library/svelte'
import type { ComponentProps } from 'svelte'
import { describe, test, expect } from 'vitest'

import SuggestedAction from '$lib/search/resultsIndicator/SuggestedAction.svelte'

describe('SuggestedAction.svelte', () => {
    function renderSuggestedAction(options?: Partial<ComponentProps<SuggestedAction>>): void {
        render(SuggestedAction, { ...options })
    }

    test('renders action with title and suggestion', async () => {
        renderSuggestedAction({
            hasSkippedItems: true,
            hasSuggestedItems: true,
            progress: {
                done: true,
                matchCount: 600,
                durationMs: 5260,
                skipped: [
                    {
                        reason: 'error',
                        title: 'this is an error',
                        message: 'vv much an error',
                        severity: 'error',
                        suggested: {
                            title: "here's a title",
                            message: "here's a message",
                        },
                    },
                ],
            },
            isError: false,
            mostSevere: {
                reason: 'info',
                title: 'info here',
                message: 'vv much an info',
                severity: 'info',
                suggested: {
                    title: "here's a title",
                    message: "here's a message",
                },
            },
        })

        const title = document.getElementsByClassName('info-badge')
        expect(title).toHaveLength(1)

        const interpunct = document.getElementsByClassName('separator')
        expect(interpunct).toHaveLength(1)

        const action = document.getElementsByClassName('action-badge')
        expect(action).toHaveLength(1)
    })

    test('renders title only when there is no suggested items', async () => {
        renderSuggestedAction({
            hasSkippedItems: true,
            hasSuggestedItems: false,
            progress: {
                done: true,
                matchCount: 600,
                durationMs: 5260,
                skipped: [
                    {
                        reason: 'error',
                        title: 'this is an error',
                        message: 'vv much an error',
                        severity: 'error',
                    },
                ],
            },
            isError: false,
            mostSevere: {
                reason: 'info',
                title: 'info here',
                message: 'vv much an info',
                severity: 'info',
            },
        })

        const title = document.getElementsByClassName('info-badge')
        expect(title).toHaveLength(1)

        const interpunct = document.getElementsByClassName('separator')
        expect(interpunct).toHaveLength(0)

        const action = document.getElementsByClassName('action-badge')
        expect(action).toHaveLength(0)
    })
})
