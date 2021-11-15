import { boolean } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { LsifUploadConnectionFields, LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'

import { CodeIntelUploadsPage, CodeIntelUploadsPageProps } from './CodeIntelUploadsPage'

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

const testUploads: LsifUploadFields[] = [
    {
        ...uploadPrototype,
        id: '6',
        state: LSIFUploadState.UPLOADING,
        uploadedAt: '2020-06-15T15:25:00+00:00',
    },
    {
        ...uploadPrototype,
        id: '5',
        state: LSIFUploadState.QUEUED,
        uploadedAt: '2020-06-15T12:20:30+00:00',
        placeInQueue: 1,
    },
    {
        ...uploadPrototype,
        id: '4',
        state: LSIFUploadState.PROCESSING,
        uploadedAt: '2020-06-15T12:20:30+00:00',
        startedAt: '2020-06-15T12:25:30+00:00',
    },
    {
        ...uploadPrototype,
        id: '3',
        state: LSIFUploadState.COMPLETED,
        uploadedAt: '2020-06-14T12:20:30+00:00',
        startedAt: '2020-06-14T12:25:30+00:00',
        finishedAt: '2020-06-14T12:30:30+00:00',
    },
    {
        ...uploadPrototype,
        id: '2',
        state: LSIFUploadState.ERRORED,
        uploadedAt: '2020-06-13T12:20:30+00:00',
        startedAt: '2020-06-13T12:25:30+00:00',
        finishedAt: '2020-06-13T12:30:30+00:00',
        failure: 'Upload failed to complete: dial tcp: lookup gitserver-8.gitserver on 10.165.0.10:53: no such host',
    },
    {
        ...uploadPrototype,
        id: '1',
        state: LSIFUploadState.DELETING,
        uploadedAt: '2020-06-14T12:20:30+00:00',
        startedAt: '2020-06-14T12:25:30+00:00',
        finishedAt: '2020-06-14T12:30:30+00:00',
    },
]

const now = () => new Date('2020-06-15T15:25:00+00:00')

const makeResponse = (uploads: LsifUploadFields[]): LsifUploadConnectionFields => ({
    __typename: 'LSIFUploadConnection',
    nodes: uploads,
    totalCount: uploads.length,
    pageInfo: {
        __typename: 'PageInfo',
        endCursor: null,
        hasNextPage: false,
    },
})

const story: Meta = {
    title: 'web/codeintel/list/CodeIntelUploadPage',
    decorators: [story => <div className="p-3 container">{story()}</div>],
    parameters: {
        component: CodeIntelUploadsPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelUploadsPageProps> = args => {
    const queryCommitGraphMetadata = () =>
        of({
            stale: boolean('staleCommitGraph', false),
            updatedAt: boolean('previouslyUpdatedCommitGraph', true) ? now() : null,
        })

    return (
        <WebStory>
            {props => <CodeIntelUploadsPage {...props} queryCommitGraphMetadata={queryCommitGraphMetadata} {...args} />}
        </WebStory>
    )
}

const defaults: Partial<CodeIntelUploadsPageProps> = {
    now,
    queryLsifUploadsByRepository: () => of(makeResponse([])),
}

export const EmptyGlobalPage = Template.bind({})
EmptyGlobalPage.args = {
    ...defaults,
    queryLsifUploadsList: () => of(makeResponse([])),
}

export const GlobalPage = Template.bind({})
GlobalPage.args = {
    ...defaults,
    queryLsifUploadsList: () => of(makeResponse(testUploads)),
}

export const EmptyRepositoryPage = Template.bind({})
EmptyRepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
    queryLsifUploadsByRepository: () => of(makeResponse(testUploads)),
}
