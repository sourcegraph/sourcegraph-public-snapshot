import { render } from '@testing-library/react'
import React from 'react'
import { BrowserRouter } from 'react-router-dom'

import { RouterLink } from './RouterLink'

describe('RouterLink', () => {
    it('renders router link correctly', () => {
        const { asFragment } = render(
            <BrowserRouter>
                <RouterLink to="/"> Link to docs </RouterLink>
            </BrowserRouter>
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
