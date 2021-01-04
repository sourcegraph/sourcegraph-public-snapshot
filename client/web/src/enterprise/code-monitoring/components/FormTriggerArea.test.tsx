import { mount } from 'enzyme'
import React from 'react'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'
import { FormTriggerArea } from './FormTriggerArea'

describe('FormTriggerArea', () => {
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    test('Error message is shown when query is invalid', () => {
        let component = mount(
            <FormTriggerArea
                query="test"
                triggerCompleted={false}
                onQueryChange={sinon.spy()}
                setTriggerCompleted={sinon.spy()}
            />
        )
        act(() => {
            const triggerButton = component.find('.test-trigger-button')
            triggerButton.simulate('click')
            clock.tick(600)
        })
        component = component.update()
        expect(component).toMatchSnapshot()
    })
})
