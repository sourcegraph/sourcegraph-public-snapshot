import { describe, expect, test } from '@jest/globals'
import { render } from '@testing-library/react'

import { LinkOrSpan } from './LinkOrSpan'

describe('LinkOrSpan', () => {
    test('render a link when "to" is set', () => {
        const component = render(<LinkOrSpan to="http://example.com">foo</LinkOrSpan>)
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('render a span when "to" is undefined', () => {
        const component = render(<LinkOrSpan to={undefined}>foo</LinkOrSpan>)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
