import { useState } from 'react'

import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { TabBar, TabsConfig, TabName } from './TabBar'

const { add } = storiesOf('web/batches/batch-spec/TabBar', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const CREATE_TABS: TabsConfig[] = [{ name: 'configuration', isEnabled: true }]

add('creating a new batch change', () => (
    <WebStory>{props => <TabBar {...props} activeTabName="configuration" tabsConfig={CREATE_TABS} />}</WebStory>
))

add('editing unexecuted batch spec', () => {
    const [activeTabName, setActiveTabName] = useState<TabName>('batch spec')

    const tabsConfig: TabsConfig[] = [
        {
            name: 'configuration',
            isEnabled: true,
            handler: {
                type: 'button',
                onClick: () => setActiveTabName('configuration'),
            },
        },
        {
            name: 'batch spec',
            isEnabled: true,
            handler: {
                type: 'button',
                onClick: () => setActiveTabName('batch spec'),
            },
        },
    ]

    return <WebStory>{props => <TabBar {...props} tabsConfig={tabsConfig} activeTabName={activeTabName} />}</WebStory>
})

const EXECUTING_TABS: TabsConfig[] = [
    {
        name: 'configuration',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/configuration',
        },
    },
    {
        name: 'batch spec',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/spec',
        },
    },
    {
        name: 'execution',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/execution',
        },
    },
]

add('executing a batch spec', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={EXECUTING_TABS}
                activeTabName={select('Active tab', ['configuration', 'batch spec', 'execution'], 'execution')}
            />
        )}
    </WebStory>
))

const PREVIEWING_TABS: TabsConfig[] = [
    {
        name: 'configuration',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/configuration',
        },
    },
    {
        name: 'batch spec',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/spec',
        },
    },
    {
        name: 'execution',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/execution',
        },
    },
    {
        name: 'preview',
        isEnabled: true,
        handler: {
            type: 'link',
            to: '/preview',
        },
    },
]

add('previewing an execution result', () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={PREVIEWING_TABS}
                activeTabName={select('Active tab', ['configuration', 'batch spec', 'execution', 'preview'], 'preview')}
            />
        )}
    </WebStory>
))
