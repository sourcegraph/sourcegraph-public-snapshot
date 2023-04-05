import { Story, Meta } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { SiteConfigurationHistoryItem } from './SiteConfigurationChangeList'

const config: Meta = {
    title: 'web/site-admin/SiteConfigurationHistoryItem',
    component: SiteConfigurationHistoryItem,
}

export default config

export const HistoryItemWithNoAuthor: Story = () => {
    return (
        <WebStory>
            {() => (
                <SiteConfigurationHistoryItem
                    node={{
                        author: null,
                        createdAt: new Date().toISOString(),
                        diff: `--- ID: 122
+++ ID: 123
@@ -330,6 +330,7 @@
    "url": "https://github.com/"
    },
    {
+   "apiScope": "read_api",
    "clientID": "foo",
-   "clientSecret": "REDACTED-DATA-CHUNK-abcd1234",
    "displayName": "GitLab",`,
                        id: '1',
                    }}
                />
            )}
        </WebStory>
    )
}

export const HistoryItemWithAuthor: Story = () => {
    return (
        <WebStory>
            {() => (
                <SiteConfigurationHistoryItem
                    node={{
                        author: {
                            __typename: 'User',
                            id: '1',
                            displayName: 'Jane Doe',
                            username: 'jdoe',
                            avatarURL: null,
                        },
                        createdAt: new Date().toISOString(),
                        diff: `--- ID: 122
+++ ID: 123
@@ -330,6 +330,7 @@
    "url": "https://github.com/"
    },
    {
+   "apiScope": "read_api",
    "clientID": "foo",
-   "clientSecret": "REDACTED-DATA-CHUNK-abcd1234",
    "displayName": "GitLab",`,
                        id: '1',
                    }}
                />
            )}
        </WebStory>
    )
}

export const HistoryItemWithAuthorButNoDisplayName: Story = () => {
    return (
        <WebStory>
            {() => (
                <SiteConfigurationHistoryItem
                    node={{
                        author: {
                            __typename: 'User',
                            id: '1',
                            displayName: null,
                            username: 'jdoe',
                            avatarURL: null,
                        },
                        createdAt: new Date().toISOString(),
                        diff: `--- ID: 122
+++ ID: 123
@@ -330,6 +330,7 @@
    "url": "https://github.com/"
    },
    {
+   "apiScope": "read_api",
    "clientID": "foo",
-   "clientSecret": "REDACTED-DATA-CHUNK-abcd1234",
    "displayName": "GitLab",`,
                        id: '1',
                    }}
                />
            )}
        </WebStory>
    )
}

export const HistoryItemWithAuthorWithAvatarURL: Story = () => {
    return (
        <WebStory>
            {() => (
                <SiteConfigurationHistoryItem
                    node={{
                        author: {
                            __typename: 'User',
                            id: '1',
                            displayName: 'Beyang Liu',
                            username: 'beyang',
                            avatarURL: 'https://avatars2.githubusercontent.com/u/1646931?v=4',
                        },
                        createdAt: new Date().toISOString(),
                        diff: `--- ID: 122
+++ ID: 123
@@ -330,6 +330,7 @@
    "url": "https://github.com/"
    },
    {
+   "apiScope": "read_api",
    "clientID": "foo",
-   "clientSecret": "REDACTED-DATA-CHUNK-abcd1234",
    "displayName": "GitLab",`,
                        id: '1',
                    }}
                />
            )}
        </WebStory>
    )
}
