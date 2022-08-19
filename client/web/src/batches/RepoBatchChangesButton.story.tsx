import { Meta, Story, DecoratorFn } from '@storybook/react'
import { of } from 'rxjs'

import { WebStory } from '../components/WebStory'

import { queryRepoChangesetsStats as _queryRepoChangesetsStats } from './backend'
import { RepoBatchChangesButton } from './RepoBatchChangesButton'

const decorator: DecoratorFn = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/repo',
    decorators: [decorator],
}

export default config
let openValue = 0
let mergedValue = 0
const queryRepoChangesetsStats: typeof _queryRepoChangesetsStats = () =>
    of({
        changesetsStats: {
            open: openValue,
            merged: mergedValue,
        },
    })

export const RepoButton: Story = args => (
    <WebStory>
        {() => {
            openValue = args.open
            mergedValue = args.merged

            return (
                <RepoBatchChangesButton repoName="Awesome Repo" queryRepoChangesetsStats={queryRepoChangesetsStats} />
            )
        }}
    </WebStory>
)
RepoButton.argTypes = {
    open: {
        control: { type: 'number' },
        defaultValue: 2,
    },
    merged: {
        control: { type: 'number' },
        defaultValue: 47,
    },
}

RepoButton.storyName = 'RepoButton'
