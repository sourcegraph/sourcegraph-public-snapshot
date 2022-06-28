import { number } from '@storybook/addon-knobs'
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

const queryRepoChangesetsStats: typeof _queryRepoChangesetsStats = () =>
    of({
        changesetsStats: {
            open: number('Open', 2),
            merged: number('Merged', 47),
        },
    })

export const RepoButton: Story = () => (
    <WebStory>
        {() => <RepoBatchChangesButton repoName="Awesome Repo" queryRepoChangesetsStats={queryRepoChangesetsStats} />}
    </WebStory>
)

RepoButton.storyName = 'RepoButton'
