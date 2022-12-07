import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import {
    ExecutionLogEntryFields,
    LsifIndexFields,
    LSIFIndexState,
    LsifUploadFields,
    LSIFUploadState,
    PreciseSupportLevel,
    SearchBasedSupportLevel,
} from '../../../../graphql-operations'
import { UseCodeIntelStatusPayload, UseRequestLanguageSupportParameters } from '../hooks/useCodeIntelStatus'

import { CodeIntelligenceBadgeContentProps } from './CodeIntelligenceBadgeContent'
import { CodeIntelligenceBadgeMenu } from './CodeIntelligenceBadgeMenu'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/codeintel/enterprise/CodeIntelligenceBadgeMenu',
    decorators: [decorator],
}

export default config

const uploadPrototype: Omit<LsifUploadFields, 'id' | 'state' | 'uploadedAt'> = {
    __typename: 'LSIFUpload',
    inputCommit: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
    tags: [],
    inputRoot: 'web/',
    inputIndexer: 'scip-typescript',
    indexer: { name: 'scip-typescript', url: 'https://github.com/sourcegraph/scip-typescript' },
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
    auditLogs: [],
}

const executionLogPrototype: ExecutionLogEntryFields = {
    key: 'log',
    command: ['lsif-go', '-v'],
    startTime: '2020-06-15T15:25:00+00:00',
    exitCode: 0,
    out: 'foo\nbar\baz\n',
    durationMilliseconds: 123456,
}

const indexPrototype: Omit<LsifIndexFields, 'id' | 'state' | 'queuedAt'> = {
    __typename: 'LSIFIndex',
    inputCommit: '',
    tags: [],
    inputRoot: 'web/',
    inputIndexer: 'scip-typescript',
    indexer: { name: 'scip-typescript', url: 'https://github.com/sourcegraph/scip-typescript' },
    failure: null,
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
    steps: {
        setup: [executionLogPrototype],
        preIndex: [
            { root: '/', image: 'node:alpine', commands: ['yarn'], logEntry: executionLogPrototype },
            { root: '/web', image: 'node:alpine', commands: ['yarn'], logEntry: executionLogPrototype },
        ],
        index: {
            indexerArgs: ['-p', '.'],
            outfile: 'index.lsif',
            logEntry: executionLogPrototype,
        },
        upload: executionLogPrototype,
        teardown: [executionLogPrototype],
    },
    associatedUpload: null,
    shouldReindex: false,
}

const completedUpload = {
    root: 'lib/',
    indexer: { name: 'scip-typescript', url: 'https://github.com/sourcegraph/scip-typescript' },
    uploads: [
        {
            ...uploadPrototype,
            id: '41',
            state: LSIFUploadState.COMPLETED,
            uploadedAt: '2020-06-15T15:20:24+00:00',
            finishedAt: '2020-06-15T15:22:37+00:00',
        },
    ],
}

const failingUpload = {
    root: 'client/',
    indexer: { name: 'scip-typescript', url: 'https://github.com/sourcegraph/scip-typescript' },
    uploads: [
        {
            ...uploadPrototype,
            id: '42',
            state: LSIFUploadState.ERRORED,
            uploadedAt: '2020-06-15T15:20:24+00:00',
            finishedAt: '2020-06-15T15:22:37+00:00',
        },
        {
            ...uploadPrototype,
            id: '43',
            state: LSIFUploadState.QUEUED,
            uploadedAt: '2020-06-15T15:23:47+00:00',
        },
        {
            ...uploadPrototype,
            id: '44',
            state: LSIFUploadState.UPLOADING,
            uploadedAt: '2020-06-15T15:24:59+00:00',
        },
    ],
}

const failingIndex = {
    root: 'client/',
    indexer: { name: 'scip-typescript', url: 'https://github.com/sourcegraph/scip-typescript' },
    indexes: [
        {
            ...indexPrototype,
            id: '42',
            state: LSIFIndexState.ERRORED,
            queuedAt: '2020-06-15T15:20:24+00:00',
            finishedAt: '2020-06-15T15:22:37+00:00',
        },
        {
            ...indexPrototype,
            id: '43',
            state: LSIFIndexState.QUEUED,
            queuedAt: '2020-06-15T15:23:47+00:00',
        },
        {
            ...indexPrototype,
            id: '44',
            state: LSIFIndexState.QUEUED,
            queuedAt: '2020-06-15T15:24:59+00:00',
        },
    ],
}

const preciseSupport = [
    {
        supportLevel: PreciseSupportLevel.NATIVE,
        indexers: [
            {
                name: 'lsif-go',
                url: 'https://github.com/sourcegraph/lsif-go',
            },
        ],
    },
]

const multiplePreciseSupport = [
    {
        supportLevel: PreciseSupportLevel.NATIVE,
        indexers: [
            {
                name: 'lsif-go',
                url: 'https://github.com/sourcegraph/lsif-go',
            },
        ],
    },
    {
        supportLevel: PreciseSupportLevel.NATIVE,
        indexers: [
            {
                name: 'scip-typescript',
                url: 'https://github.com/sourcegraph/scip-typescript',
            },
            {
                // Note: not shown
                name: 'lsif-node',
                url: 'https://github.com/microsoft/lsif-node',
            },
        ],
    },
    {
        // Note: not show as non-native
        supportLevel: PreciseSupportLevel.THIRD_PARTY,
        indexers: [
            {
                name: 'rust-analyzer',
                url: '',
            },
        ],
    },
]

const searchBasedSupport = [
    {
        language: 'Perl',
        supportLevel: SearchBasedSupportLevel.BASIC,
    },
]

const emptyPayload: UseCodeIntelStatusPayload = {
    availableIndexers: [],
    activeUploads: [],
    recentUploads: [],
    recentIndexes: [],
    preciseSupport: [],
    searchBasedSupport: [],
}

const now = () => new Date('2020-06-15T15:25:00+00:00')

const defaultProps: CodeIntelligenceBadgeContentProps = {
    repoName: 'repoName',
    revision: 'commitID',
    filePath: 'foo/bar/baz.bonk',
    settingsCascade: { subjects: null, final: null },
    isStorybook: true,
    now,
    useCodeIntelStatus: () => ({ data: emptyPayload, loading: false }),
    useRequestedLanguageSupportQuery: () => ({
        data: { languages: ['ocaml'] },
        loading: false,
        error: undefined,
    }),
    useRequestLanguageSupportQuery: ({ onCompleted }: UseRequestLanguageSupportParameters) => [
        () =>
            Promise.resolve({ data: {}, loading: false }).then(value => {
                if (onCompleted) {
                    onCompleted()
                }
                return value
            }),
        { loading: false },
    ],
}

const withPayload = (payload: Partial<UseCodeIntelStatusPayload>): typeof defaultProps => ({
    ...defaultProps,
    useCodeIntelStatus: () => ({ data: { ...emptyPayload, ...payload }, loading: false }),
})

export const Unsupported: Story = () => <CodeIntelligenceBadgeMenu {...defaultProps} />
export const Unavailable: Story = () => <CodeIntelligenceBadgeMenu {...withPayload({ searchBasedSupport })} />

export const MultipleProjects: Story = () => (
    <CodeIntelligenceBadgeMenu {...withPayload({ preciseSupport: multiplePreciseSupport })} />
)

MultipleProjects.storyName = 'Multiple projects'

export const MultipleProjectsOneEnabled: Story = () => (
    <CodeIntelligenceBadgeMenu {...withPayload({ recentUploads: [completedUpload], preciseSupport })} />
)

MultipleProjectsOneEnabled.storyName = 'Multiple projects, one enabled'

export const ProcessingError: Story = () => (
    <CodeIntelligenceBadgeMenu {...withPayload({ recentUploads: [completedUpload, failingUpload] })} />
)

ProcessingError.storyName = 'Processing error'

export const IndexingError: Story = () => (
    <CodeIntelligenceBadgeMenu {...withPayload({ recentUploads: [completedUpload], recentIndexes: [failingIndex] })} />
)

IndexingError.storyName = 'Indexing error'

export const MultipleErrors: Story = () => (
    <CodeIntelligenceBadgeMenu
        {...withPayload({ recentUploads: [completedUpload, failingUpload], recentIndexes: [failingIndex] })}
    />
)

MultipleErrors.storyName = 'Multiple errors'
