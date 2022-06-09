import { useState } from 'react'

import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { TabBar, TabsConfig, TabKey } from './TabBar'

const { add } = storiesOf('web/batches/batch-spec/TabBar', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const CREATE_TABS: TabsConfig[] = [{ key: 'configuration', isEnabled: true }]

add('creating a new batch change', () => (
    <WebStory>{props => <TabBar {...props} activeTabKey="configuration" tabsConfig={CREATE_TABS} />}</WebStory>
))

add('editing unexecuted batch spec', () => {
    const [activeTabKey, setActiveTabKey] = useState<TabKey>('spec')

    const tabsConfig: TabsConfig[] = [
        {
            key: 'configuration',
            isEnabled: true,
            handler: {
                type: 'button',
                onClick: () => setActiveTabKey('configuration'),
            },
        },
        {
            key: 'spec',
            isEnabled: true,
            handler: {
                type: 'button',
                onClick: () => setActiveTabKey('spec'),
            },
        },
    ]

    return <WebStory>{props => <TabBar {...props} tabsConfig={tabsConfig} activeTabKey={activeTabKey} />}</WebStory>
})

const EXECUTING_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: true, handler: { type: 'link' } },
]

add('executing a batch spec', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={EXECUTING_TABS}
                activeTabKey={select('Active tab', ['configuration', 'spec', 'execution'], 'execution')}
            />
        )}
    </WebStory>
))

const PREVIEWING_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: true, handler: { type: 'link' } },
    { key: 'preview', isEnabled: true, handler: { type: 'link' } },
]

add('previewing an execution result', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={PREVIEWING_TABS}
                activeTabKey={select('Active tab', ['configuration', 'spec', 'execution', 'preview'], 'preview')}
            />
        )}
    </WebStory>
))
