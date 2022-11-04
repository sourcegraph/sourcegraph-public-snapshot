import { useRef } from 'react'

import { act, fireEvent, screen } from '@testing-library/react'
import { Route } from 'react-router'
import { spy, assert } from 'sinon'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { useScrollManager } from './useScrollManager'

const TestPage = ({ id }: { id: string }) => {
    const divRef = useRef(null)
    useScrollManager('test-page-scroll', divRef)

    return (
        <div ref={divRef} data-testid={id}>
            {id}
        </div>
    )
}

const TestApp = () => (
    <main>
        <Route path="/page-1" render={() => <TestPage id="page-1" />} />
        <Route path="/page-2" render={() => <TestPage id="page-2" />} />
    </main>
)

describe('useScrollManager', () => {
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalScrollTo = Element.prototype.scrollTo
    const scrollToMock = spy()
    beforeAll(() => {
        // scrollTo is not supported in JSDOM, so we mock it for this one test
        // https://github.com/jsdom/jsdom/issues/1422
        Element.prototype.scrollTo = scrollToMock
    })
    afterAll(() => {
        Element.prototype.scrollTo = originalScrollTo
    })

    it('handles scroll correctly', () => {
        jest.useFakeTimers()

        const wrapper = renderWithBrandedContext(<TestApp />)
        act(() => {
            wrapper.history.push('/page-1')
        })

        const pageOneContainer = screen.getByTestId('page-1')
        const PAGE_ONE_SCROLL_POSITION = 100
        fireEvent.scroll(pageOneContainer, { target: { scrollTop: PAGE_ONE_SCROLL_POSITION } })
        jest.advanceTimersByTime(250) // Wait over 200ms for scroll position to be saved by scroll manager

        // Navigate to other page
        act(() => {
            wrapper.history.push('/page-2')
        })

        const pageTwoContainer = screen.getByTestId('page-2')
        const PAGE_TWO_SCROLL_POSITION = 300
        fireEvent.scroll(pageTwoContainer, { target: { scrollTop: PAGE_TWO_SCROLL_POSITION } })
        jest.advanceTimersByTime(250) // Wait over 200ms for scroll position to be saved by scroll manager

        // Navigate backwards to first page
        act(() => {
            wrapper.history.goBack()
        })
        // Check that we attempt to scroll back to the correct position
        assert.callCount(scrollToMock, 1)
        assert.calledWith(scrollToMock, 0, PAGE_ONE_SCROLL_POSITION)

        // Navigate forwards to second page
        act(() => {
            wrapper.history.goForward()
        })
        // Check that we attempt to scroll back to the correct position
        assert.callCount(scrollToMock, 2)
        assert.calledWith(scrollToMock, 0, PAGE_TWO_SCROLL_POSITION)

        // Navigate directly to the first page
        act(() => {
            wrapper.history.push('/page-1')
        })
        // Check that we have not attempted to scroll (callCount is still 2)
        assert.callCount(scrollToMock, 2)
    })
})
