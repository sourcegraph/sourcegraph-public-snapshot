import { storiesOf } from '@storybook/react'
import { boolean } from '@storybook/addon-knobs'
import React from 'react'
import { GitCommitNode } from './GitCommitNode'
import { subDays } from 'date-fns'
import { GitCommitFields } from '../../graphql-operations'
import { WebStory } from '../../components/WebStory'

const { add } = storiesOf('web/GitCommitNode', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

export const gitCommitNodeStubData: GitCommitFields = {
    id: 'commit123',
    abbreviatedOID: 'abcdefg',
    oid: 'abcdefghijklmnopqrstuvwxyz12345678904321',
    author: {
        date: subDays(new Date(), 5).toISOString(),
        person: {
            avatarURL: 'https://avatars0.githubusercontent.com/u/19534377?v=4&s=48',
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
            avatarURL: 'https://avatars0.githubusercontent.com/u/19534377?v=4&s=48',
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
        'adasdasd asdaldla sdlas oqfo qov ov qogo vqov wob ifbwrboiwo fijefiow jierierjor iejf aofk sodjv irg eaoif j',
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
    message: 'asdqc3ircnr kfk2 k2 gk g23kg 23kg 2kg 2k3g 2kg23ig02gi 9i 90sug wug98 sgu9 w9 guwij  ',
}

add('Full customizable', () => (
    <WebStory>
        {() => (
            <div className="card">
                <GitCommitNode
                    node={gitCommitNodeStubData}
                    compact={boolean('compact', false)}
                    expandCommitMessageBody={boolean('expandCommitMessageBody', false)}
                    showSHAAndParentsRow={boolean('showSHAAndParentsRow', false)}
                    hideExpandCommitMessageBody={boolean('hideExpandCommitMessageBody', false)}
                />
            </div>
        )}
    </WebStory>
))
add('Compact', () => (
    <WebStory>
        {() => (
            <div className="card">
                <GitCommitNode
                    node={gitCommitNodeStubData}
                    compact={true}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                />
            </div>
        )}
    </WebStory>
))
add('Commit message expanded', () => (
    <WebStory>
        {() => (
            <div className="card">
                <GitCommitNode
                    node={gitCommitNodeStubData}
                    compact={false}
                    expandCommitMessageBody={true}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={false}
                />
            </div>
        )}
    </WebStory>
))
add('SHA and parent shown', () => (
    <WebStory>
        {() => (
            <div className="card">
                <GitCommitNode
                    node={gitCommitNodeStubData}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={true}
                    hideExpandCommitMessageBody={false}
                />
            </div>
        )}
    </WebStory>
))
add('Expand commit message btn hidden', () => (
    <WebStory>
        {() => (
            <div className="card">
                <GitCommitNode
                    node={gitCommitNodeStubData}
                    compact={false}
                    expandCommitMessageBody={false}
                    showSHAAndParentsRow={false}
                    hideExpandCommitMessageBody={true}
                />
            </div>
        )}
    </WebStory>
))
