import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { LSIFIndexState } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../../../components/WebStory'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'

import { CodeIntelUploadPage, CodeIntelUploadPageProps } from './CodeIntelUploadPage'

const uploadPrototype: Omit<LsifUploadFields, 'id' | 'state' | 'uploadedAt'> = {
    __typename: 'LSIFUpload',
    inputCommit: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
    inputRoot: 'web/',
    inputIndexer: 'lsif-tsc',
    failure: null,
    isLatestForRepo: false,
    startedAt: null,
    finishedAt: null,
    placeInQueue: null,
    projectRoot: {
        url: '',
        path: 'web/',
        repository: {
            url: '',
            name: 'github.com/sourcegraph/sourcegraph',
        },
        commit: {
            url: '',
            oid: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
            abbreviatedOID: '9ea5e9f',
        },
    },
    associatedIndex: null,
}

const dependendentPrototype = {
    ...uploadPrototype,

    inputCommit: 'a24285d041f0c779b3c60f7937f6508d2474472b',
    inputRoot: 'lib/',
    state: LSIFUploadState.COMPLETED,
    uploadedAt: '2020-06-14T12:10:30+00:00',
    startedAt: '2020-06-14T12:10:30+00:00',
    finishedAt: '2020-06-14T12:10:30+00:00',
    projectRoot: {
        url: '',
        path: 'lib/',
        repository: {
            url: '',
            name: 'github.com/sourcegraph/dependent',
        },
        commit: {
            url: '',
            oid: 'a24285d041f0c779b3c60f7937f6508d2474472b',
            abbreviatedOID: 'a24285d',
        },
    },
}

const dependents = [
    { ...dependendentPrototype, id: 'd1' },
    { ...dependendentPrototype, id: 'd2' },
]

const dependencyPrototype = {
    ...uploadPrototype,

    inputCommit: 'f28cec95e8bc5d6b703aa1f953898a45c785db76',
    inputRoot: 'lib/',
    state: LSIFUploadState.COMPLETED,
    uploadedAt: '2020-06-14T12:10:30+00:00',
    startedAt: '2020-06-14T12:10:30+00:00',
    finishedAt: '2020-06-14T12:10:30+00:00',
    projectRoot: {
        url: '',
        path: 'lib/',
        repository: {
            url: '',
            name: 'github.com/sourcegraph/dependency',
        },
        commit: {
            url: '',
            oid: 'f28cec95e8bc5d6b703aa1f953898a45c785db76',
            abbreviatedOID: 'f28cec9',
        },
    },
}

const dependencies = [
    { ...dependencyPrototype, id: 'd3' },
    { ...dependencyPrototype, id: 'd4' },
    { ...dependencyPrototype, id: 'd5' },
]

const now = () => new Date('2020-06-15T15:25:00+00:00')

const story: Meta = {
    title: 'web/codeintel/detail/CodeIntelUploadPage',
    decorators: [story => <div className="p-3 container">{story()}</div>],
    parameters: {
        component: CodeIntelUploadPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelUploadPageProps> = args => (
    <WebStory>{props => <CodeIntelUploadPage {...props} {...args} />}</WebStory>
)

const defaults: Partial<CodeIntelUploadPageProps> = {
    now,
    queryLsifUploadsList: ({ dependencyOf }: { dependencyOf?: string | null }) =>
        dependencyOf === undefined
            ? of({
                  nodes: dependents,
                  totalCount: dependents.length,
                  pageInfo: {
                      __typename: 'PageInfo',
                      endCursor: null,
                      hasNextPage: false,
                  },
              })
            : of({
                  nodes: dependencies,
                  totalCount: dependencies.length,
                  pageInfo: {
                      __typename: 'PageInfo',
                      endCursor: null,
                      hasNextPage: false,
                  },
              }),
}

export const Uploading = Template.bind({})
Uploading.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.UPLOADING,
            uploadedAt: '2020-06-15T15:25:00+00:00',
        }),
}

export const Queued = Template.bind({})
Queued.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.QUEUED,
            uploadedAt: '2020-06-15T12:20:30+00:00',
            placeInQueue: 1,
        }),
}

export const Processing = Template.bind({})
Processing.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.PROCESSING,
            uploadedAt: '2020-06-15T12:20:30+00:00',
            startedAt: '2020-06-15T12:25:30+00:00',
        }),
}

export const Completed = Template.bind({})
Completed.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.COMPLETED,
            uploadedAt: '2020-06-14T12:20:30+00:00',
            startedAt: '2020-06-14T12:25:30+00:00',
            finishedAt: '2020-06-14T12:30:30+00:00',
        }),
}

export const Errored = Template.bind({})
Errored.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.ERRORED,
            uploadedAt: '2020-06-13T12:20:30+00:00',
            startedAt: '2020-06-13T12:25:30+00:00',
            finishedAt: '2020-06-13T12:30:30+00:00',
            failure:
                'Upload failed to complete: dial tcp: lookup gitserver-8.gitserver on 10.165.0.10:53: no such host',
        }),
}

export const Deleting = Template.bind({})
Deleting.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.DELETING,
            uploadedAt: '2020-06-14T12:20:30+00:00',
            startedAt: '2020-06-14T12:25:30+00:00',
            finishedAt: '2020-06-14T12:30:30+00:00',
        }),
}

export const FailedUpload = Template.bind({})
FailedUpload.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.ERRORED,
            uploadedAt: '2020-06-13T12:20:30+00:00',
            startedAt: null,
            finishedAt: '2020-06-13T12:20:31+00:00',
            failure: 'Upload failed to complete: object store error:\n * XMinioStorageFull etc etc',
        }),
}

export const AssociatedIndex = Template.bind({})
AssociatedIndex.args = {
    ...defaults,
    queryLisfUploadFields: () =>
        of({
            ...uploadPrototype,
            id: '1',
            state: LSIFUploadState.PROCESSING,
            uploadedAt: '2020-06-15T12:20:30+00:00',
            startedAt: '2020-06-15T12:25:30+00:00',
            associatedIndex: {
                id: '2',
                state: LSIFIndexState.COMPLETED,
                queuedAt: '2020-06-15T12:15:10+00:00',
                startedAt: '2020-06-15T12:20:20+00:00',
                finishedAt: '2020-06-15T12:25:30+00:00',
                placeInQueue: null,
            },
        }),
}
