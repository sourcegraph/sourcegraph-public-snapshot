import sinon from 'sinon'

import { createDispatcher } from './dispatcher'

describe('dispatcher', () => {
    it('allows stores to register and can dispatch events', () => {
        const dispatcher = createDispatcher()
        const storeA = sinon.spy()
        const storeB = sinon.spy()

        dispatcher.register(storeA)
        dispatcher.register(storeB)

        const event = { type: 'testEvent' }
        dispatcher.dispatch(event)

        sinon.assert.calledWithMatch(storeA, event)
        sinon.assert.calledWithMatch(storeB, event)
    })
})
