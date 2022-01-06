import { render } from '@testing-library/react'
import React from 'react'

import { Alert } from './Alert'
import { ALERT_VARIANTS } from './constants'

describe('Alert', () => {
    it('renders a simple Alert correctly', () => {
        const { container } = render(<Alert>Simple Alert</Alert>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <div
              class="alert"
            >
              Simple Alert
            </div>
        `)
    })

    it.each(ALERT_VARIANTS)("renders variant '%s' correctly", variant => {
        const { container } = render(
            <Alert variant={variant}>
                <h4>Too many matching repositories</h4>
                Use a 'repo:' or 'repogroup:' filter to narrow your search.
            </Alert>
        )
        expect(container.firstChild).toMatchSnapshot()
    })

    it('renders Alert content correctly', () => {
        const { container } = render(<Alert>Too many matching repositories</Alert>)
        expect(container.firstChild).toHaveTextContent('Too many matching repositories')
    })
})
