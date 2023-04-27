import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { PerforceDepotChangelistNode } from './PerforceDepotChangelistNode'

const config: Meta = {
    title: 'web/PerforceDepotChangeListNode',
    component: PerforceDepotChangelistNode,
}

export default config

const node = {
    id: '123456',
    oid: '123456',
    abbreviatedOID: '123456',
    message: '123456 - test change\n[p4-fusion: depot-paths = "//go/": change = 123456]',
    subject: 'test change',
    body: null,
    author: {
        person: {
            avatarURL: null,
            name: 'admin',
            email: 'admin@perforce-server',
            displayName: 'admin',
            user: null,
            __typename: 'Person',
        },
        date: '2021-08-13T19:24:59Z',
        __typename: 'Signature',
    },
    committer: {
        person: {
            avatarURL: null,
            name: 'admin',
            email: 'admin@perforce-server',
            displayName: 'admin',
            user: null,
            __typename: 'Person',
        },
        date: '2021-08-13T19:24:59Z',
        __typename: 'Signature',
    },
    parents: [
        {
            oid: 'd9994dc548fd79b473ce05198c88282890983fa9',
            abbreviatedOID: 'd9994dc',
            url: '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
            __typename: 'GitCommit',
        },
    ],
    url: '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
    canonicalURL: '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
    externalURLs: [],
    tree: {
        canonicalURL: '/perforce.sgdev.org/go@a77d1b309bce5dbb63581bca8c2caac013ec9387',
        __typename: 'GitTree',
    },
}

export const PerforceDepotChangelistItem: Story = () => (
    <WebStory>{() => <PerforceDepotChangelistNode node={node} />}</WebStory>
)

PerforceDepotChangelistItem.storyName = 'Perforce changelist item'
