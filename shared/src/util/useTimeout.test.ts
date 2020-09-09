import { renderHook } from '@testing-library/react-hooks'
import * as sinon from 'sinon'
import { useTimeout } from './useTimeout'

describe('useTimeout()', () => {
    let clock: sinon.SinonFakeTimers
    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    it('should call the callback after specified time elapses', () => {
        const callback = sinon.spy(() => {
            // noop
        })

        const { result } = renderHook(() => useTimeout())
        result.current(callback, 2000)
        sinon.assert.notCalled(callback)
        clock.tick(2000)
        sinon.assert.calledOnce(callback)
    })

    it('should cancel previous timeout on subsequent invocation', () => {
        const callbackOne = sinon.spy(() => {
            // noop
        })
        const callbackTwo = sinon.spy(() => {
            // noop
        })

        const { result } = renderHook(() => useTimeout())
        result.current(callbackOne, 1000)
        clock.tick(500)
        result.current(callbackTwo, 1000)
        clock.tick(500)
        sinon.assert.notCalled(callbackOne)
        sinon.assert.notCalled(callbackTwo)
        clock.tick(500)
        sinon.assert.notCalled(callbackOne)
        sinon.assert.calledOnce(callbackTwo)
    })

    it('should cancel timeout on component unmount', () => {
        const callback = sinon.spy(() => {
            // noop
        })

        const { result, unmount } = renderHook(() => useTimeout())
        result.current(callback, 1000)
        unmount()
        clock.tick(1000)
        sinon.assert.notCalled(callback)
    })
})
