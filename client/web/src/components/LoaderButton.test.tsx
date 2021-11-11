import { render } from '@testing-library/react'
import React from 'react'

import { LoaderButton } from './LoaderButton'

describe('LoaderButton', () => {
    it('should render a loading spinner when loading prop is true', () => {
        expect(render(<LoaderButton label="Test" loading={true} />).asFragment()).toMatchSnapshot()
    })

    it('should not render a loading spinner when loading prop is false', () => {
        expect(render(<LoaderButton label="Test" loading={false} />).asFragment()).toMatchSnapshot()
    })
})
