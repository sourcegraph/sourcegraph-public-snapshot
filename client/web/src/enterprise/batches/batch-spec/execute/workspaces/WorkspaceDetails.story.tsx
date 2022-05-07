import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { CardBody, Card } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceByIDResult } from '../../../../../graphql-operations'
import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from '../../../preview/list/backend'
import {
    HIDDEN_WORKSPACE,
    QUEUED_WORKSPACE,
    mockWorkspace,
    PROCESSING_WORKSPACE,
    SKIPPED_WORKSPACE,
    UNSUPPORTED_WORKSPACE,
    LOTS_OF_STEPS_WORKSPACE,
    FAILED_WORKSPACE,
    CANCELING_WORKSPACE,
    CANCELED_WORKSPACE,
} from '../../batch-spec.mock'
import {
    BATCH_SPEC_WORKSPACE_BY_ID,
    queryBatchSpecWorkspaceStepFileDiffs as _queryBatchSpecWorkspaceStepFileDiffs,
} from '../backend'

import { WorkspaceDetails } from './WorkspaceDetails'

const queryChangesetSpecFileDiffs = () =>
    of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

const queryBatchSpecWorkspaceStepFileDiffs = () =>
    of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

const MOCK_FILE_DIFF_QUERIES = {
    queryBatchSpecWorkspaceStepFileDiffs,
    queryChangesetSpecFileDiffs,
}

const { add } = storiesOf('web/batches/batch-spec/execute/WorkspaceDetails', module).addDecorator(story => (
    <div className="d-flex w-100" style={{ height: '95vh' }}>
        <Card className="w-100 overflow-auto flex-grow-1" style={{ backgroundColor: 'var(--color-bg-1)' }}>
            <div className="w-100">
                <CardBody>{story()}</CardBody>
            </div>
        </Card>
    </div>
))

function addStory(
    name: string,
    node: BatchSpecWorkspaceByIDResult['node'],
    queries: {
        queryBatchSpecWorkspaceStepFileDiffs?: typeof _queryBatchSpecWorkspaceStepFileDiffs
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    } = {}
) {
    add(name, () => {
        const mocks = new WildcardMockLink([
            {
                request: {
                    query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
                    variables: MATCH_ANY_PARAMETERS,
                },
                result: { data: { node } },
                nMatches: Number.POSITIVE_INFINITY,
            },
        ])

        return (
            <BrandedStory>
                {props => (
                    <MockedTestProvider link={mocks}>
                        <WorkspaceDetails {...props} {...queries} deselectWorkspace={noop} id="random" />
                    </MockedTestProvider>
                )}
            </BrandedStory>
        )
    })
}

addStory('Hidden workspace', HIDDEN_WORKSPACE)
addStory('Workspace not found', null)
addStory('Visible workspace: complete', mockWorkspace(), MOCK_FILE_DIFF_QUERIES)
addStory('Visible workspace: complete with lots of steps', LOTS_OF_STEPS_WORKSPACE, MOCK_FILE_DIFF_QUERIES)
addStory('Visible workspace: queued', QUEUED_WORKSPACE)
addStory('Visible workspace: processing', PROCESSING_WORKSPACE)
addStory('Visible workspace: skipped', SKIPPED_WORKSPACE)
addStory('Visible workspace: unsupported', UNSUPPORTED_WORKSPACE)
addStory('Visible workspace: failed', FAILED_WORKSPACE, MOCK_FILE_DIFF_QUERIES)
addStory('Visible workspace: canceling', CANCELING_WORKSPACE)
addStory('Visible workspace: canceled', CANCELED_WORKSPACE)
