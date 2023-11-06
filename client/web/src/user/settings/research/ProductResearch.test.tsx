import { beforeEach, describe, expect, test } from '@jest/globals'
import { render, type RenderResult } from '@testing-library/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ProductResearchPage } from './ProductResearch'

describe('ProductResearchPage', () => {
    let queries: RenderResult

    beforeEach(() => {
        queries = render(
            <ProductResearchPage
                isCodyApp={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={{
                    emails: [{ email: 'test@sourcegraph.com', isPrimary: true, verified: true }],
                }}
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
