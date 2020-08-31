import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { VisibleChangesetSpecNode } from './VisibleChangesetSpecNode'
import { addDays } from 'date-fns'
import { VisibleChangesetSpecFields, ChangesetSpecType } from '../../../graphql-operations'
import { of } from 'rxjs'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/apply/VisibleChangesetSpecNode', module).addDecorator(story => (
    <div className="p-3 container web-content changeset-spec-list__grid">{story()}</div>
))

export const visibleChangesetSpecStories: Record<string, VisibleChangesetSpecFields> = {
    'Import changeset': {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv1',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
        description: {
            __typename: 'ExistingChangesetReference',
            baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
            externalID: '123',
        },
    },
    'Create changeset published': {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv2',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
        description: {
            __typename: 'GitBranchChangesetDescription',
            baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
            baseRef: 'master',
            headRef: 'cool-branch',
            body: 'Body text',
            commits: [{ message: 'Commit message' }],
            diffStat: {
                added: 10,
                changed: 8,
                deleted: 2,
            },
            published: true,
            title: 'Add prettier to repository',
        },
    },
    'Create changeset not published': {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv3',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
        description: {
            __typename: 'GitBranchChangesetDescription',
            baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
            baseRef: 'master',
            headRef: 'cool-branch',
            body: 'Body text',
            commits: [{ message: 'Commit message' }],
            diffStat: {
                added: 10,
                changed: 8,
                deleted: 2,
            },
            published: false,
            title: 'Add prettier to repository',
        },
    },
}

const queryEmptyFileDiffs = () =>
    of({ fileDiffs: { totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] } })

for (const storyName of Object.keys(visibleChangesetSpecStories)) {
    add(storyName, () => (
        <WebStory webStyles={webStyles}>
            {props => (
                <VisibleChangesetSpecNode
                    {...props}
                    node={visibleChangesetSpecStories[storyName]}
                    queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                />
            )}
        </WebStory>
    ))
}
