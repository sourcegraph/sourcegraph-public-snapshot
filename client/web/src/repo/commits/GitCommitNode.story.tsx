import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { subDays } from 'date-fns'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import type { GitCommitFields } from '../../graphql-operations'

import { GitCommitNode } from './GitCommitNode'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

window.context.experimentalFeatures = { perforceChangelistMapping: 'enabled' }

const config: Meta = {
    title: 'web/GitCommitNode',
    parameters: {},
    decorators: [decorator],
}

export default config

const gitCommitNode: GitCommitFields = {
    id: 'commit123',
    abbreviatedOID: 'abcdefg',
    oid: 'abcdefghijklmnopqrstuvwxyz12345678904321',
    perforceChangelist: null,
    author: {
        date: subDays(new Date(), 5).toISOString(),
        person: {
            avatarURL: 'http://test.test/useravatar',
            displayName: 'alice',
            email: 'alice@sourcegraph.com',
            name: 'Alice',
            user: {
                id: 'alice123',
                url: '/users/alice',
                displayName: 'Alice',
                username: 'alice',
            },
        },
    },
    committer: {
        date: subDays(new Date(), 5).toISOString(),
        person: {
            avatarURL: 'http://test.test/useravatar',
            displayName: 'alice',
            email: 'alice@sourcegraph.com',
            name: 'Alice',
            user: {
                id: 'alice123',
                url: '/users/alice',
                displayName: 'Alice',
                username: 'alice',
            },
        },
    },
    body: 'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.',
    parents: [
        {
            abbreviatedOID: '987654',
            oid: '98765432101234abcdefghijklmnopqrstuvwxyz',
            perforceChangelist: null,
            url: '/commits/987654',
        },
    ],
    subject: 'Super awesome commit',
    url: '/commits/abcdefg',
    tree: null,
    canonicalURL: 'asd',
    externalURLs: [],
    message:
        'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore.',
}

export const FullCustomizable: StoryFn = args => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={args.compact}
                    expandCommitMessageBody={args.expandCommitMessageBody}
                    showSHAAndParentsRow={args.showSHAAndParentsRow}
                    hideExpandCommitMessageBody={args.hideExpandCommitMessageBody}
                    preferAbsoluteTimestamps={args.preferAbsoluteTimestamps}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)
FullCustomizable.argTypes = {
    compact: {
        control: { type: 'boolean' },
    },
    expandCommitMessageBody: {
        control: { type: 'boolean' },
    },
    showSHAAndParentsRow: {
        control: { type: 'boolean' },
    },
    hideExpandCommitMessageBody: {
        control: { type: 'boolean' },
    },
    preferAbsoluteTimestamps: {
        control: { type: 'boolean' },
    },
}
FullCustomizable.args = {
    compact: false,
    expandCommitMessageBody: false,
    showSHAAndParentsRow: false,
    hideExpandCommitMessageBody: false,
    preferAbsoluteTimestamps: false,
}

FullCustomizable.storyName = 'Full customizable'

export const Compact: StoryFn = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={true}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)

export const CommitMessageExpand: StoryFn = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={true}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)

CommitMessageExpand.storyName = 'Commit message expanded'

export const SHAAndParentShown: StoryFn = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={true}
                    hideExpandCommitMessageBody={false}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)

SHAAndParentShown.storyName = 'SHA and parent shown'

export const ExpandCommitMessageButtonHidden: StoryFn = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={true}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)

ExpandCommitMessageButtonHidden.storyName = 'Expand commit message btn hidden'

const perforceChangelistNode: GitCommitFields = {
    id: 'commit123',
    abbreviatedOID: 'abcdefg',
    oid: 'abcdefghijklmnopqrstuvwxyz12345678904321',
    perforceChangelist: {
        __typename: 'PerforceChangelist',
        cid: '12345',
        canonicalURL: '/go/-/changelist/12345',
    },
    author: {
        date: subDays(new Date(), 5).toISOString(),
        person: {
            avatarURL: 'http://test.test/useravatar',
            displayName: 'alice',
            email: 'alice@sourcegraph.com',
            name: 'Alice',
            user: {
                id: 'alice123',
                url: '/users/alice',
                displayName: 'Alice',
                username: 'alice',
            },
        },
    },
    committer: {
        date: subDays(new Date(), 5).toISOString(),
        person: {
            avatarURL: 'http://test.test/useravatar',
            displayName: 'alice',
            email: 'alice@sourcegraph.com',
            name: 'Alice',
            user: {
                id: 'alice123',
                url: '/users/alice',
                displayName: 'Alice',
                username: 'alice',
            },
        },
    },
    body: 'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.',
    parents: [
        {
            abbreviatedOID: '987654',
            oid: '98765432101234abcdefghijklmnopqrstuvwxyz',
            perforceChangelist: {
                __typename: 'PerforceChangelist',
                cid: '12344',
                canonicalURL: '/go/-/changelist/12344',
            },
            url: '/commits/987654',
        },
    ],
    subject: 'Super awesome commit',
    url: '/commits/abcdefg',
    tree: null,
    canonicalURL: 'asd',
    externalURLs: [],
    message:
        'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore.',
}

export const PerforceChangelist: StoryFn = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={perforceChangelistNode}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={true}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </Card>
        )}
    </WebStory>
)

PerforceChangelist.storyName = 'Perforce changelist'
