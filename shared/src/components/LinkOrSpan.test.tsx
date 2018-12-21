import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from './Link'
import { LinkOrSpan } from './LinkOrSpan'

describe('LinkOrSpan', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('render a link when "to" is set', () => {
        const component = renderer.create(<LinkOrSpan to="http://example.com">foo</LinkOrSpan>)
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('render a span when "to" is undefined', () => {
        const component = renderer.create(<LinkOrSpan to={undefined}>foo</LinkOrSpan>)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
