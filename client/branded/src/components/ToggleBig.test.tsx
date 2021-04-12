import { mount } from 'enzyme'
import React from 'react'
import sinon from 'sinon'

import { ToggleBig } from './ToggleBig'

describe('ToggleBig', () => {
    test('value is false', () => {
        expect(mount(<ToggleBig value={false} />)).toMatchSnapshot()
    })

    test('value is true', () => {
        expect(mount(<ToggleBig value={true} />)).toMatchSnapshot()
    })

    test('disabled', () => {
        const onToggle = sinon.spy(() => undefined)
        const component = mount(<ToggleBig onToggle={onToggle} disabled={true} />)

        component.find('.toggle-big').simulate('click')
        sinon.assert.notCalled(onToggle)
        expect(component).toMatchSnapshot()
    })

    test('className', () => expect(mount(<ToggleBig className="c" />)).toMatchSnapshot())
})
