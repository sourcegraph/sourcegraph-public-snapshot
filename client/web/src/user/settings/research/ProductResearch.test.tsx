import React from 'react'
import { render, RenderResult } from '@testing-library/react'
import * as sinon from 'sinon'
import { ProductResearchPage } from './ProductResearch'
import { AuthenticatedUser } from '../../../auth'
import { TelemetryService } from '../../../../../shared/src/telemetry/telemetryService'

describe('ProductResearchPage', () => {
    let queries: RenderResult
    const mockTelemetryService = { logViewEvent: sinon.spy(), log: sinon.spy() } as TelemetryService
    const mockAuthenticatedUser = { email: 'test@sourcegraph.com' } as AuthenticatedUser

    beforeEach(() => {
        queries = render(
            <ProductResearchPage telemetryService={mockTelemetryService} authenticatedUser={mockAuthenticatedUser} />
        )
    })

    test('Renders page correctly', () => {
        expect(queries.getByText('Product research and feedback')).toBeVisible()
    })

    test('renders sign up now link correctly', () => {
        expect(queries.getByText('Sign up now').closest('a')?.href).toMatchInlineSnapshot(
            '"https://share.hsforms.com/1tkScUc65Tm-Yu98zUZcLGw1n7ku?email=test@sourcegraph.com"'
        )
    })
})
