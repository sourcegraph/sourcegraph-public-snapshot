import React from 'react'
import { render, RenderResult } from '@testing-library/react'
import { ProductResearchArea } from './ProductResearch'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'

describe('ProductResearchArea', () => {
    let queries: RenderResult

    beforeEach(() => {
        queries = render(
            <ProductResearchArea
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={{ email: 'test@sourcegraph.com' }}
            />
        )
    })

    test('Renders page correctly', () => {
        expect(queries.getByText('Product research and feedback')).toBeVisible()
    })

    test('renders sign up now link correctly', () => {
        expect(queries.getByText('Sign up now').closest('a')?.href).toMatchInlineSnapshot(
            '"https://share.hsforms.com/1tkScUc65Tm-Yu98zUZcLGw1n7ku?email=test%40sourcegraph.com"'
        )
    })
})
