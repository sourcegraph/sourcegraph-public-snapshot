import React from 'react'

import { storiesOf } from '@storybook/react'
import { subMinutes } from 'date-fns/esm'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    BatchSpecWorkspaceByIDResult,
    BatchSpecWorkspaceState,
    HiddenBatchSpecWorkspaceFields,
    VisibleBatchSpecWorkspaceFields,
} from '../../../graphql-operations'

import { BATCH_SPEC_WORKSPACE_BY_ID } from './backend'
import { WorkspaceDetails } from './WorkspaceDetails'

const { add } = storiesOf('web/batches/execution/WorkspaceDetails', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const HIDDEN_WORKSPACE: HiddenBatchSpecWorkspaceFields = {
    __typename: 'HiddenBatchSpecWorkspace',
    id: 'id123',
    queuedAt: subMinutes(new Date(), 10).toISOString(),
    startedAt: subMinutes(new Date(), 8).toISOString(),
    finishedAt: subMinutes(new Date(), 2).toISOString(),
    state: BatchSpecWorkspaceState.COMPLETED,
    diffStat: {
        __typename: 'DiffStat',
        added: 10,
        changed: 2,
        deleted: 5,
    },
    placeInQueue: null,
    onlyFetchWorkspace: false,
    ignored: false,
    unsupported: false,
    cachedResultFound: false,
}

const VISIBLE_WORKSPACE: VisibleBatchSpecWorkspaceFields = {
    ...HIDDEN_WORKSPACE,
    __typename: 'VisibleBatchSpecWorkspace',
    steps: [],
    path: '',
    branch: {
        displayName: 'asdf',
    },
    executor: null,
    stages: null,
    searchResultPaths: ['asdf'],
    failureMessage: null,
    changesetSpecs: [],
    repository: {
        name: 'github.com/sourcegraph/automation-testing',
        url: '/github.com/sourcegraph/automation-testing',
    },
}

function addStory(name: string, node: BatchSpecWorkspaceByIDResult['node']) {
    add(name, () => {
        const mocks = new WildcardMockLink([
            {
                request: {
                    query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
                    variables: MATCH_ANY_PARAMETERS,
                },
                result: {
                    data: {
                        node,
                    },
                },
                nMatches: Number.POSITIVE_INFINITY,
            },
        ])

        return (
            <WebStory>
                {props => (
                    <MockedTestProvider link={mocks}>
                        <WorkspaceDetails {...props} deselectWorkspace={noop} id="random" />
                    </MockedTestProvider>
                )}
            </WebStory>
        )
    })
}

addStory('Hidden workspace', HIDDEN_WORKSPACE)
addStory('Visible workspace', VISIBLE_WORKSPACE)
addStory('Not found', null)
