import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Panel } from './Panel'

describe('Panel', () => {
    it('renders correctly positioned at page bottom', () => {
        expect(render(<Panel ariaLabel="Test panel" />).asFragment()).toMatchSnapshot()
    })
})
