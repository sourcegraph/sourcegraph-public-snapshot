import React from 'react'
import sinon from 'sinon'

import { isErrorLike } from '@sourcegraph/common'
import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { RouterLink } from './RouterLink'

describe('RouterLink', () => {
    it('renders router link correctly', () => {
        const { asFragment } = renderWithRouter(<RouterLink to="/docs">Link to docs</RouterLink>)
        expect(asFragment()).toMatchSnapshot()
    })

    it('should throw errors when using RouterLink outside of Router', () => {
        const environmentStub = sinon.stub(process.env, 'NODE_ENV').value('development')

        try {
            renderWithRouter(<RouterLink to="/docs">Link to docs</RouterLink>)
        } catch (error) {
            if (isErrorLike(error)) {
                expect(error.message).toBe('Please use the `AnchorLink` component outside of `react-router`')
            } else {
                throw new Error('Unexpected errors')
            }
        }

        environmentStub.restore()
    })
})
