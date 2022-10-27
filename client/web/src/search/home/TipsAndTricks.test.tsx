import { render, cleanup, RenderResult } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { TipsAndTricks } from './TipsAndTricks'

const mockLogFunc = sinon.spy()
const mockExampleLink = {
    label: 'Negation:',
    trackEventName: 'HomepageExampleNegationClicked',
    query: '-file:tests',
    to: '/search?q=context:global+r:tests+-file:tests+-file:%28%5E%7C/%29vendor/+auth%28&patternType=literal',
}
const mockMoreLink = {
    label: 'More search features',
    href: 'https://docs.sourcegraph.com/code_search/explanations/features',
    trackEventName: 'HomepageExampleMoreSearchFeaturesClicked',
}
const renderTipsAndTricks = (): RenderResult =>
    render(
        <MemoryRouter initialEntries={['/']}>
            <TipsAndTricks
                examples={[mockExampleLink]}
                moreLink={mockMoreLink}
                telemetryService={{ ...NOOP_TELEMETRY_SERVICE, log: mockLogFunc }}
            />
        </MemoryRouter>
    )

describe('TipsAndTricks.tsx', () => {
    afterAll(cleanup)

    beforeEach(() => {
        mockLogFunc.resetHistory()
    })

    test('triggers event log on example link click', () => {
        const { getByText } = renderTipsAndTricks()
        getByText(mockExampleLink.label).querySelector('a')?.click()
        expect(mockLogFunc.withArgs(mockExampleLink.trackEventName).calledOnce).toBeTruthy()
    })

    test('triggers event log on more link click', () => {
        const { getByText } = renderTipsAndTricks()
        getByText(mockMoreLink.label).click()
        expect(mockLogFunc.withArgs(mockMoreLink.trackEventName).calledOnce).toBeTruthy()
    })
})
