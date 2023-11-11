import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { renderWithBrandedContext } from '../../../testing'

import { RouterLink } from './RouterLink'

describe('RouterLink', () => {
    it('renders router link correctly', () => {
        const { asFragment } = renderWithBrandedContext(<RouterLink to="/docs">Link to docs</RouterLink>)
        expect(asFragment()).toMatchSnapshot()
    })
    it('renders absolute URL correctly ', () => {
        const { asFragment } = render(<RouterLink to="https://sourcegraph.com">SourceGraph</RouterLink>)
        expect(asFragment()).toMatchSnapshot()
    })
})
