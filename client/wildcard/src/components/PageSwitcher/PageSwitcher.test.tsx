import { render, RenderResult, cleanup, fireEvent } from '@testing-library/react'
import sinon from 'sinon'

import { PageSwitcher, PageSwitcherProps } from './PageSwitcher'

describe('PageSwitcher', () => {
    let queries: RenderResult
    const renderWithProps = (props: PageSwitcherProps): RenderResult => render(<PageSwitcher {...props} />)

    const goToNextPage = sinon.spy()
    const goToPreviousPage = sinon.spy()
    const goToFirstPage = sinon.spy()
    const goToLastPage = sinon.spy()

    beforeEach(() => {
        goToNextPage.resetHistory()
        goToPreviousPage.resetHistory()
        goToFirstPage.resetHistory()
        goToLastPage.resetHistory()
    })
    afterEach(cleanup)

    describe('Simple pagination', () => {
        beforeEach(() => {
            queries = renderWithProps({
                totalLabel: 'elements',
                totalCount: 10,
                goToNextPage,
                goToPreviousPage,
                goToFirstPage,
                goToLastPage,
                hasNextPage: true,
                hasPreviousPage: false,
            })
        })

        it('will render four buttons', () => {
            expect(queries.getAllByRole('button')).toHaveLength(4)
        })
    })
})
