import { subDays, subHours, subMinutes } from 'date-fns'

import {
    BatchSpecWorkspaceState,
    BatchSpecWorkspaceStepFields,
    ChangesetSpecType,
    HiddenBatchSpecWorkspaceFields,
    VisibleBatchSpecWorkspaceFields,
} from '../../../graphql-operations'

const now = new Date()

export const mockStep = (
    number: number,
    step?: Partial<BatchSpecWorkspaceStepFields>
): BatchSpecWorkspaceStepFields => ({
    __typename: 'BatchSpecWorkspaceStep',
    cachedResultFound: false,
    container: 'ubuntu:18.04',
    diffStat: { __typename: 'DiffStat', added: 10, changed: 5, deleted: 5 },
    environment: [],
    exitCode: 0,
    finishedAt: subMinutes(now, 1).toISOString(),
    ifCondition: null,
    number,
    outputLines: ['stdout: Hello World', 'stdout: '],
    outputVariables: [],
    run: `echo Hello World Step ${number} | tee -a $(find -name README.md)`,
    skipped: false,
    startedAt: subMinutes(now, 2).toISOString(),
    ...step,
})

export const mockWorkspace = (
    numberOfSteps = 1,
    workspace?: Partial<VisibleBatchSpecWorkspaceFields>
): VisibleBatchSpecWorkspaceFields => ({
    __typename: 'VisibleBatchSpecWorkspace',
    id: 'test-1234',
    state: BatchSpecWorkspaceState.COMPLETED,
    searchResultPaths: ['/some/path'],
    queuedAt: subHours(now, 1).toISOString(),
    startedAt: subHours(now, 1).toISOString(),
    finishedAt: now.toISOString(),
    failureMessage: null,
    placeInQueue: null,
    path: '/some/path',
    onlyFetchWorkspace: false,
    ignored: false,
    unsupported: false,
    cachedResultFound: false,
    steps: new Array(numberOfSteps).fill(0).map((_item, index) => mockStep(index + 1)),
    changesetSpecs: [
        {
            description: {
                baseRef: 'main',
                baseRepository: {
                    __typename: 'Repository',
                    name: 'github.com/sourcegraph-testing/batch-changes-test-repo',
                    url: '/github.com/sourcegraph-testing/batch-changes-test-repo',
                },
                body: 'My first batch change!',
                diffStat: { __typename: 'DiffStat', added: 100, changed: 50, deleted: 90 },
                headRef: 'hello-world',
                published: null,
                title: 'Hello World',
                __typename: 'GitBranchChangesetDescription',
            },
            id: 'test-1234',
            type: ChangesetSpecType.BRANCH,
            __typename: 'VisibleChangesetSpec',
        },
    ],
    diffStat: { __typename: 'DiffStat', added: 100, changed: 50, deleted: 90, ...workspace?.diffStat },
    stages: {
        __typename: 'BatchSpecWorkspaceStages',
        setup: [
            {
                command: [],
                durationMilliseconds: 0,
                exitCode: 0,
                key: 'setup.fs',
                out: '',
                startTime: subMinutes(now, 10).toISOString(),
                __typename: 'ExecutionLogEntry',
            },
        ],
        srcExec: {
            command: ['src', 'batch', 'exec', '-f', 'input.json'],
            durationMilliseconds: null,
            exitCode: null,
            key: 'step.src.0',
            out:
                'stdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2022-04-21T06:26:59.055Z","status":"STARTED","metadata":{}}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2022-04-21T06:26:59.055Z","status":"PROGRESS","metadata":{"total":1}}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2022-04-21T06:26:59.188Z","status":"PROGRESS","metadata":{"done":1,"total":1}}\nstdout: {"operation":"PREPARING_DOCKER_IMAGES","timestamp":"2022-04-21T06:26:59.188Z","status":"SUCCESS","metadata":{}}\nstdout: {"operation":"DETERMINING_WORKSPACE_TYPE","timestamp":"2022-04-21T06:26:59.188Z","status":"STARTED","metadata":{}}\n',
            startTime: subMinutes(now, 10).toISOString(),
            __typename: 'ExecutionLogEntry',
        },
        teardown: [],
        ...workspace?.stages,
    },
    executor: {
        active: true,
        architecture: 'arm64',
        dockerVersion: '20.10.12',
        executorVersion: '0.0.0+dev',
        firstSeenAt: subDays(now, 10).toISOString(),
        gitVersion: '2.35.1',
        hostname: 'some-super-long-hostname.at-some-address.id-123450123123902304723749827498724',
        id: 'test-1234',
        igniteVersion: '',
        lastSeenAt: subMinutes(now, 2).toISOString(),
        os: 'darwin',
        queueName: 'batches',
        srcCliVersion: '3.38.0',
        __typename: 'Executor',
        ...workspace?.executor,
    },
    ...workspace,
    repository: {
        __typename: 'Repository',
        name: 'github.com/sourcegraph-testing/batch-changes-test-repo',
        url: '/github.com/sourcegraph-testing/batch-changes-test-repo',
        ...workspace?.repository,
    },
    branch: {
        __typename: 'GitRef',
        displayName: 'main',
        ...workspace?.branch,
    },
})

export const QUEUED_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.QUEUED,
    placeInQueue: 2,
    startedAt: null,
    finishedAt: null,
    diffStat: null,
    changesetSpecs: [],
})

export const PROCESSING_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.PROCESSING,
    finishedAt: null,
    diffStat: null,
    changesetSpecs: [],
})

export const SKIPPED_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.SKIPPED,
    queuedAt: null,
    startedAt: null,
    finishedAt: null,
    stages: null,
    executor: null,
    diffStat: null,
    changesetSpecs: null,
    ignored: true,
})

export const UNSUPPORTED_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.SKIPPED,
    queuedAt: null,
    startedAt: null,
    finishedAt: null,
    stages: null,
    executor: null,
    diffStat: null,
    changesetSpecs: null,
    unsupported: true,
})

export const LOTS_OF_STEPS_WORKSPACE = mockWorkspace(20)

export const HIDDEN_WORKSPACE: HiddenBatchSpecWorkspaceFields = {
    __typename: 'HiddenBatchSpecWorkspace',
    id: 'id123',
    queuedAt: subMinutes(new Date(), 10).toISOString(),
    startedAt: subMinutes(new Date(), 8).toISOString(),
    finishedAt: subMinutes(new Date(), 2).toISOString(),
    state: BatchSpecWorkspaceState.COMPLETED,
    diffStat: {
        __typename: 'DiffStat',
        added: 10,
        changed: 2,
        deleted: 5,
    },
    placeInQueue: null,
    onlyFetchWorkspace: false,
    ignored: false,
    unsupported: false,
    cachedResultFound: false,
}

export const FAILED_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.FAILED,
    failureMessage: 'failed to perform src-cli step: command failed',
    steps: [mockStep(1), mockStep(2, { exitCode: 1, diffStat: null })],
})

export const CANCELING_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.CANCELING,
    finishedAt: null,
    diffStat: null,
    changesetSpecs: [],
})

export const CANCELED_WORKSPACE = mockWorkspace(1, {
    state: BatchSpecWorkspaceState.CANCELED,
    finishedAt: null,
    diffStat: null,
    changesetSpecs: [],
})
