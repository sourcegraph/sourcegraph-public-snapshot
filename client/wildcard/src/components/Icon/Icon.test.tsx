import React from 'react'

import { render } from '@testing-library/react'

import { SourcegraphIcon } from '../SourcegraphIcon'

import { Icon } from './Icon'

describe('Icon', () => {
    it('renders a simple icon correctly', () => {
        const { asFragment } = render(<Icon as={SourcegraphIcon} />)
        expect(asFragment()).toMatchSnapshot()
    })
})
