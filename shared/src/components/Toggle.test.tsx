import React from 'react'
import { Toggle } from './Toggle'
import { mount } from 'enzyme'

describe('Toggle', () => {
    test('value is false', () => {
        const component = mount(<Toggle value={false} />)
        expect(component).toMatchSnapshot()
    })

    test('value is true', () => {
        const component = mount(<Toggle value={true} />)
        expect(component).toMatchSnapshot()
    })

    test('disabled', () => {
        const component = mount(<Toggle disabled={true} />)
        expect(component).toMatchSnapshot()

        // Clicking while disabled is a noop.
        component.simulate('click')
        expect(component).toMatchSnapshot()
    })

    test('className', () => expect(mount(<Toggle className="c" />)).toMatchSnapshot())
})
