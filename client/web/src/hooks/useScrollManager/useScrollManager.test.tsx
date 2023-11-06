import { useRef } from 'react'

import { jest, afterAll, beforeAll, describe, it } from '@jest/globals'
import { act, fireEvent, screen } from '@testing-library/react'
import { Routes, Route } from 'react-router-dom'
import { spy, assert } from 'sinon'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

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
    <Routes>
        <Route path="/page-1" element={<TestPage id="page-1" />} />
        <Route path="/page-2" element={<TestPage id="page-2" />} />
    </Routes>
)

describe('useScrollManager', () => {
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
            wrapper.navigateRef.current?.('/page-1')
        })
        // Called when pushing history
        assert.callCount(scrollToMock, 1)
        assert.calledWith(scrollToMock, 0, 0)

        const pageOneContainer = screen.getByTestId('page-1')
        const PAGE_ONE_SCROLL_POSITION = 100
        fireEvent.scroll(pageOneContainer, { target: { scrollTop: PAGE_ONE_SCROLL_POSITION } })
        jest.advanceTimersByTime(250) // Wait over 200ms for scroll position to be saved by scroll manager

        // Navigate to other page
        act(() => {
            wrapper.navigateRef.current?.('/page-2')
        })
        // Called when pushing history
        assert.callCount(scrollToMock, 2)
        assert.calledWith(scrollToMock, 0, 0)

        const pageTwoContainer = screen.getByTestId('page-2')
        const PAGE_TWO_SCROLL_POSITION = 300
        fireEvent.scroll(pageTwoContainer, { target: { scrollTop: PAGE_TWO_SCROLL_POSITION } })
        jest.advanceTimersByTime(250) // Wait over 200ms for scroll position to be saved by scroll manager

        // Navigate backwards to first page
        act(() => {
            wrapper.navigateRef.current?.(-1)
        })
        // Check that we attempt to scroll back to the correct position
        assert.callCount(scrollToMock, 3)
        assert.calledWith(scrollToMock, 0, PAGE_ONE_SCROLL_POSITION)

        // Navigate forwards to second page
        act(() => {
            wrapper.navigateRef.current?.(1)
        })
        // Check that we attempt to scroll back to the correct position
        assert.callCount(scrollToMock, 4)
        assert.calledWith(scrollToMock, 0, PAGE_TWO_SCROLL_POSITION)

        // Navigate directly to the first page
        act(() => {
            wrapper.navigateRef.current?.('/page-1')
        })
        // Check that we attempted to scroll back to top on push
        assert.callCount(scrollToMock, 5)
        assert.calledWith(scrollToMock, 0, 0)

        // Replace history with second page
        act(() => {
            wrapper.navigateRef.current?.('/page-2', { replace: true })
        })
        // Check that we did not attempt to scroll on history replace
        assert.callCount(scrollToMock, 5)
    })
})
