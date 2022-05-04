import { render, RenderResult, cleanup, fireEvent, waitFor, screen } from '@testing-library/react'

import { Tooltip } from './Tooltip'

const TooltipTest = () => (
    <>
        <Tooltip />
        <div>
            Hover on{' '}
            <strong data-testid="hoverable-1" data-tooltip="Tooltip 1">
                me
            </strong>
            , or{' '}
            <strong data-testid="hoverable-2" data-tooltip="Tooltip 2">
                me
            </strong>
        </div>
    </>
)

describe('Tooltip', () => {
    let queries: RenderResult

    afterEach(cleanup)

    describe('Hoverable Tooltip', () => {
        beforeEach(() => {
            queries = render(<TooltipTest />)
        })

        it('Shows tooltip properly on hover', async () => {
            fireEvent.mouseOver(queries.getByTestId('hoverable-1'))

            await waitFor(() => {
                expect(screen.getByRole('tooltip')).toHaveTextContent('Tooltip 1')
                expect(screen.getByRole('tooltip').closest('.show.fade')).toBeInTheDocument()
            })

            expect(document.body).toMatchSnapshot()
        })

        it('Handles multiple tooltips properly', async () => {
            fireEvent.mouseOver(queries.getByTestId('hoverable-1'))

            await waitFor(() => {
                expect(screen.getByRole('tooltip')).toHaveTextContent('Tooltip 1')
                expect(screen.getByRole('tooltip').closest('.show.fade')).toBeInTheDocument()
            })

            fireEvent.mouseOver(queries.getByTestId('hoverable-2'))

            await waitFor(() => {
                expect(screen.getByRole('tooltip')).toHaveTextContent('Tooltip 2')
                expect(screen.getByRole('tooltip').closest('.show.fade')).toBeInTheDocument()
            })

            expect(document.body).toMatchSnapshot()
        })
    })
})
