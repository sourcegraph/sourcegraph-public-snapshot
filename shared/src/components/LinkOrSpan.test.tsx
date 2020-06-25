import React from 'react'
import { LinkOrSpan } from './LinkOrSpan'
import { mount } from 'enzyme'

describe('LinkOrSpan', () => {
    test('render a link when "to" is set', () => {
        expect(mount(<LinkOrSpan to="http://example.com">foo</LinkOrSpan>)).toMatchSnapshot()
    })

    test('render a span when "to" is undefined', () => {
        expect(mount(<LinkOrSpan to={undefined}>foo</LinkOrSpan>)).toMatchSnapshot()
    })
})
