// @vitest-environment jsdom
import InfoBadge from '$lib/search/resultsIndicator/InfoBadge.svelte'
import { render } from '@testing-library/svelte'
import type { ComponentProps } from 'svelte'
import { describe, test, expect } from 'vitest'

describe("InfoBadge.svelte", () => {
    function renderInfoBadge(options?: Partial<ComponentProps<InfoBadge>>): void {
        render(InfoBadge, { ...options })
    }

    test('render as error if severity === error', async () => {
        renderInfoBadge({
            searchProgress: {
                matchCount: 0,
                durationMs: 3800,
                skipped: [
                    {
                        reason: 'error',
                        title: 'this is an error',
                        message: 'vv much an error',
                        severity: 'error'
                    }
                ]
            },
            state: 'error',
        })

        const errorClass = document.getElementsByClassName('error-text')
        expect(errorClass).toHaveLength(1)

        const errorText = errorClass[0].textContent
        expect(errorText).toBe("0 results in 3.80s")
    })

    test('render as info if severity === info', () => {
        renderInfoBadge({
            searchProgress: {
                matchCount: 433,
                durationMs: 5600,
                skipped: [
                    {
                        reason: 'info',
                        title: 'this is info',
                        message: 'vv much info',
                        severity: 'info'
                    }
                ]
            },
            state: 'info',
        })

        const infoClass = document.getElementsByClassName('error-text')
        expect(infoClass).toHaveLength(0)
    })
})
