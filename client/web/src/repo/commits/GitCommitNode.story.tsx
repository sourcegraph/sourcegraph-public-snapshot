import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'

import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import { GitCommitFields } from '../../graphql-operations'

import { GitCommitNode } from './GitCommitNode'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/GitCommitNode',
    parameters: { disableSnapshot: false },
    decorators: [decorator],
}

export default config

const gitCommitNode: GitCommitFields = {
    id: 'commit123',
    abbreviatedOID: 'abcdefg',
    oid: 'abcdefghijklmnopqrstuvwxyz12345678904321',
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
    body:
        'Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.',
    parents: [
        {
            abbreviatedOID: '987654',
            oid: '98765432101234abcdefghijklmnopqrstuvwxyz',
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

export const FullCustomizable: Story = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={boolean('compact', false)}
                    expandCommitMessageBody={boolean('expandCommitMessageBody', false)}
                    showSHAAndParentsRow={boolean('showSHAAndParentsRow', false)}
                    hideExpandCommitMessageBody={boolean('hideExpandCommitMessageBody', false)}
                    preferAbsoluteTimestamps={boolean('preferAbsoluteTimestamps', false)}
                />
            </Card>
        )}
    </WebStory>
)

FullCustomizable.storyName = 'Full customizable'

export const Compact: Story = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={true}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                />
            </Card>
        )}
    </WebStory>
)

export const CommitMessageExpand: Story = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={true}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                />
            </Card>
        )}
    </WebStory>
)

CommitMessageExpand.storyName = 'Commit message expanded'

export const SHAAndParentShown: Story = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={true}
                    hideExpandCommitMessageBody={false}
                />
            </Card>
        )}
    </WebStory>
)

SHAAndParentShown.storyName = 'SHA and parent shown'

export const ExpandCommitMessageButtonHidden: Story = () => (
    <WebStory>
        {() => (
            <Card>
                <GitCommitNode
                    node={gitCommitNode}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={true}
                />
            </Card>
        )}
    </WebStory>
)

ExpandCommitMessageButtonHidden.storyName = 'Expand commit message btn hidden'
