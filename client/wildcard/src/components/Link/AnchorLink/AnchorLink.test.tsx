import { render } from '@testing-library/react'
import React from 'react'

import { AnchorLink } from './AnchorLink'

describe('AnchorLink', () => {
    it('renders anchor link correctly', () => {
        const { asFragment } = render(<AnchorLink to="#"> Link to docs </AnchorLink>)
        expect(asFragment()).toMatchSnapshot()
    })
})
