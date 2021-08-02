import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql/schema'

import { BatchSpecExecutionFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchSpecExecutionDetailsPage } from './BatchSpecExecutionDetailsPage'

const { add } = storiesOf('web/batches/execution/BatchSpecExecutionDetailsPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const batchSpecExecutionCompleted = (): BatchSpecExecutionFields => ({
    id: 'QmF0Y2hTcGVjRXhlY3V0aW9uOiI3QUJtNnR5NVYzWSI=',
    inputSpec:
        "name: hello-world\ndescription: Add Hello World to READMEs\n\n# Find all repositories that contain a README.md file.\non:\n  - repositoriesMatchingQuery: file:README.md\n\n# In each repository, run this command. Each repository's resulting diff is captured.\nsteps:\n  - run: echo Hello World | tee -a $(find -name README.md)\n    container: ubuntu:18.04\n\n# Describe the changeset (e.g., GitHub pull request) you want for each repository.\nchangesetTemplate:\n  title: Hello World\n  body: My first batch change!\n  branch: hello-world # Push the commit to this branch.\n  commit:\n    message: Append Hello World to all README.md files\n  published: false\n",
    state: BatchSpecExecutionState.COMPLETED,
    createdAt: '2021-08-02T15:10:13Z',
    startedAt: '2021-08-02T15:10:13Z',
    finishedAt: '2021-08-02T15:10:37Z',
    failure: null,
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
                    '/tmp/2812429741:/work',
                    '--ssh',
                    '--name',
                    'USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b',
                    'sourcegraph/ignite-ubuntu:insiders',
                ],
                startTime: '2021-08-02T17:10:13+02:00',
                exitCode: 0,
                durationMilliseconds: 4893,
                out:
                    'stdout: time="2021-08-02T17:10:14+02:00" level=info msg="Created VM with ID \\"d966c3891988743e\\" and name \\"USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b\\""\nstdout: time="2021-08-02T17:10:15+02:00" level=info msg="Networking is handled by \\"docker-bridge\\""\nstdout: time="2021-08-02T17:10:15+02:00" level=info msg="Started Firecracker VM \\"d966c3891988743e\\" in a container with ID \\"ff215e4c82b2e9c984148815f093b4b258b229681ffad9551af2d10195af699a\\""\nstdout: time="2021-08-02T17:10:15+02:00" level=info msg="Waiting for the ssh daemon within the VM to start..."\n',
            },
        ],
        srcPreview: {
            key: 'step.src.0',
            command: [
                'ignite',
                'exec',
                'USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b',
                '--',
                'cd /work && SRC_ENDPOINT=http://USERNAME_REMOVED:PASSWORD_REMOVED@192.168.1.34:3080 SRC_ACCESS_TOKEN=SRC_ACCESS_TOKEN_REMOVED HOME=/home/mrnugget PATH=/home/mrnugget/google-cloud-sdk/bin:/home/mrnugget/bin:/home/mrnugget/.yarn/bin:/home/mrnugget/.config/yarn/global/node_modules/.bin:/usr/local/heroku/bin:/home/mrnugget/code/go/bin:/home/mrnugget/.asdf/shims:/usr/local/opt/asdf/bin:/usr/local/bin:/home/mrnugget/.cargo/bin:/home/mrnugget/.local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/usr/local/go/bin:/home/mrnugget/.fzf/bin src batch preview -f spec.yml -text-only -skip-errors -n mrnugget',
            ],
            startTime: '2021-08-02T17:10:18+02:00',
            exitCode: 0,
            durationMilliseconds: 15718,
            out:
                'stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-08-02T15:10:18.956Z","status":"STARTED"}\nstdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-08-02T15:10:18.957Z","status":"SUCCESS"}\nstdout: {"operation":"RESOLVING_NAMESPACE","timestamp":"2021-08-02T15:10:18.957Z","status":"STARTED"}\nstdout: {"operation":"RESOLVING_NAMESPACE","timestamp":"2021-08-02T15:10:18.962Z","status":"SUCCESS","message":"Namespace: VXNlcjox"}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2021-08-02T15:10:18.962Z","status":"STARTED"}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2021-08-02T15:10:18.962Z","status":"PROGRESS","message":"0% done"}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2021-08-02T15:10:23.724Z","status":"PROGRESS","message":"0% done"}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2021-08-02T15:10:23.724Z","status":"PROGRESS","message":"100% done"}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2021-08-02T15:10:23.724Z","status":"SUCCESS"}\nstdout: {"operation":"DETERMINING_WORKSPACE_TYPE","timestamp":"2021-08-02T15:10:23.724Z","status":"STARTED"}\nstdout: {"operation":"DETERMINING_WORKSPACE_TYPE","timestamp":"2021-08-02T15:10:23.724Z","status":"SUCCESS","message":"BIND"}\nstdout: {"operation":"RESOLVING_REPOSITORIES","timestamp":"2021-08-02T15:10:23.724Z","status":"STARTED"}\nstdout: {"operation":"RESOLVING_REPOSITORIES","timestamp":"2021-08-02T15:10:23.774Z","status":"SUCCESS","message":"1 unsupported repositories"}\nstdout: {"operation":"DETERMINING_WORKSPACES","timestamp":"2021-08-02T15:10:23.774Z","status":"STARTED"}\nstdout: {"operation":"DETERMINING_WORKSPACES","timestamp":"2021-08-02T15:10:23.774Z","status":"SUCCESS","message":"Found 18 workspaces with steps to execute"}\nstdout: {"operation":"CHECKING_CACHE","timestamp":"2021-08-02T15:10:23.774Z","status":"STARTED"}\nstdout: {"operation":"CHECKING_CACHE","timestamp":"2021-08-02T15:10:23.775Z","status":"SUCCESS","message":"Found 0 cached changeset specs; 18 tasks need to be executed"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:23.775Z","status":"STARTED"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:23.775Z","status":"PROGRESS","message":"running: 0, executed: 0, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:24.775Z","status":"PROGRESS","message":"running: 4, executed: 0, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:25.776Z","status":"PROGRESS","message":"running: 4, executed: 3, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:26.775Z","status":"PROGRESS","message":"running: 4, executed: 3, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:27.776Z","status":"PROGRESS","message":"running: 4, executed: 4, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:28.776Z","status":"PROGRESS","message":"running: 4, executed: 6, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:29.775Z","status":"PROGRESS","message":"running: 4, executed: 9, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:30.775Z","status":"PROGRESS","message":"running: 4, executed: 11, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:31.776Z","status":"PROGRESS","message":"running: 4, executed: 13, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:32.776Z","status":"PROGRESS","message":"running: 2, executed: 16, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:33.776Z","status":"PROGRESS","message":"running: 1, executed: 17, built: 0, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:34.243Z","status":"PROGRESS","message":"running: 0, executed: 18, built: 18, errored: 0"}\nstdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-08-02T15:10:34.243Z","status":"SUCCESS"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.243Z","status":"STARTED","message":"Sending 18 changeset specs"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.251Z","status":"PROGRESS","message":"Uploaded 1 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.257Z","status":"PROGRESS","message":"Uploaded 2 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.263Z","status":"PROGRESS","message":"Uploaded 3 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.269Z","status":"PROGRESS","message":"Uploaded 4 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.276Z","status":"PROGRESS","message":"Uploaded 5 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.282Z","status":"PROGRESS","message":"Uploaded 6 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.288Z","status":"PROGRESS","message":"Uploaded 7 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.294Z","status":"PROGRESS","message":"Uploaded 8 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.305Z","status":"PROGRESS","message":"Uploaded 9 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.312Z","status":"PROGRESS","message":"Uploaded 10 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.318Z","status":"PROGRESS","message":"Uploaded 11 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.325Z","status":"PROGRESS","message":"Uploaded 12 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.331Z","status":"PROGRESS","message":"Uploaded 13 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.337Z","status":"PROGRESS","message":"Uploaded 14 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.347Z","status":"PROGRESS","message":"Uploaded 15 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.352Z","status":"PROGRESS","message":"Uploaded 16 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.359Z","status":"PROGRESS","message":"Uploaded 17 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.368Z","status":"PROGRESS","message":"Uploaded 18 out of 18"}\nstdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-08-02T15:10:34.368Z","status":"SUCCESS"}\nstdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-08-02T15:10:34.368Z","status":"STARTED"}\nstdout: {"operation":"CREATING_BATCH_SPEC","timestamp":"2021-08-02T15:10:34.382Z","status":"SUCCESS","message":"http://USERNAME_REMOVED:PASSWORD_REMOVED@192.168.1.34:3080/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiI3QXUza05ORUw4QSI="}\n',
        },
        teardown: [
            {
                key: 'teardown.firecracker.stop',
                command: ['ignite', 'stop', 'USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b'],
                startTime: '2021-08-02T17:10:34+02:00',
                exitCode: 0,
                durationMilliseconds: 2976,
                out:
                    'stdout: time="2021-08-02T17:10:34+02:00" level=info msg="Removing the container with ID \\"ff215e4c82b2e9c984148815f093b4b258b229681ffad9551af2d10195af699a\\" from the \\"docker-bridge\\" network"\nstdout: time="2021-08-02T17:10:37+02:00" level=info msg="Stopped VM with name \\"USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b\\" and ID \\"d966c3891988743e\\""\n',
            },
            {
                key: 'teardown.firecracker.remove',
                command: ['ignite', 'rm', '-f', 'USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b'],
                startTime: '2021-08-02T17:10:37+02:00',
                exitCode: 0,
                durationMilliseconds: 53,
                out:
                    'stdout: time="2021-08-02T17:10:37+02:00" level=info msg="Removed VM with name \\"USERNAME_REMOVED-7bfa6fa0-d1e6-48ef-bec8-aed90784204b\\" and ID \\"d966c3891988743e\\""\n',
            },
        ],
    },
    placeInQueue: null,
    batchSpec: {
        applyURL: '/users/mrnugget/batch-changes/apply/QmF0Y2hTcGVjOiI3QXUza05ORUw4QSI=',
    },
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

const batchSpecExecutionErrored = (): BatchSpecExecutionFields => ({
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

add('Completed', () => (
    <EnterpriseWebStory>
        {props => (
            <BatchSpecExecutionDetailsPage
                {...props}
                executionID="123123"
                fetchBatchSpecExecution={() => of(batchSpecExecutionCompleted())}
            />
        )}
    </EnterpriseWebStory>
))

add('Errored', () => (
    <EnterpriseWebStory>
        {props => (
            <BatchSpecExecutionDetailsPage
                {...props}
                executionID="123123"
                fetchBatchSpecExecution={() => of(batchSpecExecutionErrored())}
            />
        )}
    </EnterpriseWebStory>
))
