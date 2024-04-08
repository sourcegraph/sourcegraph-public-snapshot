import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { CardBody, Card } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import type { BatchSpecWorkspaceByIDResult } from '../../../../../graphql-operations'
import type { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from '../../../preview/list/backend'
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
    WORKSPACE_STEP_OUTPUT_LINES_PAGE_ONE,
    WORKSPACE_STEP_OUTPUT_LINES_PAGE_TWO,
} from '../../batch-spec.mock'
import {
    BATCH_SPEC_WORKSPACE_BY_ID,
    type queryBatchSpecWorkspaceStepFileDiffs as _queryBatchSpecWorkspaceStepFileDiffs,
    BATCH_SPEC_WORKSPACE_STEP,
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

const decorator: Decorator = story => (
    <div className="d-flex w-100" style={{ height: '95vh' }}>
        <Card className="w-100 overflow-auto flex-grow-1" style={{ backgroundColor: 'var(--color-bg-1)' }}>
            <div className="w-100">
                <CardBody>{story()}</CardBody>
            </div>
        </Card>
    </div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/execute/WorkspaceDetails',
    decorators: [decorator],
}

export default config

interface BaseStoryProps {
    node?: BatchSpecWorkspaceByIDResult['node']
    queries?: {
        queryBatchSpecWorkspaceStepFileDiffs?: typeof _queryBatchSpecWorkspaceStepFileDiffs
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    }
}

const BaseStory: React.FunctionComponent<BaseStoryProps> = ({ node, queries = {} }) => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BATCH_SPEC_WORKSPACE_BY_ID),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: { node } },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(BATCH_SPEC_WORKSPACE_STEP),
                variables: {
                    workspaceID: 'test-1234',
                    stepIndex: 1,
                    first: 500,
                    after: null,
                },
            },
            result: {
                data: WORKSPACE_STEP_OUTPUT_LINES_PAGE_ONE,
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(BATCH_SPEC_WORKSPACE_STEP),
                variables: {
                    workspaceID: 'test-1234',
                    stepIndex: 1,
                    first: 500,
                    after: '500',
                },
            },
            result: {
                data: WORKSPACE_STEP_OUTPUT_LINES_PAGE_TWO,
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <BrandedStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <WorkspaceDetails
                        {...props}
                        {...queries}
                        deselectWorkspace={noop}
                        id="random"
                        telemetryRecorder={noOpTelemetryRecorder}
                    />
                </MockedTestProvider>
            )}
        </BrandedStory>
    )
}

export const WorkspaceNotFound: StoryFn = () => <BaseStory />
WorkspaceNotFound.storyName = 'Workspace not found'

export const VisibleWorkspaceComplete: StoryFn = () => (
    <BaseStory node={mockWorkspace()} queries={MOCK_FILE_DIFF_QUERIES} />
)
VisibleWorkspaceComplete.storyName = 'Visible workspace: complete'

export const HiddenWorkspace: StoryFn = () => <BaseStory node={HIDDEN_WORKSPACE} />
HiddenWorkspace.storyName = 'Hidden workspace'

export const VisibleWorkspaceProcessing: StoryFn = () => <BaseStory node={PROCESSING_WORKSPACE} />
VisibleWorkspaceProcessing.storyName = 'Visible workspace: processing'

export const VisibleWorkspaceQueued: StoryFn = () => <BaseStory node={QUEUED_WORKSPACE} />
VisibleWorkspaceQueued.storyName = 'Visible workspace: queued'

export const VisibleWorkspaceSkipped: StoryFn = () => <BaseStory node={SKIPPED_WORKSPACE} />
VisibleWorkspaceSkipped.storyName = 'Visible workspace: skipped'

export const VisibleWorkspaceUnsupported: StoryFn = () => <BaseStory node={UNSUPPORTED_WORKSPACE} />
VisibleWorkspaceUnsupported.storyName = 'Visible workspace: unsupported'

export const VisibleWorkspaceCompleteWithLotsOfSteps: StoryFn = () => (
    <BaseStory node={LOTS_OF_STEPS_WORKSPACE} queries={MOCK_FILE_DIFF_QUERIES} />
)
VisibleWorkspaceCompleteWithLotsOfSteps.storyName = 'Visible workspace: complete with lots of steps'

export const VisibleWorkspaceFailed: StoryFn = () => (
    <BaseStory node={FAILED_WORKSPACE} queries={MOCK_FILE_DIFF_QUERIES} />
)
VisibleWorkspaceFailed.storyName = 'Visible workspace: failed'

export const VisibleWorkspaceCanceling: StoryFn = () => <BaseStory node={CANCELING_WORKSPACE} />
VisibleWorkspaceCanceling.storyName = 'Visible workspace: canceling'

export const VisibleWorkspaceCanceled: StoryFn = () => <BaseStory node={CANCELED_WORKSPACE} />
VisibleWorkspaceCanceled.storyName = 'Visible workspace: canceled'
