import type { Meta, StoryFn } from '@storybook/react'
import { spy } from 'sinon'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { StreamingProgress } from './StreamingProgress'

const config: Meta = {
    title: 'branded/search-ui/results/progress/StreamingProgress',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
        },
    },
}

export default config

const onSearchAgain = spy()

const render = () => (
    <>
        <H2>0 results, in progress</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 0,
                    matchCount: 0,
                    skipped: [],
                }}
                state="loading"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>0 results, in progress, traced</H2>
        <div className="d-flex align-items-center  my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 0,
                    matchCount: 0,
                    skipped: [],
                    trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
                }}
                state="loading"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
                showTrace={true}
            />
        </div>

        <H2>1 result from 1 repository, in progress</H2>
        <div className="d-flex align-items-center  my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 500,
                    matchCount: 1,
                    repositoriesCount: 1,
                    skipped: [],
                }}
                state="loading"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>Big numbers, done</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 52500,
                    matchCount: 1234567,
                    repositoriesCount: 8901,
                    skipped: [],
                }}
                state="complete"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>Big numbers, done, traced</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 52500,
                    matchCount: 1234567,
                    repositoriesCount: 8901,
                    skipped: [],
                    trace: 'https://sourcegraph.test:3443/-/debug/jaeger/trace/abcdefg',
                }}
                state="complete"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
                showTrace={true}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with info</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 1500,
                    matchCount: 2,
                    repositoriesCount: 2,
                    skipped: [
                        {
                            reason: 'repository-fork',
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
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, loading, skipped with info</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 1500,
                    matchCount: 2,
                    repositoriesCount: 2,
                    skipped: [
                        {
                            reason: 'repository-fork',
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
                telemetryRecorder={noOpTelemetryRecorder}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with warning</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 1500,
                    matchCount: 2,
                    repositoriesCount: 2,
                    skipped: [
                        {
                            reason: 'repository-fork',
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
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>

        <H2>2 results from 2 repositories, complete, skipped with warning, limit hit, traced</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 1500,
                    matchCount: 2,
                    repositoriesCount: 2,
                    skipped: [
                        {
                            reason: 'repository-fork',
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
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                showTrace={true}
            />
        </div>

        <H2>2 results from 2 repositories, loading, skipped with warning</H2>
        <div className="d-flex align-items-center my-2">
            <StreamingProgress
                query=""
                progress={{
                    durationMs: 1500,
                    matchCount: 2,
                    repositoriesCount: 2,
                    skipped: [
                        {
                            reason: 'repository-fork',
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
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                onSearchAgain={onSearchAgain}
            />
        </div>
    </>
)

export const StreamingProgressStory: StoryFn = () => <BrandedStory>{() => <>{render()}</>}</BrandedStory>

StreamingProgressStory.storyName = 'StreamingProgress'
