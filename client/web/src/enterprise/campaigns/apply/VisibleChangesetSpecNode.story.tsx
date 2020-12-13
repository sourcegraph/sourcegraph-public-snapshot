import { storiesOf } from '@storybook/react'
import React from 'react'
import { VisibleChangesetSpecNode } from './VisibleChangesetSpecNode'
import { addDays } from 'date-fns'
import { VisibleChangesetSpecFields, ChangesetSpecType, Scalars } from '../../../graphql-operations'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { ChangesetSpecOperation } from '../../../../../shared/src/graphql-operations'

const { add } = storiesOf('web/campaigns/apply/VisibleChangesetSpecNode', module).addDecorator(story => (
    <div className="p-3 container web-content changeset-spec-list__grid">{story()}</div>
))

function baseChangeset(
    published: Scalars['PublishedValue'],
    overrides: Partial<VisibleChangesetSpecFields> = {}
): VisibleChangesetSpecFields {
    return {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv2' + String(published) + JSON.stringify(overrides),
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
        operations: [],
        delta: {
            titleChanged: false,
        },
        changeset: null,
        description: {
            __typename: 'GitBranchChangesetDescription',
            baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
            baseRef: 'master',
            headRef: 'cool-branch',
            body: 'Body text',
            commits: [
                {
                    subject: 'This is the first line of the commit message',
                    body: `And the more explanatory body. And the more explanatory body.
And the more explanatory body. And the more explanatory body.
And the more explanatory body. And the more explanatory body.
And the more explanatory body. And the more explanatory body. And the more explanatory body.
And the more explanatory body. And the more explanatory body. And the more explanatory body.`,
                    author: {
                        avatarURL: null,
                        displayName: 'john',
                        email: 'john@test.not',
                        user: { displayName: 'lejohn', url: '/users/lejohn', username: 'john' },
                    },
                },
            ],
            diffStat: {
                added: 10,
                changed: 8,
                deleted: 2,
            },
            title: 'Add prettier to repository',
            published,
        },
        ...overrides,
    }
}

export const visibleChangesetSpecStories: Record<string, VisibleChangesetSpecFields> = {
    'Import changeset': {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv1',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
        operations: [ChangesetSpecOperation.IMPORT],
        delta: {
            titleChanged: false,
        },
        changeset: null,
        description: {
            __typename: 'ExistingChangesetReference',
            baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
            externalID: '123',
        },
    },
    'Create changeset published': baseChangeset(true, {
        operations: [ChangesetSpecOperation.PUSH, ChangesetSpecOperation.PUBLISH],
        delta: {
            titleChanged: false,
        },
        changeset: null,
    }),
    'Create changeset draft': baseChangeset('draft', {
        operations: [ChangesetSpecOperation.PUSH, ChangesetSpecOperation.PUBLISH_DRAFT],
        delta: {
            titleChanged: false,
        },
        changeset: null,
    }),
    'Create changeset not published': baseChangeset(false),
    'Update changeset title': baseChangeset(true, {
        operations: [ChangesetSpecOperation.UPDATE],
        delta: {
            titleChanged: true,
        },
        changeset: {
            __typename: 'ExternalChangeset',
            id: '123123',
            title: 'the old title',
        },
    }),
    'Undraft changeset': baseChangeset(true, {
        operations: [ChangesetSpecOperation.UNDRAFT],
        delta: {
            titleChanged: false,
        },
        changeset: {
            __typename: 'ExternalChangeset',
            id: '123123',
            title: 'Le draft changeset',
        },
    }),
    'Reopen changeset': baseChangeset(true, {
        operations: [ChangesetSpecOperation.REOPEN],
        delta: {
            titleChanged: false,
        },
        changeset: {
            __typename: 'ExternalChangeset',
            id: '123123',
            title: 'Le closed changeset',
        },
    }),
    'Close changeset': baseChangeset(true, {
        operations: [ChangesetSpecOperation.CLOSE],
        delta: {
            titleChanged: false,
        },
        changeset: {
            __typename: 'ExternalChangeset',
            id: '123123',
            title: 'Le open changeset',
        },
    }),
}

const queryEmptyFileDiffs = () =>
    of({ fileDiffs: { totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] } })

for (const storyName of Object.keys(visibleChangesetSpecStories)) {
    add(storyName, () => (
        <EnterpriseWebStory>
            {props => (
                <VisibleChangesetSpecNode
                    {...props}
                    node={visibleChangesetSpecStories[storyName]}
                    queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                />
            )}
        </EnterpriseWebStory>
    ))
}
