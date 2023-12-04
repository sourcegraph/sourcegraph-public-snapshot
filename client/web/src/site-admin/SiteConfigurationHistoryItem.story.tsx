import type { StoryFn, Meta } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { SiteConfigurationHistoryItem } from './SiteConfigurationChangeList'

const config: Meta = {
    title: 'web/site-admin/SiteConfigurationHistoryItem',
    component: SiteConfigurationHistoryItem,
}

export default config

export const HistoryItemWithNoAuthor: StoryFn = () => (
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

HistoryItemWithNoAuthor.storyName = 'History item with no author'

export const HistoryItemWithAuthor: StoryFn = () => (
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

HistoryItemWithAuthor.storyName = 'History item with author'

export const HistoryItemWithAuthorButNoDisplayName: StoryFn = () => (
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

HistoryItemWithAuthorButNoDisplayName.storyName = 'History item with author without display name'

export const HistoryItemWithAuthorWithAvatarURL: StoryFn = () => (
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

HistoryItemWithAuthorWithAvatarURL.storyName = 'History item with author and avatar URL'
