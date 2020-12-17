import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { UploadConnection } from './backend'
import { CodeIntelUploadsPage } from './CodeIntelUploadsPage'

const { add } = storiesOf('web/codeintel/list/CodeIntelUploadPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Page', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadsPage
                {...props}
                now={now}
                fetchLsifUploads={fetch(
                    {
                        id: '1',
                        state: LSIFUploadState.UPLOADING,
                        uploadedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: null,
                        finishedAt: null,
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        id: '2',
                        state: LSIFUploadState.QUEUED,
                        uploadedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: null,
                        finishedAt: null,
                        placeInQueue: 3,
                        failure: null,
                    },
                    {
                        id: '3',
                        state: LSIFUploadState.PROCESSING,
                        uploadedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: null,
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        id: '4',
                        state: LSIFUploadState.COMPLETED,
                        uploadedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        id: '5',
                        state: LSIFUploadState.ERRORED,
                        uploadedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: 'Whoops! The server encountered a boo-boo handling this input.',
                        placeInQueue: null,
                    }
                )}
            />
        )}
    </EnterpriseWebStory>
))

const fetch = (
    ...uploads: Omit<
        LsifUploadFields,
        '__typename' | 'projectRoot' | 'inputCommit' | 'inputRoot' | 'inputIndexer' | 'isLatestForRepo'
    >[]
): (() => Observable<UploadConnection>) => () =>
    of({
        nodes: uploads.map(upload => ({
            __typename: 'LSIFUpload',
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
            inputCommit: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
            inputRoot: 'web/',
            inputIndexer: 'lsif-tsc',
            isLatestForRepo: false,
            ...upload,
        })),
        totalCount: 10,
        pageInfo: {
            __typename: 'PageInfo',
            endCursor: 'fakenextpage',
            hasNextPage: true,
        },
    })

const now = () => new Date('2020-06-15T15:25:00+00:00')
