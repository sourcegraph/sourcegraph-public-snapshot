import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { H4 } from '../Typography'

import { Alert } from './Alert'
import { ALERT_VARIANTS } from './constants'

describe('Alert', () => {
    it('renders a simple Alert correctly', () => {
        const { container } = render(<Alert>Simple Alert</Alert>)
        expect(container.firstChild).toMatchInlineSnapshot(`
            <div
              aria-live="polite"
              class=""
              role="alert"
            >
              Simple Alert
            </div>
        `)
    })

    it.each(ALERT_VARIANTS)("renders variant '%s' correctly", variant => {
        const { container } = render(
            <Alert variant={variant}>
                <H4>Too many matching repositories</H4>
                Use a 'repo:' filter to narrow your search.
            </Alert>
        )
        expect(container.firstChild).toMatchSnapshot()
    })

    it('renders Alert content correctly', () => {
        const { container } = render(<Alert>Too many matching repositories</Alert>)
        expect(container.firstChild).toHaveTextContent('Too many matching repositories')
    })
})
