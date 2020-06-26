import React from 'react'
import { mount } from 'enzyme'
import { HiddenPatchNode } from './HiddenPatchNode'

describe('HiddenPatchNode', () => {
    test('renders', () => {
        expect(mount(<HiddenPatchNode />).children()).toMatchSnapshot()
    })
})
