// @vitest-environment jsdom

import { faker } from '@faker-js/faker'
import { render, screen } from '@testing-library/svelte'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'svelte'
import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest'

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

    describe('props', () => {
        beforeEach(() => {
            useFakeTimers()
        })

        afterEach(() => {
            useRealTimers()
        })

        test.each([
            ['default options', {}],
            ['no suffix', { addSuffix: false }],
            ['strict', { strict: true }],
            ['no suffix and strict', { addSuffix: false, strict: true }],
            ['absolute time', { showAbsolute: true }],
        ])('%s', (_, options) => {
            renderTimestamp(options)
            expect(screen.getByTestId('timestamp').textContent).toMatchSnapshot()
        })

        test('automatically updates as time passes', async () => {
            renderTimestamp({ date: faker.defaultRefDate() })
            const element = screen.getByTestId('timestamp')
            expect(element.textContent).toMatchInlineSnapshot('"less than a minute ago"')

            await vi.advanceTimersByTimeAsync(42 * 60 * 1000)
            expect(element.textContent).toMatchInlineSnapshot('"42 minutes ago"')
        })
    })
})
