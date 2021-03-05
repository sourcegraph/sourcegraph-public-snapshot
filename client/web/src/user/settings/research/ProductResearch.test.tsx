import React from 'react'
import { render, RenderResult } from '@testing-library/react'
import { ProductResearchPage } from './ProductResearch'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'

describe('ProductResearchPage', () => {
    let queries: RenderResult

    beforeEach(() => {
        queries = render(
            <ProductResearchPage
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
            '"https://info.sourcegraph.com/product-research?email=test%40sourcegraph.com"'
        )
    })
})
