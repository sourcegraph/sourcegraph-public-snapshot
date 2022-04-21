import React from 'react'

import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { CardBody, Card } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceByIDResult } from '../../../graphql-operations'

import { BATCH_SPEC_WORKSPACE_BY_ID } from './backend'
import { WorkspaceDetails } from './WorkspaceDetails'
import {
    HIDDEN_WORKSPACE,
    QUEUED_WORKSPACE,
    mockWorkspace,
    PROCESSING_WORKSPACE,
    SKIPPED_WORKSPACE,
    UNSUPPORTED_WORKSPACE,
    LOTS_OF_STEPS_WORKSPACE,
} from './WorkspaceDetails.mock'

const { add } = storiesOf('web/batches/execution/WorkspaceDetails', module).addDecorator(story => (
    <div className="d-flex w-100" style={{ height: '95vh' }}>
        <Card className="w-100 overflow-auto flex-grow-1" style={{ backgroundColor: 'var(--color-bg-1)' }}>
            <div className="w-100">
                <CardBody>{story()}</CardBody>
            </div>
        </Card>
    </div>
))

function addStory(name: string, node: BatchSpecWorkspaceByIDResult['node']) {
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
                        <WorkspaceDetails {...props} deselectWorkspace={noop} id="random" />
                    </MockedTestProvider>
                )}
            </BrandedStory>
        )
    })
}

addStory('Hidden workspace', HIDDEN_WORKSPACE)
addStory('Workspace not found', null)
addStory('Visible workspace: complete', mockWorkspace())
addStory('Visible workspace: complete with lots of steps', LOTS_OF_STEPS_WORKSPACE)
addStory('Visible workspace: queued', QUEUED_WORKSPACE)
addStory('Visible workspace: processing', PROCESSING_WORKSPACE)
addStory('Visible workspace: skipped', SKIPPED_WORKSPACE)
addStory('Visible workspace: unsupported', UNSUPPORTED_WORKSPACE)
