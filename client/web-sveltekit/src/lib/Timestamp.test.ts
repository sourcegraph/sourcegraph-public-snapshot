// @vitest-environment jsdom

import { faker } from '@faker-js/faker'
import { render, screen } from '@testing-library/svelte'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'svelte'
import { describe, test, expect, vi } from 'vitest'

import { useFakeTimers, useRealTimers } from '$mocks'

import Timestamp from './Timestamp.svelte'

describe('Timestamp.svelte', () => {
    function renderTimestamp(options?: Partial<ComponentProps<Timestamp>>) {
        const date = faker.date.recent()
        render(Timestamp, { date, ...options })
    }

    test('show tooltip when hovering', async () => {
        const user = userEvent.setup()
        renderTimestamp()
        await user.hover(screen.getByTestId('timestamp'))
        const tooltip = await screen.findByRole('tooltip')
        expect(tooltip.textContent).toMatchInlineSnapshot('"2021-05-23 12:57:34 PM "')
    })

    test('automatically updates as time passes', async () => {
        useFakeTimers()

        renderTimestamp({ date: faker.defaultRefDate() })
        const element = screen.getByTestId('timestamp')
        expect(element.textContent).toMatchInlineSnapshot('"less than a minute ago"')

        // Advance timer by 42 minutes
        await vi.advanceTimersByTimeAsync(42 * 60 * 1000)
        expect(element.textContent).toMatchInlineSnapshot('"42 minutes ago"')

        useRealTimers()
    })

    test.each([{}, { addSuffix: false }, { strict: true }, { addSuffix: false, strict: true }, { showAbsolute: true }])(
        'props: %o',
        (_, options) => {
            useFakeTimers()

            renderTimestamp(options)
            expect(screen.getByTestId('timestamp').textContent).toMatchSnapshot()

            useRealTimers()
        }
    )
})
