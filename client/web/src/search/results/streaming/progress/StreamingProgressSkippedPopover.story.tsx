import * as React from 'react'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../../../components/WebStory'
import { Progress } from '../../../stream'
import { StreamingProgressSkippedPopover } from './StreamingProgressSkippedPopover'

const { add } = storiesOf(
    'web/search/results/streaming/progress/StreamingProgressSkippedPopover',
    module
).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/IyiXZIbPHK447NCXov0AvK/13928-Streaming-search?node-id=280%3A17768',
    },
    chromatic: { viewports: [350] },
})

add('popover', () => {
    const progress: Progress = {
        done: true,
        durationMs: 1500,
        matchCount: 2,
        repositoriesCount: 2,
        skipped: [
            {
                reason: 'excluded-fork',
                message: '',
                severity: 'info',
                title: '10k forked repositories excluded',
                suggested: {
                    title: 'include forked',
                    queryExpression: 'fork:yes',
                },
            },
            {
                reason: 'excluded-archive',
                message: '',
                severity: 'info',
                title: '60k archived repositories excluded',
                suggested: {
                    title: 'include archived',
                    queryExpression: 'archived:yes',
                },
            },
            {
                reason: 'shard-timedout',
                message:
                    'Search timed out before some repositories could be searched. Try reducing scope of your query with repo: or other filters.',
                severity: 'warn',
                title: 'Search timed out',
                suggested: {
                    title: 'increase timeout',
                    queryExpression: 'timeout:2m',
                },
            },
        ],
    }

    return <WebStory>{() => <StreamingProgressSkippedPopover progress={progress} />}</WebStory>
})
