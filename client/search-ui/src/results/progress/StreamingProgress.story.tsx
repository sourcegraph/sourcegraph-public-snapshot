import { Meta, Story } from '@storybook/react'
import { spy } from 'sinon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { H2 } from '@sourcegraph/wildcard'

import { StreamingProgress } from './StreamingProgress'

const config: Meta = {
    title: 'search-ui/results/progress/StreamingProgress',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
        },
        chromatic: { viewports: [1200], disableSnapshot: false },
    },
}

export default config

const onSearchAgain = spy()

const render = () => (
    <>
        <H2>0 results, in progress</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
                    durationMs: 0,
                    matchCount: 0,
                    skipped: [],
                }}
                state="loading"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>0 results, in progress, traced</H2>
        <div className="d-flex align-items-center  my-2">
            <StreamingProgress
                progress={{
                    durationMs: 0,
                    matchCount: 0,
                    skipped: [],
                    trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
                }}
                state="loading"
                onSearchAgain={onSearchAgain}
                showTrace={true}
            />
        </div>

        <H2>1 result from 1 repository, in progress</H2>
        <div className="d-flex align-items-center  my-2">
            <StreamingProgress
                progress={{
                    durationMs: 500,
                    matchCount: 1,
                    repositoriesCount: 1,
                    skipped: [],
                }}
                state="loading"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>Big numbers, done</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
                    durationMs: 52500,
                    matchCount: 1234567,
                    repositoriesCount: 8901,
                    skipped: [],
                }}
                state="complete"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>Big numbers, done, traced</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
                    durationMs: 52500,
                    matchCount: 1234567,
                    repositoriesCount: 8901,
                    skipped: [],
                    trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
                }}
                state="complete"
                onSearchAgain={onSearchAgain}
                showTrace={true}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with info</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
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
                }}
                state="complete"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, loading, skipped with info</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
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
                }}
                state="loading"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with warning</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
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
                }}
                state="complete"
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with warning, limit hit, traced</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
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
                            reason: 'shard-match-limit',
                            message: 'Search limit hit',
                            severity: 'warn',
                            title: 'Search limit hit',
                            suggested: {
                                title: 'count:all',
                                queryExpression: 'count:all',
                            },
                        },
                    ],
                    trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
                }}
                state="complete"
                onSearchAgain={onSearchAgain}
                showTrace={true}
            />
        </div>

        <H2>2 results from 2 repositories, loading, skipped with warning</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                progress={{
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
                }}
                state="loading"
                onSearchAgain={onSearchAgain}
            />
        </div>
    </>
)

export const StreamingProgressStory: Story = () => <BrandedStory>{() => <>{render()}</>}</BrandedStory>

StreamingProgressStory.storyName = 'StreamingProgress'
