import { render } from '@testing-library/react'
import React from 'react'

import { ButtonLink } from './LinkOrButton'

describe('LinkOrButton', () => {
    test('render a link when "to" is set', () => {
        const component = render(<ButtonLink to="http://example.com">foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('render a button when "to" is undefined', () => {
        const component = render(<ButtonLink to={undefined}>foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
