import React from 'react'
import renderer from 'react-test-renderer'
import { LinkOrButton } from './LinkOrButton'

describe('LinkOrButton', () => {
    test('render a link when "to" is set', () => {
        const component = renderer.create(<LinkOrButton to="http://example.com">foo</LinkOrButton>)
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('render a button when "to" is undefined', () => {
        const component = renderer.create(<LinkOrButton to={undefined}>foo</LinkOrButton>)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
