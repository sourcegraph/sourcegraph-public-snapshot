import { useState } from 'react'

import { select } from '@storybook/addon-knobs'
import { DecoratorFn, Story, Meta } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { TabBar, TabsConfig, TabKey } from './TabBar'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/TabBar',
    decorators: [decorator],
}

export default config

const CREATE_TABS: TabsConfig[] = [{ key: 'configuration', isEnabled: true }]

export const CreateNewBatchChange: Story = () => (
    <WebStory>{props => <TabBar {...props} activeTabKey="configuration" tabsConfig={CREATE_TABS} />}</WebStory>
)

CreateNewBatchChange.storyName = 'creating a new batch change'

export const EditUnexecutedBatchSpec: Story = () => {
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
}

EditUnexecutedBatchSpec.storyName = 'editing unexecuted batch spec'

const EXECUTING_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: true, handler: { type: 'link' } },
]

export const ExecuteBatchSpec: Story = () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={EXECUTING_TABS}
                activeTabKey={select('Active tab', ['configuration', 'spec', 'execution'], 'execution')}
            />
        )}
    </WebStory>
)

ExecuteBatchSpec.storyName = 'executing a batch spec'

const PREVIEWING_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: true, handler: { type: 'link' } },
    { key: 'preview', isEnabled: true, handler: { type: 'link' } },
]

export const PreviewExecutionResult: Story = () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={PREVIEWING_TABS}
                activeTabKey={select('Active tab', ['configuration', 'spec', 'execution', 'preview'], 'preview')}
            />
        )}
    </WebStory>
)

PreviewExecutionResult.storyName = 'previewing an execution result'

const LOCAL_TABS: TabsConfig[] = [
    { key: 'configuration', isEnabled: true, handler: { type: 'link' } },
    { key: 'spec', isEnabled: true, handler: { type: 'link' } },
    { key: 'execution', isEnabled: false },
    { key: 'preview', isEnabled: true, handler: { type: 'link' } },
]

export const LocallyExecutedSpec: Story = () => (
    <WebStory>
        {props => (
            <TabBar
                {...props}
                tabsConfig={LOCAL_TABS}
                activeTabKey={select('Active tab', ['configuration', 'spec', 'preview'], 'preview')}
            />
        )}
    </WebStory>
)

LocallyExecutedSpec.storyName = 'for a locally-executed spec'
