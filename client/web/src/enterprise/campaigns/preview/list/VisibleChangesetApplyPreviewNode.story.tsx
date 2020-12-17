import { storiesOf } from '@storybook/react'
import React from 'react'
import { VisibleChangesetApplyPreviewNode } from './VisibleChangesetApplyPreviewNode'
import {
    VisibleChangesetSpecFields,
    ChangesetSpecType,
    Scalars,
    VisibleChangesetApplyPreviewFields,
} from '../../../../graphql-operations'
import { of } from 'rxjs'
import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'
import { ChangesetSpecOperation } from '../../../../../../shared/src/graphql-operations'

const { add } = storiesOf('web/campaigns/preview/VisibleChangesetApplyPreviewNode', module).addDecorator(story => (
    <div className="p-3 container web-content preview-list__grid">{story()}</div>
))

const testRepo = { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' }

function baseChangesetSpec(
    published: Scalars['PublishedValue'],
    overrides: Partial<VisibleChangesetSpecFields> = {}
): VisibleChangesetSpecFields {
    return {
        __typename: 'VisibleChangesetSpec',
        id: 'someidv2' + String(published) + JSON.stringify(overrides),
        type: ChangesetSpecType.EXISTING,
        description: {
            __typename: 'GitBranchChangesetDescription',
            baseRepository: testRepo,
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

export const visibleChangesetApplyPreviewNodeStories: Record<string, VisibleChangesetApplyPreviewFields> = {
    'Import changeset': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.IMPORT],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsAttach',
            changesetSpec: {
                __typename: 'VisibleChangesetSpec',
                id: 'someidv1',
                type: ChangesetSpecType.EXISTING,
                description: {
                    __typename: 'ExistingChangesetReference',
                    baseRepository: { name: 'github.com/sourcegraph/testrepo', url: 'https://test.test/repo' },
                    externalID: '123',
                },
            },
        },
    },
    'Create changeset published': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.PUSH, ChangesetSpecOperation.PUBLISH],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsAttach',
            changesetSpec: baseChangesetSpec(true),
        },
    },
    'Create changeset draft': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.PUSH, ChangesetSpecOperation.PUBLISH_DRAFT],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsAttach',
            changesetSpec: baseChangesetSpec('draft'),
        },
    },
    'Create changeset not published': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsAttach',
            changesetSpec: baseChangesetSpec(false),
        },
    },
    'Update changeset title': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.UPDATE],
        delta: {
            titleChanged: true,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsUpdate',
            changesetSpec: baseChangesetSpec(true),
            changeset: {
                id: '123123',
                title: 'the old title',
            },
        },
    },
    'Undraft changeset': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.UNDRAFT],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsUpdate',
            changesetSpec: baseChangesetSpec(true),
            changeset: {
                id: '123123',
                title: 'Le draft changeset',
            },
        },
    },
    'Reopen changeset': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.REOPEN],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsUpdate',
            changesetSpec: baseChangesetSpec(true),
            changeset: {
                id: '123123',
                title: 'Le closed changeset',
            },
        },
    },
    'Close changeset': {
        __typename: 'VisibleChangesetApplyPreview',
        operations: [ChangesetSpecOperation.CLOSE],
        delta: {
            titleChanged: false,
        },
        targets: {
            __typename: 'VisibleApplyPreviewTargetsDetach',
            changeset: {
                id: '123123',
                title: 'Le open changeset',
                repository: testRepo,
                diffStat: {
                    added: 2,
                    changed: 8,
                    deleted: 10,
                },
            },
        },
    },
}

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

for (const storyName of Object.keys(visibleChangesetApplyPreviewNodeStories)) {
    add(storyName, () => (
        <EnterpriseWebStory>
            {props => (
                <VisibleChangesetApplyPreviewNode
                    {...props}
                    node={visibleChangesetApplyPreviewNodeStories[storyName]}
                    queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                />
            )}
        </EnterpriseWebStory>
    ))
}
