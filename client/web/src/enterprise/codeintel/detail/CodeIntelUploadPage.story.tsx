import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CodeIntelUploadPage } from './CodeIntelUploadPage'

const { add } = storiesOf('web/codeintel/detail/CodeIntelUploadPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Uploading', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadPage
                {...props}
                fetchLsifUpload={fetch({
                    state: LSIFUploadState.UPLOADING,
                    uploadedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: null,
                    finishedAt: null,
                    failure: null,
                    placeInQueue: null,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Queued', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadPage
                {...props}
                fetchLsifUpload={fetch({
                    state: LSIFUploadState.QUEUED,
                    uploadedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: null,
                    finishedAt: null,
                    placeInQueue: 3,
                    failure: null,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Processing', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadPage
                {...props}
                fetchLsifUpload={fetch({
                    state: LSIFUploadState.PROCESSING,
                    uploadedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: null,
                    failure: null,
                    placeInQueue: null,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Completed', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadPage
                {...props}
                fetchLsifUpload={fetch({
                    state: LSIFUploadState.COMPLETED,
                    uploadedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: '2020-06-15T12:30:30+00:00',
                    failure: null,
                    placeInQueue: null,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Errored', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelUploadPage
                {...props}
                fetchLsifUpload={fetch({
                    state: LSIFUploadState.ERRORED,
                    uploadedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: '2020-06-15T12:30:30+00:00',
                    failure:
                        'Upload failed to complete: dial tcp: lookup gitserver-8.gitserver on 10.165.0.10:53: no such host',
                    placeInQueue: null,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))
const fetch = (
    upload: Pick<LsifUploadFields, 'state' | 'uploadedAt' | 'startedAt' | 'finishedAt' | 'failure' | 'placeInQueue'>
): (() => Observable<LsifUploadFields>) => () =>
    of({
        __typename: 'LSIFUpload',
        id: '1234',
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
    })

const now = () => new Date('2020-06-15T15:25:00+00:00')
