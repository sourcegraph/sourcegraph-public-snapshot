import { cleanup, within, fireEvent, act } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { RevisionsPopover, type RevisionsPopoverProps } from './RevisionsPopover'
import { MOCK_PROPS, MOCK_REQUESTS } from './RevisionsPopover.mocks'

describe('RevisionsPopover', () => {
    let renderResult: RenderWithBrandedContextResult

    const fetchMoreNodes = async (currentTab: HTMLElement) => {
        fireEvent.click(within(currentTab).getByText('Show more'))
        await waitForNextApolloResponse()
    }

    const renderPopover = (props?: Partial<RevisionsPopoverProps>): RenderWithBrandedContextResult =>
        renderWithBrandedContext(
            <MockedTestProvider mocks={MOCK_REQUESTS}>
                <RevisionsPopover {...MOCK_PROPS} {...props} />
            </MockedTestProvider>,
            { route: `/${MOCK_PROPS.repoName}` }
        )

    const waitForInputDebounce = () => act(() => new Promise(resolve => setTimeout(resolve, 200)))

    afterEach(cleanup)

    describe('Branches', () => {
        let branchesTab: HTMLElement

        beforeEach(async () => {
            renderResult = renderPopover()

            fireEvent.click(renderResult.getByText('Branches'))
            await waitForNextApolloResponse()

            branchesTab = renderResult.getByRole('tabpanel', { name: 'Branches' })
        })

        it('renders correct number of results', () => {
            expect(within(branchesTab).getAllByRole('link')).toHaveLength(50)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent(
                '100 branches total (showing first 50)'
            )
            expect(within(branchesTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(branchesTab).getByText('GIT_BRANCH-0-display-name')
            expect(firstNode).toBeVisible()

            const firstLink = firstNode.closest('a')
            expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@GIT_BRANCH-0-abbrev-name`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(branchesTab)
            expect(within(branchesTab).getAllByRole('link')).toHaveLength(100)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent('100 branches total')
            expect(within(branchesTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(branchesTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            await waitForInputDebounce()
            await waitForNextApolloResponse()

            expect(within(branchesTab).getAllByRole('link')).toHaveLength(2)
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent('2 branches matching some query')
        })

        it('displays no results correctly', async () => {
            const searchInput = within(branchesTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some other query' } })

            await waitForInputDebounce()
            await waitForNextApolloResponse()

            expect(within(branchesTab).queryByRole('link')).not.toBeInTheDocument()
            expect(within(branchesTab).getByTestId('summary')).toHaveTextContent(
                'No branches matching some other query'
            )
        })

        describe('Speculative results', () => {
            beforeEach(async () => {
                cleanup()
                renderResult = renderPopover({ showSpeculativeResults: true })

                fireEvent.click(renderResult.getByText('Branches'))
                await waitForNextApolloResponse()

                branchesTab = renderResult.getByRole('tabpanel', { name: 'Branches' })
            })

            it('displays results correctly by displaying a single speculative result', async () => {
                const searchInput = within(branchesTab).getByRole('searchbox')
                fireEvent.change(searchInput, { target: { value: 'some other query' } })

                await waitForInputDebounce()
                await waitForNextApolloResponse()

                expect(within(branchesTab).getByRole('link')).toBeInTheDocument()

                const firstNode = within(branchesTab).getByText('some other query')
                expect(firstNode).toBeVisible()

                const firstLink = firstNode.closest('a')
                expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@some%20other%20query`)
            })
        })
    })

    describe('Tags', () => {
        let tagsTab: HTMLElement

        beforeEach(async () => {
            renderResult = renderPopover()

            fireEvent.click(renderResult.getByText('Tags'))
            await waitForNextApolloResponse()

            tagsTab = renderResult.getByRole('tabpanel', { name: 'Tags' })
        })

        it('renders correct number of results', () => {
            expect(within(tagsTab).getAllByRole('link')).toHaveLength(50)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('100 tags total (showing first 50)')
            expect(within(tagsTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(tagsTab).getByText('GIT_TAG-0-display-name')
            expect(firstNode).toBeVisible()

            const firstLink = firstNode.closest('a')
            expect(firstLink?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@GIT_TAG-0-abbrev-name`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(tagsTab)
            expect(within(tagsTab).getAllByRole('link')).toHaveLength(100)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('100 tags total')
            expect(within(tagsTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(tagsTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            await waitForInputDebounce()
            await waitForNextApolloResponse()

            expect(within(tagsTab).getAllByRole('link')).toHaveLength(2)
            expect(within(tagsTab).getByTestId('summary')).toHaveTextContent('2 tags matching some query')
        })
    })

    describe('Commits', () => {
        let commitsTab: HTMLElement

        beforeEach(async () => {
            renderResult = renderPopover()

            fireEvent.click(renderResult.getByText('Commits'))
            await waitForNextApolloResponse()

            commitsTab = renderResult.getByRole('tabpanel', { name: 'Commits' })
        })

        it('renders correct number of results', () => {
            expect(within(commitsTab).getAllByRole('link')).toHaveLength(15)
            expect(within(commitsTab).getByText('Show more')).toBeVisible()
        })

        it('renders result nodes correctly', () => {
            const firstNode = within(commitsTab).getByText('git-commit-oid-0')
            expect(firstNode).toBeVisible()
            expect(within(commitsTab).getByText('Commit 0: Hello world')).toBeVisible()
            expect(firstNode.closest('a')?.getAttribute('href')).toBe(`/${MOCK_PROPS.repoName}@git-commit-oid-0`)
        })

        it('fetches remaining results correctly', async () => {
            await fetchMoreNodes(commitsTab)
            expect(within(commitsTab).getAllByRole('link')).toHaveLength(30)
            expect(within(commitsTab).queryByText('Show more')).not.toBeInTheDocument()
        })

        it('searches correctly', async () => {
            const searchInput = within(commitsTab).getByRole('searchbox')
            fireEvent.change(searchInput, { target: { value: 'some query' } })

            await waitForInputDebounce()
            await waitForNextApolloResponse()

            expect(within(commitsTab).getAllByRole('link')).toHaveLength(2)
            expect(within(commitsTab).getByTestId('summary')).toHaveTextContent('2 commits matching some query')
        })

        describe('Against a speculative revision', () => {
            beforeEach(async () => {
                cleanup()
                renderResult = renderPopover({ currentRev: 'non-existent-revision' })

                fireEvent.click(renderResult.getByText('Commits'))
                await waitForNextApolloResponse()

                commitsTab = renderResult.getByRole('tabpanel', { name: 'Commits' })
            })

            it('renders 0 results', () => {
                expect(within(commitsTab).queryByRole('link')).not.toBeInTheDocument()
                expect(within(commitsTab).queryByText('Show more')).not.toBeInTheDocument()
                expect(within(commitsTab).getByTestId('summary')).toHaveTextContent('No commits')
            })
        })
    })
})
