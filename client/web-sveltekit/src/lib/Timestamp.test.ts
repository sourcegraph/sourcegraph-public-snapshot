// @vitest-environment jsdom

import { faker } from '@faker-js/faker'
import { render, screen } from '@testing-library/svelte'
import type { ComponentProps } from 'svelte'
import { describe, test, expect, vi } from 'vitest'

import { useFakeTimers, useRealTimers } from '$testing/mocks'

import Timestamp from './Timestamp.svelte'

describe('Timestamp.svelte', () => {
    function renderTimestamp(options?: Partial<ComponentProps<Timestamp>>): void {
        const date = faker.date.recent()
        render(Timestamp, { date, ...options })
    }

    test('automatically updates as time passes', async () => {
        useFakeTimers()

        renderTimestamp({ date: faker.defaultRefDate() })
        const element = screen.getByTestId('timestamp')
        expect(element.textContent).toMatchInlineSnapshot('"less than a minute ago"')

        // Advance timer by 9 minutes
        await vi.advanceTimersByTimeAsync(9 * 60 * 1000)
        expect(element.textContent).toMatchInlineSnapshot('"9 minutes ago"')

        useRealTimers()
    })

    test.each([{}, { hideSuffix: true }, { strict: true }, { hideSuffix: true, strict: true }, { showAbsolute: true }])(
        'props: %o',
        options => {
            useFakeTimers()

            renderTimestamp(options)
            expect(screen.getByTestId('timestamp').textContent).toMatchSnapshot()

            useRealTimers()
        }
    )
})
