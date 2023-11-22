import { render, type RenderResult, cleanup, waitFor, act, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { wait } from '@testing-library/user-event/dist/utils'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { Tooltip } from './Tooltip'

const TooltipTest = () => (
    <>
        Hover on{' '}
        <Tooltip content="Tooltip 1">
            <strong data-testid="trigger-1">me</strong>
        </Tooltip>
        , or{' '}
        <Tooltip content="Tooltip 2">
            <strong data-testid="trigger-2">me</strong>
        </Tooltip>
        , but nothing for{' '}
        <Tooltip content="">
            <strong data-testid="trigger-3">empty string</strong>
        </Tooltip>{' '}
        or{' '}
        <Tooltip content={null}>
            <strong data-testid="trigger-4">null</strong>
        </Tooltip>
    </>
)

describe('Tooltip', () => {
    let rendered: RenderResult

    afterEach(cleanup)

    beforeEach(() => {
        rendered = render(<TooltipTest />)
    })

    it('displays content when the trigger is hovered', async () => {
        userEvent.hover(rendered.getByTestId('trigger-1'))

        await waitFor(() => {
            expect(rendered.getByTestId('trigger-1')).toHaveAttribute('aria-describedby', 'tooltip-1')
            expect(rendered.getByTestId('trigger-2')).not.toHaveAttribute('aria-describedby')

            // Should be one tooltip for visual users, and a second for use with aria-describedby
            const tooltip = rendered.getByRole('tooltip')

            expect(tooltip).toBeInTheDocument()
            expect(tooltip).toHaveTextContent('Tooltip 1')
            expect(tooltip).toHaveAttribute('id', 'tooltip-1')
        })

        fireEvent.pointerLeave(rendered.getByTestId('trigger-1'))
        userEvent.hover(rendered.getByTestId('trigger-2'))

        await waitFor(() => {
            expect(rendered.getByTestId('trigger-1')).not.toHaveAttribute('aria-describedby')
            expect(rendered.getByTestId('trigger-2')).toHaveAttribute('aria-describedby', 'tooltip-2')

            // Should be one tooltip for visual users, and a second for use with aria-describedby
            const tooltip = rendered.getByRole('tooltip')

            expect(tooltip).toBeInTheDocument()
            expect(tooltip).toHaveTextContent('Tooltip 2')
            expect(tooltip).toHaveAttribute('id', 'tooltip-2')
        })
    })

    it('does not display a tooltip on hover for empty content', async () => {
        userEvent.hover(rendered.getByTestId('trigger-3'))
        await act(async () => {
            await wait(100)
        })
        expect(rendered.queryByRole('tooltip')).not.toBeInTheDocument()

        userEvent.hover(rendered.getByTestId('trigger-4'))
        await act(async () => {
            await wait(100)
        })
        expect(rendered.queryByRole('tooltip')).not.toBeInTheDocument()
    })

    it('hides content when the ESC key is pressed', async () => {
        userEvent.hover(rendered.getByTestId('trigger-1'))

        await waitFor(() => {
            expect(rendered.getByRole('tooltip')).toBeInTheDocument()
        })

        // Not sure why `userEvent.type(rendered.getByTestId('trigger-1'), '{esc}')` doesn't work
        await userEvent.type(rendered.getByTestId('trigger-1'), '_{esc}', { delay: 1 })

        await waitFor(() => {
            expect(rendered.queryByRole('tooltip')).not.toBeInTheDocument()
        })
    })

    it('does not hide content when the trigger is clicked', async () => {
        userEvent.hover(rendered.getByTestId('trigger-1'))

        await waitFor(() => {
            expect(rendered.getByRole('tooltip')).toBeInTheDocument()
        })

        userEvent.click(rendered.getByTestId('trigger-1'))

        await waitFor(() => {
            expect(rendered.getByRole('tooltip')).toBeInTheDocument()
        })
    })
})
