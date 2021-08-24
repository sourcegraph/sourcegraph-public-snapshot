import { storiesOf } from '@storybook/react'
import React, { useCallback } from 'react'
import { of } from 'rxjs'

import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelUploadsPage } from './CodeIntelUploadsPage'

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

const { add } = storiesOf('web/codeintel/list/CodeIntelUploadPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Empty', () => {
    const fetchLsifUploads = useCallback(
        () =>
            of({
                nodes: [],
                totalCount: 0,
                pageInfo: {
                    __typename: 'PageInfo',
                    endCursor: null,
                    hasNextPage: false,
                },
            }),
        []
    )

    return (
        <EnterpriseWebStory>
            {props => <CodeIntelUploadsPage {...props} now={now} fetchLsifUploads={fetchLsifUploads} />}
        </EnterpriseWebStory>
    )
})

add('SiteAdminPage', () => {
    const fetchLsifUploads = useCallback(
        () =>
            of({
                nodes: testUploads,
                totalCount: testUploads.length,
                pageInfo: {
                    __typename: 'PageInfo',
                    endCursor: null,
                    hasNextPage: false,
                },
            }),
        []
    )

    return (
        <EnterpriseWebStory>
            {props => <CodeIntelUploadsPage {...props} now={now} fetchLsifUploads={fetchLsifUploads} />}
        </EnterpriseWebStory>
    )
})

for (const { fresh, updated } of [
    { fresh: true, updated: true },
    { fresh: true, updated: false },
    { fresh: false, updated: true },
    { fresh: false, updated: false },
]) {
    add(`${fresh ? 'Fresh' : 'Stale'}${updated ? '' : 'Unupdated'}RepositoryPage`, () => {
        const fetchLsifUploads = useCallback(
            () =>
                of({
                    nodes: testUploads,
                    totalCount: testUploads.length,
                    pageInfo: {
                        __typename: 'PageInfo',
                        endCursor: null,
                        hasNextPage: false,
                    },
                }),
            []
        )

        return (
            <EnterpriseWebStory>
                {props => (
                    <CodeIntelUploadsPage
                        {...props}
                        repo={{ id: 'sourcegraph' }}
                        now={now}
                        fetchLsifUploads={fetchLsifUploads}
                        fetchCommitGraphMetadata={() => of({ stale: !fresh, updatedAt: updated ? null : now() })}
                    />
                )}
            </EnterpriseWebStory>
        )
    })
}
