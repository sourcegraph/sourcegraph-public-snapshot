import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { AnchorLink } from './AnchorLink'

describe('AnchorLink', () => {
    it('renders anchor link correctly', () => {
        const { asFragment } = render(<AnchorLink to="#"> Link to docs </AnchorLink>)
        expect(asFragment()).toMatchSnapshot()
    })
})
