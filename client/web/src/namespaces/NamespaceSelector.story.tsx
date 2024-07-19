import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../components/WebStory'
import type { ViewerAffiliatedNamespacesResult } from '../graphql-operations'

import { NamespaceSelector } from './NamespaceSelector'

const decorator: Decorator = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/namespaces/NamespaceSelector',
    decorators: [decorator],
}

export default config

type Namespace = ViewerAffiliatedNamespacesResult['viewer']['affiliatedNamespaces']['nodes'][number]

const sampleNamespaces: Namespace[] = [
    { __typename: 'User', id: 'user1', namespaceName: 'alice' },
    { __typename: 'Org', id: 'org1', namespaceName: 'abc' },
    { __typename: 'Org', id: 'org2', namespaceName: 'xyz' },
]

export const Default: StoryFn = () => (
    <WebStory>
        {() => (
            <NamespaceSelector
                namespaces={sampleNamespaces}
                onSelect={namespace => console.log('Selected namespace:', namespace)}
            />
        )}
    </WebStory>
)

export const Loading: StoryFn = () => (
    <WebStory>
        {() => (
            <NamespaceSelector
                namespaces={[]}
                loading={true}
                onSelect={namespace => console.log('Selected namespace:', namespace)}
            />
        )}
    </WebStory>
)
