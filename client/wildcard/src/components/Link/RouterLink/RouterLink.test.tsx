import React from 'react'

import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { RouterLink } from './RouterLink'

describe('RouterLink', () => {
    it('renders router link correctly', () => {
        const { asFragment } = renderWithRouter(<RouterLink to="/docs">Link to docs</RouterLink>)
        expect(asFragment()).toMatchSnapshot()
    })
})
