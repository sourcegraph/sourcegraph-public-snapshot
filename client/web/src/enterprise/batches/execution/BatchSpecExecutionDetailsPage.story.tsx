import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql/schema'

import { BatchSpecExecutionFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { fetchBatchSpecExecution } from './backend'
import { BatchSpecExecutionDetailsPage } from './BatchSpecExecutionDetailsPage'

const { add } = storiesOf('web/batches/execution/BatchSpecExecutionDetailsPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const batchSpecExecution = (): BatchSpecExecutionFields => ({
    id: '1234',
    inputSpec: 'speccccc.yml',
    createdAt: '2021-08-02T14:50:15Z',
    startedAt: '2021-08-02T14:50:16Z',
    finishedAt: '2021-08-02T14:50:22Z',
    failure: 'failed to perform src-cli step: command failed',
    state: BatchSpecExecutionState.ERRORED,
    steps: {
        setup: [
            {
                key: 'setup.firecracker.start',
                command: [
                    'ignite',
                    'run',
                    '--runtime',
                    'docker',
                    '--network-plugin',
                    'docker-bridge',
                    '--cpus',
                    '4',
                    '--memory',
                    '12G',
                    '--size',
                    '20G',
                    '--copy-files',
                    '/tmp/781490878:/work',
                    '--ssh',
                    '--name',
                    'USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3',
                    'sourcegraph/ignite-ubuntu:insiders',
                ],
                startTime: '2021-08-02T16:50:16+02:00',
                exitCode: 0,
                durationMilliseconds: 4889,
                out:
                    'stdout: time="2021-08-02T16:50:17+02:00" level=info msg="Created VM with ID \\"32cc4f205cd58550\\" and name \\"USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3\\""\nstdout: time="2021-08-02T16:50:18+02:00" level=info msg="Networking is handled by \\"docker-bridge\\""\nstdout: time="2021-08-02T16:50:18+02:00" level=info msg="Started Firecracker VM \\"32cc4f205cd58550\\" in a container with ID \\"e142627b331e3f0ff4fb2e0092c356fd8febd3cea9eb547b699e09c451d77ead\\""\nstdout: time="2021-08-02T16:50:18+02:00" level=info msg="Waiting for the ssh daemon within the VM to start..."\n',
            },
        ],
        srcPreview: {
            key: 'step.src.0',
            command: [
                'ignite',
                'exec',
                'USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3',
                '--',
                'cd /work && SRC_ENDPOINT=https://USERNAME_REMOVED:PASSWORD_REMOVED@sourcegraph.test:3443 SRC_ACCESS_TOKEN=SRC_ACCESS_TOKEN_REMOVED HOME=/home/mrnugget PATH=/home/mrnugget/google-cloud-sdk/bin:/home/mrnugget/bin:/home/mrnugget/.yarn/bin:/home/mrnugget/.config/yarn/global/node_modules/.bin:/usr/local/heroku/bin:/home/mrnugget/code/go/bin:/home/mrnugget/.asdf/shims:/usr/local/opt/asdf/bin:/usr/local/bin:/home/mrnugget/.cargo/bin:/home/mrnugget/.local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin:/home/mrnugget/.fzf/bin src batch preview -f spec.yml -text-only -skip-errors -n mrnugget',
            ],
            startTime: '2021-08-02T16:50:21+02:00',
            exitCode: 1,
            durationMilliseconds: 157,
            out:
                'stdout: {"operation":"BATCH_SPEC_EXECUTION","timestamp":"2021-08-02T14:50:21.68Z","status":"FAILURE","message":"failed to query Sourcegraph version to check for available features: Post \\"https://USERNAME_REMOVED:***@sourcegraph.test:3443/.api/graphql\\": dial tcp: lookup sourcegraph.test on 192.168.1.1:53: no such host"}\nstdout: time="2021-08-02T16:50:21+02:00" level=error msg="Process exited with status 1\\n"\n',
        },
        teardown: [
            {
                key: 'teardown.firecracker.stop',
                command: ['ignite', 'stop', 'USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3'],
                startTime: '2021-08-02T16:50:21+02:00',
                exitCode: 0,
                durationMilliseconds: 701,
                out:
                    'stdout: time="2021-08-02T16:50:21+02:00" level=info msg="Removing the container with ID \\"e142627b331e3f0ff4fb2e0092c356fd8febd3cea9eb547b699e09c451d77ead\\" from the \\"docker-bridge\\" network"\nstdout: time="2021-08-02T16:50:22+02:00" level=info msg="Stopped VM with name \\"USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3\\" and ID \\"32cc4f205cd58550\\""\n',
            },
            {
                key: 'teardown.firecracker.remove',
                command: ['ignite', 'rm', '-f', 'USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3'],
                startTime: '2021-08-02T16:50:22+02:00',
                exitCode: 0,
                durationMilliseconds: 17,
                out:
                    'stdout: time="2021-08-02T16:50:22+02:00" level=info msg="Removed VM with name \\"USERNAME_REMOVED-1c25d857-837e-4d16-bfd6-12b9034fcad3\\" and ID \\"32cc4f205cd58550\\""\n',
            },
        ],
    },
    placeInQueue: null,
    batchSpec: null,
    initiator: {
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

const fetchBatchExecution: typeof fetchBatchSpecExecution = () => of(batchSpecExecution())

add('Show', () => (
    <EnterpriseWebStory>
        {props => (
            <BatchSpecExecutionDetailsPage
                {...props}
                executionID="123123"
                fetchBatchSpecExecution={fetchBatchExecution}
            />
        )}
    </EnterpriseWebStory>
))
