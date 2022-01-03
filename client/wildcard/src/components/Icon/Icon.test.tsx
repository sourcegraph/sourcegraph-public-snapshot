import { render } from '@testing-library/react'
import React from 'react'

import { SourcegraphIcon } from '../SourcegraphIcon'

import { Icon } from './Icon'

describe('Icon', () => {
    it('renders a simple icon correctly', () => {
        const { asFragment } = render(<Icon svg={<SourcegraphIcon />} />)
        expect(asFragment()).toMatchSnapshot()
    })
})
