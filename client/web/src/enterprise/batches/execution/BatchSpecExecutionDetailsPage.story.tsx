import { storiesOf } from '@storybook/react'
import { subDays, addMinutes, addHours } from 'date-fns'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecExecutionFields, BatchSpecState } from '../../../graphql-operations'

import { BatchSpecExecutionDetailsPage } from './BatchSpecExecutionDetailsPage'

const { add } = storiesOf('web/batches/execution/BatchSpecExecutionDetailsPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const originalInput =
    "name: hello-world\ndescription: Add Hello World to READMEs\n\n# Find all repositories that contain a README.md file.\non:\n  - repositoriesMatchingQuery: file:README.md\n\n# In each repository, run this command. Each repository's resulting diff is captured.\nsteps:\n  - run: echo Hello World | tee -a $(find -name README.md)\n    container: ubuntu:18.04\n\n# Describe the changeset (e.g., GitHub pull request) you want for each repository.\nchangesetTemplate:\n  title: Hello World\n  body: My first batch change!\n  branch: hello-world # Push the commit to this branch.\n  commit:\n    message: Append Hello World to all README.md files\n  published: false\n"

const now = new Date()
const createdAt = addHours(subDays(now, 2), 12)
const startedAt = addMinutes(createdAt, 1).toISOString()
const finishedAt = addMinutes(createdAt, 5).toISOString()

const batchSpecExecutionCompleted = (): BatchSpecExecutionFields => ({
    id: 'QmF0Y2hTcGVjRXhlY3V0aW9uOiI5THpOWlBmUHhKSSI=',
    originalInput,
    state: BatchSpecState.COMPLETED,
    createdAt: createdAt.toISOString(),
    startedAt,
    finishedAt,
    failureMessage: null,
    applyURL: '/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiI5NzZxcXAydzBwYyI=',
    creator: {
        id: 'VXNlcjox',
        url: '/users/mrnugget',
        displayName: null,
    },
    namespace: {
        id: 'VXNlcjox',
        url: '/users/mrnugget',
        namespaceName: 'mrnugget',
    },
})

const batchSpecExecutionFailed = (): BatchSpecExecutionFields => ({
    id: '1234',
    originalInput,
    createdAt: createdAt.toISOString(),
    startedAt: addMinutes(createdAt, 1).toISOString(),
    finishedAt: addMinutes(createdAt, 2).toISOString(),
    failureMessage: 'failed to perform src-cli step: command failed',
    state: BatchSpecState.FAILED,
    creator: {
        id: 'VXNlcjox',
        url: '/users/mrnugget',
        displayName: null,
    },
    applyURL: null,
    namespace: {
        id: 'VXNlcjox',
        url: '/users/mrnugget',
        namespaceName: 'mrnugget',
    },
})

add('Completed', () => (
    <WebStory>
        {props => (
            <BatchSpecExecutionDetailsPage
                {...props}
                executionID="123123"
                fetchBatchSpecExecution={() => of(batchSpecExecutionCompleted())}
                expandStage="srcPreview"
            />
        )}
    </WebStory>
))

add('Failed', () => (
    <WebStory>
        {props => (
            <BatchSpecExecutionDetailsPage
                {...props}
                executionID="123123"
                fetchBatchSpecExecution={() => of(batchSpecExecutionFailed())}
                expandStage="srcPreview"
            />
        )}
    </WebStory>
))
