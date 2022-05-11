import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'

import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import { GitCommitFields } from '../../graphql-operations'

import { GitCommitNode } from './GitCommitNode'

const { add } = storiesOf('web/GitCommitNode', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({ disableSnapshot: false })

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

add('Full customizable', () => (
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
))
add('Compact', () => (
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
))
add('Commit message expanded', () => (
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
))
add('SHA and parent shown', () => (
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
))
add('Expand commit message btn hidden', () => (
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
))
