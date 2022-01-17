import { storiesOf } from '@storybook/react'
import * as React from 'react'
import sinon from 'sinon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Progress } from '@sourcegraph/shared/src/search/stream'

import { StreamingProgress } from './StreamingProgress'

const { add } = storiesOf('web/search/results/progress/StreamingProgress', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
        },
        chromatic: { viewports: [1200] },
    })
    .addDecorator(story => <div className="d-flex align-items-center">{story()}</div>)

const onSearchAgain = sinon.spy()

add('0 results, in progress', () => {
    const progress: Progress = {
        durationMs: 0,
        matchCount: 0,
        skipped: [],
    }

    return (
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="loading" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
    )
})

add('0 results, in progress, traced', () => {
    const progress: Progress = {
        durationMs: 0,
        matchCount: 0,
        skipped: [],
        trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
    }

    return (
        <BrandedStory>
            {() => (
                <StreamingProgress progress={progress} state="loading" onSearchAgain={onSearchAgain} showTrace={true} />
            )}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="loading" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="complete" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
    )
})

add('big numbers, done, traced', () => {
    const progress: Progress = {
        durationMs: 52500,
        matchCount: 1234567,
        repositoriesCount: 8901,
        skipped: [],
        trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
    }

    return (
        <BrandedStory>
            {() => (
                <StreamingProgress
                    progress={progress}
                    state="complete"
                    onSearchAgain={onSearchAgain}
                    showTrace={true}
                />
            )}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="complete" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="loading" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="complete" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
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
        <BrandedStory>
            {() => <StreamingProgress progress={progress} state="loading" onSearchAgain={onSearchAgain} />}
        </BrandedStory>
    )
})
