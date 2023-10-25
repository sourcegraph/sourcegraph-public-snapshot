import { afterEach, describe, expect, it } from '@jest/globals'
import { cleanup, within, fireEvent } from '@testing-library/react'

import type { RevisionsProps } from '@sourcegraph/branded'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { Revisions } from './Revisions'
import { DEFAULT_MOCKS, MOCK_PROPS } from './Revisions.mocks'

describe('Search Sidebar > Revisions', () => {
    const renderRevisions = (props?: Partial<RevisionsProps>, mocks = DEFAULT_MOCKS): RenderWithBrandedContextResult =>
        renderWithBrandedContext(
            <MockedTestProvider mocks={mocks}>
                <Revisions {...MOCK_PROPS} {...props} />
            </MockedTestProvider>
        )

    afterEach(cleanup)

    describe('Branches', () => {
        async function renderBranchesTab() {
            const result = renderRevisions()
            await waitForNextApolloResponse()
            return result.getByRole('tabpanel', { name: 'Branches' })
        }

        it('renders the correct number of results', async () => {
            const branchTab = await renderBranchesTab()

            expect(within(branchTab).getAllByTestId('filter-link')).toHaveLength(10)
            expect(within(branchTab).getByTestId('summary')).toHaveTextContent('10 of 16 branches')
            expect(within(branchTab).getByText('Show more')).toBeVisible()
        })

        it('fetches remaining branches', async () => {
            const branchTab = await renderBranchesTab()

            fireEvent.click(within(branchTab).getByText('Show more'))
            await waitForNextApolloResponse()
            expect(within(branchTab).getAllByTestId('filter-link')).toHaveLength(16)
            expect(within(branchTab).getByTestId('summary')).toHaveTextContent('16 of 16 branches')
            expect(within(branchTab).queryByText('Show more')).not.toBeInTheDocument()
        })
    })

    describe('Tags', () => {
        async function renderBranchesTab() {
            const result = renderRevisions()
            fireEvent.click(result.getByText('Tags'))
            await waitForNextApolloResponse()
            return result.getByRole('tabpanel', { name: 'Tags' })
        }

        it('renders the correct number of results', async () => {
            const branchTab = await renderBranchesTab()

            expect(within(branchTab).getAllByTestId('filter-link')).toHaveLength(10)
            expect(within(branchTab).getByTestId('summary')).toHaveTextContent('10 of 12 tags')
            expect(within(branchTab).getByText('Show more')).toBeVisible()
        })

        it('fetches remaining tags', async () => {
            const branchTab = await renderBranchesTab()

            fireEvent.click(within(branchTab).getByText('Show more'))
            await waitForNextApolloResponse()
            expect(within(branchTab).getAllByTestId('filter-link')).toHaveLength(12)
            expect(within(branchTab).getByTestId('summary')).toHaveTextContent('12 of 12 tags')
            expect(within(branchTab).queryByText('Show more')).not.toBeInTheDocument()
        })
    })
})
