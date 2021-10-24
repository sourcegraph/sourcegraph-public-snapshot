import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { LSIFUploadState } from '@sourcegraph/shared/src/graphql/schema'

import { WebStory } from '../../../components/WebStory'
import { LsifIndexFields, LSIFIndexState, LsifIndexStepsFields } from '../../../graphql-operations'

import { CodeIntelIndexPage, CodeIntelIndexPageProps } from './CodeIntelIndexPage'

const trim = (value: string) => {
    const firstSignificantLine = value
        .split('\n')
        .map(line => ({ length: line.length, trimmedLength: line.trimStart().length }))
        .find(({ trimmedLength }) => trimmedLength !== 0)
    if (!firstSignificantLine) {
        return value
    }

    const { length, trimmedLength } = firstSignificantLine
    return value
        .split('\n')
        .map(line => line.slice(length - trimmedLength))
        .join('\n')
        .trim()
}

const stepsPrototype: LsifIndexStepsFields = {
    setup: [],
    preIndex: [
        {
            root: 'staging/src/k8s.io/apiextensions-apiserver',
            image: 'sourcegraph/lsif-go:latest',
            commands: ['go mod download'],
            logEntry: null,
        },
    ],
    index: {
        indexerArgs: ['lsif-go', '--no-animation'],
        outfile: null,
        logEntry: null,
    },
    upload: null,
    teardown: [],
}

const completedSteps: LsifIndexStepsFields = {
    setup: [
        {
            key: 'setup.git.init',
            command: ['git', '-C', '/tmp/2150436480', 'init'],
            startTime: '2020-06-15T17:56:01Z',
            exitCode: 0,
            out: trim(`
                stderr: hint: Using 'master' as the name for the initial branch. This default branch name
                stderr: hint: is subject to change. To configure the initial branch name to use in all
                stderr: hint: of your new repositories, which will suppress this warning, call:
                stderr: hint:
                stderr: hint: \tgit config --global init.defaultBranch \\u003cname\\u003e
                stderr: hint:
                stderr: hint: Names commonly chosen instead of 'master' are 'main', 'trunk' and
                stderr: hint: 'development'. The just-created branch can be renamed via this command:
                stderr: hint:
                stderr: hint: \tgit branch -m \\u003cname\\u003e
                stdout: Initialized empty Git repository in /tmp/2150436480/.git/
            `),
            durationMilliseconds: 35,
        },
        {
            key: 'setup.git.fetch',
            command: [
                'git',
                '-C',
                '/tmp/2150436480',
                '-c',
                'protocol.version=2',
                'fetch',
                'https://USERNAME_REMOVED:PASSWORD_REMOVED@sourcegraph.com/.executors/git/github.com/kubernetes/kubernetes',
                '-t',
                '7413f44e59a4a6931bd22e8274d4068bbf28e67e',
            ],
            startTime: '2020-06-15T17:56:01Z',
            exitCode: 0,
            out: trim(`
                stderr: From https://sourcegraph.com/.executors/git/github.com/kubernetes/kubernetes
                stderr:  * branch                    7413f44e59a4a6931bd22e8274d4068bbf28e67e -\\u003e FETCH_HEAD
                stderr:  * [new tag]                 v1.9.0          -\\u003e v1.9.0
                stderr:  * [new tag]                 v1.9.0-alpha.0  -\\u003e v1.9.0-alpha.0
                stderr:  * [new tag]                 v1.9.0-alpha.1  -\\u003e v1.9.0-alpha.1
                stderr:  * [new tag]                 v1.9.0-alpha.2  -\\u003e v1.9.0-alpha.2
                stderr:  * [new tag]                 v1.9.0-alpha.3  -\\u003e v1.9.0-alpha.3
                stderr:  * [new tag]                 v1.9.0-beta.0   -\\u003e v1.9.0-beta.0
                stderr:  * [new tag]                 v1.9.0-beta.1   -\\u003e v1.9.0-beta.1
                stderr:  * [new tag]                 v1.9.0-beta.2   -\\u003e v1.9.0-beta.2
                stderr:  * [new tag]                 v1.9.1          -\\u003e v1.9.1
                stderr:  * [new tag]                 v1.9.1-beta.0   -\\u003e v1.9.1-beta.0
                stderr:  * [new tag]                 v1.9.10         -\\u003e v1.9.10
                stderr:  * [new tag]                 v1.9.10-beta.0  -\\u003e v1.9.10-beta.0
                stderr:  * [new tag]                 v1.9.11         -\\u003e v1.9.11
                stderr:  * [new tag]                 v1.9.11-beta.0  -\\u003e v1.9.11-beta.0
                stderr:  * [new tag]                 v1.9.12-beta.0  -\\u003e v1.9.12-beta.0
                stderr:  * [new tag]                 v1.9.2          -\\u003e v1.9.2
                stderr:  * [new tag]                 v1.9.2-beta.0   -\\u003e v1.9.2-beta.0
                stderr:  * [new tag]                 v1.9.3          -\\u003e v1.9.3
                stderr:  * [new tag]                 v1.9.3-beta.0   -\\u003e v1.9.3-beta.0
                stderr:  * [new tag]                 v1.9.4          -\\u003e v1.9.4
                stderr:  * [new tag]                 v1.9.4-beta.0   -\\u003e v1.9.4-beta.0
                stderr:  * [new tag]                 v1.9.5          -\\u003e v1.9.5
                stderr:  * [new tag]                 v1.9.5-beta.0   -\\u003e v1.9.5-beta.0
                stderr:  * [new tag]                 v1.9.6          -\\u003e v1.9.6
                stderr:  * [new tag]                 v1.9.6-beta.0   -\\u003e v1.9.6-beta.0
                stderr:  * [new tag]                 v1.9.7          -\\u003e v1.9.7
                stderr:  * [new tag]                 v1.9.7-beta.0   -\\u003e v1.9.7-beta.0
                stderr:  * [new tag]                 v1.9.8          -\\u003e v1.9.8
                stderr:  * [new tag]                 v1.9.8-beta.0   -\\u003e v1.9.8-beta.0
                stderr:  * [new tag]                 v1.9.9          -\\u003e v1.9.9
                stderr:  * [new tag]                 v1.9.9-beta.0   -\\u003e v1.9.9-beta.0
            `),
            durationMilliseconds: 67289,
        },
        {
            key: 'setup.git.add-remote',
            command: ['git', '-C', '/tmp/2150436480', 'remote', 'add', 'origin', 'github.com/kubernetes/kubernetes'],
            startTime: '2020-06-15T17:57:09Z',
            exitCode: 0,
            out: '',
            durationMilliseconds: 3,
        },
        {
            key: 'setup.git.checkout',
            command: ['git', '-C', '/tmp/2150436480', 'checkout', '7413f44e59a4a6931bd22e8274d4068bbf28e67e'],
            startTime: '2020-06-15T17:57:09Z',
            exitCode: 0,
            out: trim(`
                stderr: Note: switching to '7413f44e59a4a6931bd22e8274d4068bbf28e67e'.
                stderr:
                stderr: You are in 'detached HEAD' state. You can look around, make experimental
                stderr: changes and commit them, and you can discard any commits you make in this
                stderr: state without impacting any branches by switching back to a branch.
                stderr:
                stderr: If you want to create a new branch to retain commits you create, you may
                stderr: do so (now or later) by using -c with the switch command. Example:
                stderr:
                stderr:   git switch -c \\u003cnew-branch-name\\u003e
                stderr:
                stderr: Or undo this operation with:
                stderr:
                stderr:   git switch -
                stderr:
                stderr: Turn off this advice by setting config variable advice.detachedHead to false
                stderr:
                stderr: HEAD is now at 7413f44e59a Merge remote-tracking branch 'origin/master'
            `),
            durationMilliseconds: 1892,
        },
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
                '14GB',
                '--size',
                '20GB',
                '--copy-files',
                '/tmp/2150436480:/work',
                '--copy-files',
                '/vm-startup.sh:/vm-startup.sh',
                '--ssh',
                '--name',
                'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d',
                'sourcegraph/ignite-ubuntu:insiders',
            ],
            startTime: '2020-06-15T17:57:10Z',
            exitCode: 0,
            out: trim(`
                stdout: time="2020-06-15T17:58:08Z" level=info msg="Created VM with ID \\"115d7346fe5d34bf\\" and name \\"executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d\\""
                stdout: time="2020-06-15T17:58:08Z" level=info msg="Networking is handled by \\"docker-bridge\\""
                stdout: time="2020-06-15T17:58:08Z" level=info msg="Started Firecracker VM \\"115d7346fe5d34bf\\" in a container with ID \\"37006f6f42b38fcd34a3419ab218c1e6ef71a2e02771cdc682b0fdfc821ed647\\""
                stdout: time="2020-06-15T17:58:09Z" level=info msg="Waiting for the ssh daemon within the VM to start..."
            `),
            durationMilliseconds: 68911,
        },
        {
            key: 'setup.startup-script',
            command: ['ignite', 'exec', 'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d', '--', '/vm-startup.sh'],
            startTime: '2020-06-15T17:58:19Z',
            exitCode: 0,
            out: '',
            durationMilliseconds: 2151,
        },
    ],
    preIndex: stepsPrototype.preIndex.map(step => ({
        ...step,
        logEntry: {
            key: 'step.docker.0',
            command: [
                'ignite',
                'exec',
                'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d',
                '--',
                'docker run --rm --cpus 4 --memory 14GB -v /work:/data -w /data/staging/src/k8s.io/apiextensions-apiserver --entrypoint /bin/sh sourcegraph/lsif-go:latest /data/.sourcegraph-executors/879168.0_github.com_kubernetes_kubernetes@7413f44e59a4a6931bd22e8274d4068bbf28e67e.sh',
            ],
            startTime: '2020-06-15T17:58:21Z',
            exitCode: 0,
            out: trim(`
                stderr: Unable to find image 'sourcegraph/lsif-go:latest' locally
                stderr: latest: Pulling from sourcegraph/lsif-go
                stderr: e9afc4f90ab0: Pulling fs layer
                stderr: 989e6b19a265: Pulling fs layer
                stderr: af14b6c2f878: Pulling fs layer
                stderr: 5573c4b30949: Pulling fs layer
                stderr: d4020e2aa747: Pulling fs layer
                stderr: 90833167b3a6: Pulling fs layer
                stderr: 56c4bef2c0bb: Pulling fs layer
                stderr: 20d92d2b70d0: Pulling fs layer
                stderr: 61b557e60a3c: Pulling fs layer
                stderr: 5573c4b30949: Waiting
                stderr: d4020e2aa747: Waiting
                stderr: 90833167b3a6: Waiting
                stderr: 56c4bef2c0bb: Waiting
                stderr: 20d92d2b70d0: Waiting
                stderr: 61b557e60a3c: Waiting
                stderr: 989e6b19a265: Verifying Checksum
                stderr: 989e6b19a265: Download complete
                stderr: af14b6c2f878: Verifying Checksum
                stderr: af14b6c2f878: Download complete
                stderr: e9afc4f90ab0: Verifying Checksum
                stderr: e9afc4f90ab0: Download complete
                stderr: 5573c4b30949: Verifying Checksum
                stderr: 5573c4b30949: Download complete
                stderr: 56c4bef2c0bb: Verifying Checksum
                stderr: 56c4bef2c0bb: Download complete
                stderr: 20d92d2b70d0: Verifying Checksum
                stderr: 20d92d2b70d0: Download complete
                stderr: 61b557e60a3c: Verifying Checksum
                stderr: 61b557e60a3c: Download complete
                stderr: d4020e2aa747: Verifying Checksum
                stderr: d4020e2aa747: Download complete
                stderr: 90833167b3a6: Verifying Checksum
                stderr: 90833167b3a6: Download complete
                stderr: e9afc4f90ab0: Pull complete
                stderr: 989e6b19a265: Pull complete
                stderr: af14b6c2f878: Pull complete
                stderr: 5573c4b30949: Pull complete
                stderr: d4020e2aa747: Pull complete
                stderr: 90833167b3a6: Pull complete
                stderr: 56c4bef2c0bb: Pull complete
                stderr: 20d92d2b70d0: Pull complete
                stderr: 61b557e60a3c: Pull complete
                stderr: Digest: sha256:7c800b356ebb1e968a21312b0cb3e99f390a57b2e57d80a520e4a896752394aa
                stderr: Status: Downloaded newer image for sourcegraph/lsif-go:latest
                stderr: + go mod download
            `),
            durationMilliseconds: 51256,
        },
    })),
    index: {
        ...stepsPrototype.index,
        logEntry: {
            key: 'step.docker.1',
            command: [
                'ignite',
                'exec',
                'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d',
                '--',
                'docker run --rm --cpus 4 --memory 14GB -v /work:/data -w /data/staging/src/k8s.io/apiextensions-apiserver --entrypoint /bin/sh sourcegraph/lsif-go:latest /data/.sourcegraph-executors/879168.1_github.com_kubernetes_kubernetes@7413f44e59a4a6931bd22e8274d4068bbf28e67e.sh',
            ],
            startTime: '2020-06-15T17:59:13Z',
            exitCode: 0,
            out: trim(`
                stderr: + lsif-go --no-animation
                stdout: Resolving module name
                stdout: Listing dependencies
                stdout: Loading packages
                stdout: Emitting documents
                stdout: Adding import definitions
                stdout: Indexing documentation
                stdout: Indexing definitions
                stdout: Indexing references
                stdout: Linking items to definitions
                stdout: Emitting contains relations
            `),
            durationMilliseconds: 56103,
        },
    },
    upload: {
        key: 'step.src.0',
        command: [
            'ignite',
            'exec',
            'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d',
            '--',
            'cd /work/staging/src/k8s.io/apiextensions-apiserver \\u0026\\u0026 SRC_ENDPOINT=https://USERNAME_REMOVED:PASSWORD_REMOVED@sourcegraph.com src lsif upload -no-progress -repo github.com/kubernetes/kubernetes -commit 7413f44e59a4a6931bd22e8274d4068bbf28e67e -root staging/src/k8s.io/apiextensions-apiserver -upload-route /.executors/lsif/upload -file dump.lsif -associated-index-id 879168',
        ],
        startTime: '2020-06-15T18:00:09Z',
        exitCode: 0,
        out: trim(`
            stdout: \\u001b[?25l💡 Inferred arguments
            stdout: \\u001b[?25h\\u001b[?25l   \\u001b[?25lrepo: github.com/kubernetes/kubernetes
            stdout:    \\u001b[?25h\\u001b[?25lcommit: 7413f44e59a4a6931bd22e8274d4068bbf28e67e
            stdout:    \\u001b[?25h\\u001b[?25lroot: staging/src/k8s.io/apiextensions-apiserver
            stdout:    \\u001b[?25h\\u001b[?25lfile: dump.lsif
            stdout:    \\u001b[?25h\\u001b[?25lindexer: lsif-go
            stdout:    \\u001b[?25h
            stdout: \\u001b[?25h\\u001b[?25l💡 View processing status at https://sourcegraph.com/github.com/kubernetes/kubernetes/-/code-intelligence/uploads/TFNJRlVwbG9hZDoiNTUyNzg3Ig==
            stdout: \\u001b[?25h
        `),
        durationMilliseconds: 1265,
    },
    teardown: [
        {
            key: 'teardown.firecracker.remove',
            command: ['ignite', 'rm', '-f', 'executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d'],
            startTime: '2020-06-15T18:00:10Z',
            exitCode: 0,
            out: trim(`
                stdout: time="2020-06-15T18:00:10Z" level=info msg="Removing the container with ID \\"37006f6f42b38fcd34a3419ab218c1e6ef71a2e02771cdc682b0fdfc821ed647\\" from the \\"docker-bridge\\" network"
                stdout: time="2020-06-15T18:00:16Z" level=info msg="Removed VM with name \\"executors-9a1b6d36-45a5-4986-a1d0-5cf17424788d\\" and ID \\"115d7346fe5d34bf\\""
            `),
            durationMilliseconds: 6126,
        },
    ],
}

const processingSteps: LsifIndexStepsFields = {
    ...completedSteps,
    upload: {
        ...completedSteps.upload!,
        exitCode: null,
        out: trim(`
            stdout: \\u001b[?25l💡 Inferred arguments
            stdout: \\u001b[?25h\\u001b[?25l   \\u001b[?25lrepo: github.com/kubernetes/kubernetes
            stdout:    \\u001b[?25h\\u001b[?25lcommit: 7413f44e59a4a6931bd22e8274d4068bbf28e67e
            stdout:    \\u001b[?25h\\u001b[?25lroot: staging/src/k8s.io/apiextensions-apiserver
            stdout:    \\u001b[?25h\\u001b[?25lfile: dump.lsif
            stdout:    \\u001b[?25h\\u001b[?25lindexer: lsif-go
            stdout:    \\u001b[?25h
                stdout: \\u001b[?25h\\u001b[?25l💡 Uploading...
                stdout: \\u001b[?25h
        `),
        durationMilliseconds: null,
    },
    teardown: [],
}

const failedSteps: LsifIndexStepsFields = {
    ...completedSteps,
    upload: {
        ...completedSteps.upload!,
        exitCode: 127,
        out: `${
            completedSteps.upload?.out ?? ''
        }\nstderr: Failed to complete: dial tcp: lookup gitserver-8.gitserver on 10.165.0.10:53: no such host\n`,
    },
    teardown: [],
}

const indexPrototype: Omit<LsifIndexFields, 'id' | 'state' | 'queuedAt' | 'steps'> = {
    __typename: 'LSIFIndex',
    inputCommit: '',
    inputRoot: 'staging/src/k8s.io/apiextensions-apiserver/',
    inputIndexer: 'sourcegraph/lsif-go:latest',
    failure: null,
    startedAt: null,
    finishedAt: null,
    placeInQueue: null,
    projectRoot: {
        url: '',
        path: 'staging/src/k8s.io/apiextensions-apiserver/',
        repository: {
            url: '',
            name: 'github.com/kubernetes/kubernetes',
        },
        commit: {
            url: '',
            oid: '7413f44e59a4a6931bd22e8274d4068bbf28e67e',
            abbreviatedOID: '7413f44',
        },
    },
    associatedUpload: null,
}

const now = () => new Date('2020-06-15T19:25:00+00:00')

const story: Meta = {
    title: 'web/codeintel/detail/CodeIntelIndexPage',
    decorators: [story => <div className="p-3 container">{story()}</div>],
    parameters: {
        component: CodeIntelIndexPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelIndexPageProps> = args => (
    <WebStory>{props => <CodeIntelIndexPage {...props} {...args} />}</WebStory>
)

const defaults: Partial<CodeIntelIndexPageProps> = {
    now,
}

export const Queued = Template.bind({})
Queued.args = {
    ...defaults,
    queryLisfIndex: () =>
        of({
            ...indexPrototype,
            id: '1',
            state: LSIFIndexState.QUEUED,
            queuedAt: '2020-06-15T17:50:01+00:00',
            placeInQueue: 1,
            steps: stepsPrototype,
        }),
}

export const Processing = Template.bind({})
Processing.args = {
    ...defaults,
    queryLisfIndex: () =>
        of({
            ...indexPrototype,
            id: '1',
            state: LSIFIndexState.PROCESSING,
            queuedAt: '2020-06-15T17:50:01+00:00',
            startedAt: '2020-06-15T17:56:01+00:00',
            steps: processingSteps,
        }),
}

export const Completed = Template.bind({})
Completed.args = {
    ...defaults,
    queryLisfIndex: () =>
        of({
            ...indexPrototype,
            id: '1',
            state: LSIFIndexState.COMPLETED,
            queuedAt: '2020-06-15T17:50:01+00:00',
            startedAt: '2020-06-15T17:56:01+00:00',
            finishedAt: '2020-06-15T18:00:10+00:00',
            steps: completedSteps,
        }),
}

export const Errored = Template.bind({})
Errored.args = {
    ...defaults,
    queryLisfIndex: () =>
        of({
            ...indexPrototype,
            id: '1',
            state: LSIFIndexState.ERRORED,
            queuedAt: '2020-06-15T17:50:01+00:00',
            startedAt: '2020-06-15T17:56:01+00:00',
            finishedAt: '2020-06-15T18:00:10+00:00',
            failure:
                'Upload failed to complete: dial tcp: lookup gitserver-8.gitserver on 10.165.0.10:53: no such host',
            steps: failedSteps,
        }),
}

export const AssociatedUpload = Template.bind({})
AssociatedUpload.args = {
    ...defaults,
    queryLisfIndex: () =>
        of({
            ...indexPrototype,
            id: '1',
            state: LSIFIndexState.COMPLETED,
            queuedAt: '2020-06-15T17:50:01+00:00',
            startedAt: '2020-06-15T17:56:01+00:00',
            finishedAt: '2020-06-15T18:00:10+00:00',
            steps: completedSteps,
            associatedUpload: {
                id: '2',
                state: LSIFUploadState.COMPLETED,
                uploadedAt: '2020-06-15T18:00:10+00:00',
                startedAt: '2020-06-15T18:05:00+00:00',
                finishedAt: '2020-06-15T18:10:00+00:00',
                placeInQueue: null,
            },
        }),
}
