import { noop } from 'lodash'
import { createWorkbenchViewScheduler } from './scheduler'
import sinon from 'sinon'

describe('workbenchViewScheduler', () => {
    let requestID = 1
    let flush: FrameRequestCallback | undefined
    const requestAnimationFrame = sinon.spy<(callback: FrameRequestCallback) => number>(callback => {
        flush = callback
        return requestID++
    })
    const scheduler = createWorkbenchViewScheduler(noop, requestAnimationFrame)

    afterAll(() => {
        scheduler.unsubscribe()
    })

    test('createWorkbenchViewScheduler() is idempotent', () => {
        expect(createWorkbenchViewScheduler(noop, requestAnimationFrame)).toBe(scheduler)
    })

    test('only calls requestAnimationFrame() once per frame', () => {
        scheduler.schedule({ type: 'creation', viewType: 'statusBarItem' })
        scheduler.schedule({ type: 'update', viewType: 'statusBarItem' })
        scheduler.schedule({ type: 'update', viewType: 'statusBarItem' })

        sinon.assert.calledOnce(requestAnimationFrame)

        flush?.(1) // browser about to repaint

        scheduler.schedule({ type: 'deletion', viewType: 'statusBarItem' })
        scheduler.schedule({ type: 'creation', viewType: 'statusBarItem' })
        scheduler.schedule({ type: 'update', viewType: 'statusBarItem' })

        sinon.assert.calledTwice(requestAnimationFrame)
    })
})
