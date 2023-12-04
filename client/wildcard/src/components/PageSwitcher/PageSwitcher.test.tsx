import { render, type RenderResult, cleanup, fireEvent } from '@testing-library/react'
import sinon from 'sinon'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { assertAriaDisabled, assertAriaEnabled } from '@sourcegraph/testing'

import { PageSwitcher, type PageSwitcherProps } from './PageSwitcher'

describe('PageSwitcher', () => {
    const renderWithProps = (props: PageSwitcherProps): RenderResult => render(<PageSwitcher {...props} />)

    const goToNextPage = sinon.spy()
    const goToPreviousPage = sinon.spy()
    const goToFirstPage = sinon.spy()
    const goToLastPage = sinon.spy()

    let defaultProps: PageSwitcherProps

    beforeEach(() => {
        goToNextPage.resetHistory()
        goToPreviousPage.resetHistory()
        goToFirstPage.resetHistory()
        goToLastPage.resetHistory()

        defaultProps = {
            totalLabel: 'elements',
            totalCount: 10,
            goToNextPage,
            goToPreviousPage,
            goToFirstPage,
            goToLastPage,
            hasNextPage: true,
            hasPreviousPage: true,
        }
    })
    afterEach(cleanup)

    it('renders the go to first page button', () => {
        const queries = renderWithProps(defaultProps)
        const goToFirstPageButton = queries.getByRole('button', { name: 'Go to first page' })
        expect(goToFirstPageButton).toBeInTheDocument()
        assertAriaEnabled(goToFirstPageButton)
        fireEvent.click(goToFirstPageButton)
        sinon.assert.calledOnce(goToFirstPage)
    })

    it('renders the go to next page button', () => {
        const queries = renderWithProps(defaultProps)
        const goToNextPageButton = queries.getByRole('button', { name: 'Go to next page' })
        expect(goToNextPageButton).toBeInTheDocument()
        assertAriaEnabled(goToNextPageButton)
        fireEvent.click(goToNextPageButton)
        sinon.assert.calledOnce(goToNextPage)
    })

    it('renders the go to previous page button', () => {
        const queries = renderWithProps(defaultProps)
        const goToPreviousPageButton = queries.getByRole('button', { name: 'Go to previous page' })
        expect(goToPreviousPageButton).toBeInTheDocument()
        assertAriaEnabled(goToPreviousPageButton)
        fireEvent.click(goToPreviousPageButton)
        sinon.assert.calledOnce(goToPreviousPage)
    })

    it('renders the go to last page button', () => {
        const queries = renderWithProps(defaultProps)
        const goToLastPageButton = queries.getByRole('button', { name: 'Go to last page' })
        expect(goToLastPageButton).toBeInTheDocument()
        assertAriaEnabled(goToLastPageButton)
        fireEvent.click(goToLastPageButton)
        sinon.assert.calledOnce(goToLastPage)
    })

    it('disables the go to first page and go to previous page button when it has no previous page', () => {
        const queries = renderWithProps({ ...defaultProps, hasPreviousPage: false })
        const goToFirstPageButton = queries.getByRole('button', { name: 'Go to first page' })
        const goToPreviousPageButton = queries.getByRole('button', { name: 'Go to previous page' })
        expect(goToFirstPageButton).toBeInTheDocument()
        expect(goToPreviousPageButton).toBeInTheDocument()
        assertAriaDisabled(goToFirstPageButton)
        assertAriaDisabled(goToPreviousPageButton)
        fireEvent.click(goToFirstPageButton)
        fireEvent.click(goToPreviousPageButton)
        sinon.assert.notCalled(goToFirstPage)
        sinon.assert.notCalled(goToPreviousPage)
    })

    it('disables the go to last page and go to next page button when it has no next page', () => {
        const queries = renderWithProps({ ...defaultProps, hasNextPage: false })
        const goToLastPageButton = queries.getByRole('button', { name: 'Go to last page' })
        const goToNextPageButton = queries.getByRole('button', { name: 'Go to next page' })
        expect(goToLastPageButton).toBeInTheDocument()
        expect(goToNextPageButton).toBeInTheDocument()
        assertAriaDisabled(goToLastPageButton)
        assertAriaDisabled(goToNextPageButton)
        fireEvent.click(goToLastPageButton)
        fireEvent.click(goToNextPageButton)
        sinon.assert.notCalled(goToLastPage)
        sinon.assert.notCalled(goToNextPage)
    })

    it('renders label with total count', () => {
        const queries = renderWithProps(defaultProps)
        expect(queries.container.textContent!).toContain('Total elements: 10')
    })

    it("doesn't render the label when it's not set", () => {
        const queries = renderWithProps({ ...defaultProps, totalLabel: undefined })
        expect(queries.container.textContent!).not.toContain('Total')
    })
})
