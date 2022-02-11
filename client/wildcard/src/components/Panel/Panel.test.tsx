import { render } from '@testing-library/react'
import React from 'react'

import { Panel } from './Panel'

describe('Panel', () => {
    it('renders correctly positioned at page bottom', () => {
        expect(render(<Panel />).asFragment()).toMatchSnapshot()
    })
})
