import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import * as React from 'react'
import sinon from 'sinon'
import { WebStory } from '../../../../components/WebStory'
import { Progress } from '../../../stream'
import { StreamingProgress } from './StreamingProgress'

const history = createMemoryHistory()
const { add } = storiesOf('web/search/results/streaming/progress/StreamingProgress', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
    },
    chromatic: { viewports: [1200] },
})

const onSearchAgain = sinon.spy()

add('0 results, in progress', () => {
    const progress: Progress = {
        durationMs: 0,
        matchCount: 0,
        skipped: [],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="loading"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('1 result from 1 repository, in progress', () => {
    const progress: Progress = {
        durationMs: 500,
        matchCount: 1,
        repositoriesCount: 1,
        skipped: [],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="loading"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('big numbers, done', () => {
    const progress: Progress = {
        durationMs: 52500,
        matchCount: 1234567,
        repositoriesCount: 8901,
        skipped: [],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="complete"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('2 results from 2 repositories, complete, skipped with info', () => {
    const progress: Progress = {
        durationMs: 1500,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [
            {
                reason: 'excluded-fork',
                message: '10k forked repositories excluded',
                severity: 'info',
                title: '10k forked repositories excluded',
                suggested: {
                    title: 'forked:yes',
                    queryExpression: 'forked:yes',
                },
            },
            {
                reason: 'excluded-archive',
                message: '60k archived repositories excluded',
                severity: 'info',
                title: '60k archived repositories excluded',
                suggested: {
                    title: 'archived:yes',
                    queryExpression: 'archived:yes',
                },
            },
        ],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="complete"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('2 results from 2 repositories, loading, skipped with info', () => {
    const progress: Progress = {
        durationMs: 1500,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [
            {
                reason: 'excluded-fork',
                message: '10k forked repositories excluded',
                severity: 'info',
                title: '10k forked repositories excluded',
                suggested: {
                    title: 'forked:yes',
                    queryExpression: 'forked:yes',
                },
            },
            {
                reason: 'excluded-archive',
                message: '60k archived repositories excluded',
                severity: 'info',
                title: '60k archived repositories excluded',
                suggested: {
                    title: 'archived:yes',
                    queryExpression: 'archived:yes',
                },
            },
        ],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="loading"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('2 results from 2 repositories, complete, skipped with warning', () => {
    const progress: Progress = {
        durationMs: 1500,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [
            {
                reason: 'excluded-fork',
                message: '10k forked repositories excluded',
                severity: 'info',
                title: '10k forked repositories excluded',
                suggested: {
                    title: 'forked:yes',
                    queryExpression: 'forked:yes',
                },
            },
            {
                reason: 'excluded-archive',
                message: '60k archived repositories excluded',
                severity: 'info',
                title: '60k archived repositories excluded',
                suggested: {
                    title: 'archived:yes',
                    queryExpression: 'archived:yes',
                },
            },
            {
                reason: 'shard-timedout',
                message: 'Search timed out',
                severity: 'warn',
                title: 'Search timed out',
                suggested: {
                    title: 'timeout:2m',
                    queryExpression: 'timeout:2m',
                },
            },
        ],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="complete"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})

add('2 results from 2 repositories, loading, skipped with warning', () => {
    const progress: Progress = {
        durationMs: 1500,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [
            {
                reason: 'excluded-fork',
                message: '10k forked repositories excluded',
                severity: 'info',
                title: '10k forked repositories excluded',
                suggested: {
                    title: 'forked:yes',
                    queryExpression: 'forked:yes',
                },
            },
            {
                reason: 'excluded-archive',
                message: '60k archived repositories excluded',
                severity: 'info',
                title: '60k archived repositories excluded',
                suggested: {
                    title: 'archived:yes',
                    queryExpression: 'archived:yes',
                },
            },
            {
                reason: 'shard-timedout',
                message: 'Search timed out',
                severity: 'warn',
                title: 'Search timed out',
                suggested: {
                    title: 'timeout:2m',
                    queryExpression: 'timeout:2m',
                },
            },
        ],
    }

    return (
        <WebStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="loading"
                    onSearchAgain={onSearchAgain}
                    history={history}
                />
            )}
        </WebStory>
    )
})
