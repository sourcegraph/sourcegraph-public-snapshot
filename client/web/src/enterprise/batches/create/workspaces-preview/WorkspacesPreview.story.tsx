import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { MATCH_ANY_PARAMETERS, WildcardMockedResponse, WildcardMockLink } from 'wildcard-mock-link'

import { BatchSpecWorkspaceResolutionState } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/apollo'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { WORKSPACES, IMPORTING_CHANGESETS, WORKSPACE_RESOLUTION_STATUS } from '../backend'

import { WorkspacesPreview } from './WorkspacesPreview'
import {
    mockWorkspaceResolutionStatus,
    mockBatchSpecWorkspaces,
    mockBatchSpecImportingChangesets,
    mockBatchSpec,
} from './WorkspacesPreview.mock'

const { add } = storiesOf('web/batches/CreateBatchChangePage/WorkspacesPreview', module).addDecorator(story => (
    <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>
))

add('initial', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: {
                    node: {
                        __typename: 'BatchSpec',
                        workspaceResolution: null,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpec={mockBatchSpec()}
                        hasPreviewed={false}
                        previewDisabled={!boolean('Valid batch spec?', true)}
                        preview={noop}
                        batchSpecStale={false}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

const EMPTY_WORKSPACES_MOCK: WildcardMockedResponse = {
    request: {
        query: getDocumentNode(WORKSPACES),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: mockBatchSpecWorkspaces(0),
    },
    nMatches: Number.POSITIVE_INFINITY,
}

const EMPTY_IMPORTING_CHANGESETS_MOCK: WildcardMockedResponse = {
    request: {
        query: getDocumentNode(IMPORTING_CHANGESETS),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: mockBatchSpecImportingChangesets(0),
    },
    nMatches: Number.POSITIVE_INFINITY,
}

add('first preview, error', () => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: mockWorkspaceResolutionStatus(
                    select(
                        'Status',
                        [BatchSpecWorkspaceResolutionState.FAILED, BatchSpecWorkspaceResolutionState.ERRORED],
                        BatchSpecWorkspaceResolutionState.FAILED
                    ),
                    "Uh oh something bad happened and the workspace resolution failed! Here's a long error message with some bullets:\n  * This is a bullet\n  * This is another bullet\n  * This is a third bullet and it's also the most important one so it's longer than all the others wow look at that"
                ),
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        EMPTY_WORKSPACES_MOCK,
        EMPTY_IMPORTING_CHANGESETS_MOCK,
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspacesPreview
                        {...props}
                        batchSpec={mockBatchSpec()}
                        hasPreviewed={true}
                        previewDisabled={false}
                        preview={noop}
                        batchSpecStale={false}
                        excludeRepo={noop}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

// const WORKSPACES_MOCK: WildcardMockedResponse = {
//     request: {
//         query: getDocumentNode(WORKSPACES),
//         variables: MATCH_ANY_PARAMETERS,
//     },
//     result: {
//         data: mockBatchSpecWorkspaces(50),
//     },
//     nMatches: Number.POSITIVE_INFINITY,
// }

// const IMPORTING_CHANGESETS_MOCK: WildcardMockedResponse = {
//     request: {
//         query: getDocumentNode(IMPORTING_CHANGESETS),
//         variables: MATCH_ANY_PARAMETERS,
//     },
//     result: {
//         data: mockBatchSpecImportingChangesets(50),
//     },
//     nMatches: Number.POSITIVE_INFINITY,
// }

// TODO: For some reason the mock connection data for the workspaces is getting messed up
// and becomes undefined after a split second in the component. I can't currently trace
// it, and it doesn't seem reproducible from the actual page where the component is used,
// so I've disabled these for now and will come back to resolve later.
// add('first preview, success', () => {
//     const mocks = new WildcardMockLink([
//         {
//             request: {
//                 query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
//                 variables: MATCH_ANY_PARAMETERS,
//             },
//             result: {
//                 data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED),
//             },
//             nMatches: Number.POSITIVE_INFINITY,
//         },
//         WORKSPACES_MOCK,
//         IMPORTING_CHANGESETS_MOCK,
//     ])

//     return (
//         <WebStory>
//             {props => (
//                 <MockedTestProvider link={mocks}>
//                     <WorkspacesPreview
//                         {...props}
//                         batchSpec={mockBatchSpec()}
//                         hasPreviewed={true}
//                         previewDisabled={false}
//                         preview={noop}
//                         batchSpecStale={false}
//                         excludeRepo={noop}
//                     />
//                 </MockedTestProvider>
//             )}
//         </WebStory>
//     )
// })

// TODO: Disabled for the same reason as the prior one.
// add('first preview, stale', () => {
//     const mocks = new WildcardMockLink([
//         {
//             request: {
//                 query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
//                 variables: MATCH_ANY_PARAMETERS,
//             },
//             result: {
//                 data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED),
//             },
//             nMatches: Number.POSITIVE_INFINITY,
//         },
//         WORKSPACES_MOCK,
//         IMPORTING_CHANGESETS_MOCK,
//     ])

//     return (
//         <WebStory>
//             {props => (
//                 <MockedTestProvider link={mocks}>
//                     <WorkspacesPreview
//                         {...props}
//                         batchSpec={mockBatchSpec()}
//                         hasPreviewed={true}
//                         previewDisabled={false}
//                         preview={noop}
//                         batchSpecStale={true}
//                         excludeRepo={noop}
//                     />
//                 </MockedTestProvider>
//             )}
//         </WebStory>
//     )
// })

// TODO: Add these stories once the workspaces preview list is kept visible on subsequent updates
// add('subsequent preview, loading', () => {})
