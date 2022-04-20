import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { of } from 'rxjs'

import { WebStory } from '../components/WebStory'

import { queryRepoChangesetsStats as _queryRepoChangesetsStats } from './backend'
import { RepoBatchChangesButton } from './RepoBatchChangesButton'

const { add } = storiesOf('web/batches/repo', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

const queryRepoChangesetsStats: typeof _queryRepoChangesetsStats = () =>
    of({
        changesetsStats: {
            open: number('Open', 2),
            merged: number('Merged', 47),
        },
    })

add('RepoButton', () => (
    <WebStory>
        {() => <RepoBatchChangesButton repoName="Awesome Repo" queryRepoChangesetsStats={queryRepoChangesetsStats} />}
    </WebStory>
))
