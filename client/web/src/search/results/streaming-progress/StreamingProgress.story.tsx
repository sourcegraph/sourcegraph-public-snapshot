import * as React from 'react'
import { storiesOf } from '@storybook/react'
import { StreamingProgress } from './StreamingProgress'
import { WebStory } from '../../../components/WebStory'
import { Progress } from '../../stream'

const { add } = storiesOf('web/search/results/streaming-progress/StreamingProgress', module)

add('0 results, in progress', () => {
    const progress: Progress = {
        done: false,
        durationMs: 0,
        matchCount: 0,
        skipped: [],
    }

    return <WebStory>{() => <StreamingProgress progress={progress} />}</WebStory>
})

add('1 result from 1 repository, in progress', () => {
    const progress: Progress = {
        done: false,
        durationMs: 0,
        matchCount: 1,
        repositoriesCount: 1,
        skipped: [],
    }

    return <WebStory>{() => <StreamingProgress progress={progress} />}</WebStory>
})

add('2 results from 2 repositories, complete', () => {
    const progress: Progress = {
        done: true,
        durationMs: 0,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [],
    }

    return <WebStory>{() => <StreamingProgress progress={progress} />}</WebStory>
})
