import { subSeconds } from 'date-fns'

import {
    type BatchSpecListFields,
    BatchSpecSource,
    BatchSpecState,
    type BatchSpecWorkspaceFileResult,
} from '../../graphql-operations'

const COMMON_NODE_FIELDS = {
    __typename: 'BatchSpec',
    createdAt: subSeconds(new Date(), 30).toISOString(),
    startedAt: subSeconds(new Date(), 25).toISOString(),
    finishedAt: new Date().toISOString(),
    originalInput: 'name: super-cool-spec',
    description: {
        __typename: 'BatchChangeDescription',
        name: 'super-cool-spec',
    },
    source: BatchSpecSource.LOCAL,
    namespace: {
        url: '/users/courier-new',
        namespaceName: 'courier-new',
    },
    creator: {
        username: 'courier-new',
    },
    files: null,
} as const

export const successNode = (id: string): BatchSpecListFields => ({
    ...COMMON_NODE_FIELDS,
    id,
    state: BatchSpecState.COMPLETED,
})

export const NODES: BatchSpecListFields[] = [
    { ...COMMON_NODE_FIELDS, id: 'id1', state: BatchSpecState.QUEUED },
    { ...COMMON_NODE_FIELDS, id: 'id2', state: BatchSpecState.PROCESSING },
    successNode('id3'),
    { ...COMMON_NODE_FIELDS, id: 'id4', state: BatchSpecState.FAILED },
    { ...COMMON_NODE_FIELDS, id: 'id5', state: BatchSpecState.CANCELING },
    { ...COMMON_NODE_FIELDS, id: 'id6', state: BatchSpecState.CANCELED },
    {
        ...COMMON_NODE_FIELDS,
        state: BatchSpecState.COMPLETED,
        source: BatchSpecSource.REMOTE,
        id: 'id7',
        originalInput: `name: super-cool-spec
description: doing something super interesting

on:
    - repository: github.com/foo/bar
`,
        description: {
            __typename: 'BatchChangeDescription',
            name: 'remote-super-cool-spec',
        },
        files: {
            totalCount: 2,
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            nodes: [
                {
                    id: 'fileId1',
                    name: 'test.sh',
                    binary: false,
                    byteSize: 12,
                    url: 'test/url',
                },
                {
                    id: 'fileId2',
                    name: 'src-cli',
                    binary: true,
                    byteSize: 19000,
                    url: 'test/url',
                },
            ],
        },
    },
    {
        ...COMMON_NODE_FIELDS,
        state: BatchSpecState.COMPLETED,
        source: BatchSpecSource.LOCAL,
        id: 'id8',
        originalInput: `name: super-cool-spec
description: doing something super interesting

on:
    - repository: github.com/foo/bar
`,
        description: {
            __typename: 'BatchChangeDescription',
            name: 'local-super-cool-spec',
        },
        files: {
            totalCount: 1,
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            nodes: [
                {
                    id: 'fileId3',
                    name: 'test.sh',
                    binary: false,
                    byteSize: 12,
                    url: 'test/url',
                },
            ],
        },
    },
]

export const MOCK_HIGHLIGHTED_FILES: BatchSpecWorkspaceFileResult = {
    __typename: 'Query',
    node: {
        __typename: 'BatchSpecWorkspaceFile',
        id: 'fileId1',
        name: 'test.sh',
        binary: false,
        byteSize: 12,
        url: 'test/url',
        highlight: {
            aborted: false,
            __typename: 'HighlightedFile',
            html: `import { React } from 'react';

const MyComponent = () => <div>My Component</div>`,
        },
    },
}
