import React from 'react'
import { LinkOrButton } from './LinkOrButton'
import { mount } from 'enzyme'

describe('LinkOrButton', () => {
    test('render a link when "to" is set', () => {
        expect(mount(<LinkOrButton to="http://example.com">foo</LinkOrButton>)).toMatchSnapshot()
    })

    test('render a button when "to" is undefined', () => {
        expect(mount(<LinkOrButton to={undefined}>foo</LinkOrButton>)).toMatchSnapshot()
    })
})
