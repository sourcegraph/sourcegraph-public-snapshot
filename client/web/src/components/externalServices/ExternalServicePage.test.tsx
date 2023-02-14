import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { subMinutes } from 'date-fns'
import { of } from 'rxjs'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { MockedFeatureFlagsProvider } from '../../featureFlags/MockedFeatureFlagsProvider'
import { ExternalServiceKind, ExternalServiceSyncJobState } from '../../graphql-operations'

import { FETCH_EXTERNAL_SERVICE, queryExternalServiceSyncJobs as _queryExternalServiceSyncJobs } from './backend'
import { ExternalServicePage } from './ExternalServicePage'

const queryExternalServiceSyncJobs: typeof _queryExternalServiceSyncJobs = () =>
    of({
        totalCount: 4,
        pageInfo: { endCursor: null, hasNextPage: false },
        nodes: [
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: null,
                id: 'SYNCJOB1',
                state: ExternalServiceSyncJobState.CANCELING,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: null,
                id: 'SYNCJOB2',
                state: ExternalServiceSyncJobState.PROCESSING,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: 'Very bad error syncing with the code host.',
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: subMinutes(new Date(), 25).toISOString(),
                id: 'SYNCJOB3',
                state: ExternalServiceSyncJobState.FAILED,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
            {
                __typename: 'ExternalServiceSyncJob',
                failureMessage: null,
                startedAt: subMinutes(new Date(), 25).toISOString(),
                finishedAt: subMinutes(new Date(), 25).toISOString(),
                id: 'SYNCJOB4',
                state: ExternalServiceSyncJobState.COMPLETED,
                reposSynced: 5,
                repoSyncErrors: 0,
                reposAdded: 5,
                reposDeleted: 0,
                reposModified: 0,
                reposUnmodified: 0,
            },
        ],
    })

describe('ExternalServicePage', () => {
    afterAll(cleanup)

    describe('async delete ', () => {
        it('does not update the history if searchFragment didnt change', () => {
            const externalServiceId = 'service123'
            const mocks = [
                {
                    request: {
                        query: getDocumentNode(FETCH_EXTERNAL_SERVICE),
                        variables: { id: externalServiceId },
                    },
                    result: {
                        data: {
                            __typename: 'ExternalService' as const,
                            id: externalServiceId,
                            kind: ExternalServiceKind.GITHUB,
                            warning: null,
                            config: '{"githubconfig": true}',
                            displayName: 'GitHub.com',
                            webhookURL: null,
                            lastSyncError: null,
                            repoCount: 1337,
                            lastSyncAt: null,
                            nextSyncAt: null,
                            updatedAt: '2021-03-15T19:39:11Z',
                            createdAt: '2021-03-15T19:39:11Z',
                            hasConnectionCheck: true,
                            namespace: {
                                id: 'userid',
                                namespaceName: 'johndoe',
                                url: '/users/johndoe',
                            },
                        },
                    },
                    nMatches: Number.POSITIVE_INFINITY,
                },
            ]
            render(
                <MockedTestProvider mocks={mocks}>
                    <MockedFeatureFlagsProvider overrides={{ 'async-code-host-delete': true }}>
                        <ExternalServicePage
                            isLightTheme={false}
                            queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                            afterDeleteRoute="/site-admin/after-delete"
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            externalServicesFromFile={false}
                            allowEditExternalServicesWithFile={false}
                        />
                    </MockedFeatureFlagsProvider>
                </MockedTestProvider>
            )

            userEvent.click(screen.getByTestId(`code-host.${externalServiceId}.delete`))

            // act(() => {
            //     // Emulate polling by re-emiting the connection again.
            //     // This should not lead to history being updated
            //     connectionSubject.next(connection)
            //     connectionSubject.next(connection)
            //     connectionSubject.next(connection)
            // })

            // // Click "Show more" button, should cause history to be updated
            // fireEvent.click(screen.getByRole('button')!)
            // expect(currentLocation!.search).toEqual('?foo=bar&first=40')
            // fireEvent.click(screen.getByRole('button')!)
            // expect(currentLocation!.search).toEqual('?foo=bar&first=80')
        })
    })
})

// describe('ConnectionNodes', () => {
//     afterAll(cleanup)

//     it('has a "Show more" button and summary when *not* loading', () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={false}
//             />
//         )
//         expect(screen.getByRole('button')).toHaveTextContent('Show more')
//         expect(screen.getByText('2 cats total')).toBeVisible()
//         expect(screen.getByText('(showing first 1)')).toBeVisible()
//     })

//     it("*doesn't* have a 'Show more' button or summary when loading", () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={true}
//             />
//         )
//         expect(screen.queryByRole('button')).not.toBeInTheDocument()
//         expect(screen.queryByText('2 cats total')).not.toBeInTheDocument()
//         expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()
//         // NOTE: we also expect a LoadingSpinner, but that is not provided by ConnectionNodes.
//     })

//     it("doesn't have a 'Show more' button when noShowMore is true", () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={false}
//                 noShowMore={true}
//             />
//         )
//         expect(screen.queryByRole('button')).not.toBeInTheDocument()
//         expect(screen.getByText('2 cats total')).toBeVisible()
//         expect(screen.getByText('(showing first 1)')).toBeVisible()
//     })

//     it("doesn't have a 'Show more' button or a summary if hasNextPage is false ", () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [{}] })}
//                 loading={false}
//             />
//         )
//         expect(screen.queryByRole('button')).not.toBeInTheDocument()
//         expect(screen.queryByTestId('summary')).not.toBeInTheDocument()
//     })

//     it('calls the onShowMore callback', async () => {
//         const showMoreCallback = sinon.spy(() => undefined)
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={false}
//                 onShowMore={showMoreCallback}
//             />
//         )
//         fireEvent.click(screen.getByRole('button')!)
//         await waitFor(() => sinon.assert.calledOnce(showMoreCallback))
//     })

//     it("doesn't show summary info if totalCount is null", () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: null, nodes: [{}] })}
//                 loading={false}
//             />
//         )
//         expect(screen.queryByTestId('summary')).not.toBeInTheDocument()
//     })

//     it('shows a summary if noSummaryIfAllNodesVisible is false', () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [{}] })}
//                 loading={false}
//                 noSummaryIfAllNodesVisible={false}
//             />
//         )
//         expect(screen.getByText('1 cat total')).toBeVisible()
//         expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()

//         // Summary should come after the nodes.
//         expect(
//             screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
//         ).toEqual(Node.DOCUMENT_POSITION_PRECEDING)
//     })

//     it('shows a summary if nodes.length is 0', () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [] })}
//                 loading={false}
//             />
//         )
//         expect(screen.getByText('1 cat total')).toBeVisible()
//         expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()
//     })

//     it("shows 'No cats' if totalCount is 0", () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: false, totalCount: 0, nodes: [] })}
//                 loading={false}
//             />
//         )
//         expect(screen.getByText('No cats')).toBeVisible()
//     })

//     it('shows the summary at the top if connectionQuery is specified', () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={true}
//                 connectionQuery="meow?"
//             />
//         )
//         // Summary should come _before_ the nodes.
//         expect(
//             screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
//         ).toEqual(Node.DOCUMENT_POSITION_FOLLOWING)
//     })

//     it('shows the summary at the top if connectionQuery is specified', () => {
//         render(
//             <ConnectionNodes
//                 {...defaultConnectionNodesProps}
//                 connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
//                 loading={true}
//                 connectionQuery="meow?"
//             />
//         )
//         // Summary should come _before_ the nodes.
//         expect(
//             screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
//         ).toEqual(Node.DOCUMENT_POSITION_FOLLOWING)
//     })
// })
