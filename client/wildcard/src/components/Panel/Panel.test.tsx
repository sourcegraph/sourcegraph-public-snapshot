import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { Panel } from './Panel'

describe('Panel', () => {
    it('renders correctly positioned at page bottom', () => {
        expect(render(<Panel ariaLabel="Test panel" />).asFragment()).toMatchSnapshot()
    })
})
