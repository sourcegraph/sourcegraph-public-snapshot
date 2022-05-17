import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Typography } from '@sourcegraph/wildcard'

import { StreamingProgress } from './StreamingProgress'

const { add } = storiesOf('search-ui/results/progress/StreamingProgress', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
    },
    chromatic: { viewports: [1200], disableSnapshot: false },
})

const onSearchAgain = sinon.spy()

add('StreamingProgress', () => (
    <BrandedStory>
        {() => (
            <>
                <Typography.H2>0 results, in progress</Typography.H2>
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

                <Typography.H2>0 results, in progress, traced</Typography.H2>
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

                <Typography.H2>1 result from 1 repository, in progress</Typography.H2>
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

                <Typography.H2>Big numbers, done</Typography.H2>
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

                <Typography.H2>Big numbers, done, traced</Typography.H2>
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

                <Typography.H2>2 results from 2 repositories, complete, skipped with info</Typography.H2>
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

                <Typography.H2>2 results from 2 repositories, loading, skipped with info</Typography.H2>
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

                <Typography.H2>2 results from 2 repositories, complete, skipped with warning</Typography.H2>
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

                <Typography.H2>2 results from 2 repositories, loading, skipped with warning</Typography.H2>
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
        )}
    </BrandedStory>
))
