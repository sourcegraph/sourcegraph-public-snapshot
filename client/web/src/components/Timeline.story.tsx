import { mdiCheck, mdiAlertCircle } from '@mdi/js'
import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { parseISO } from 'date-fns'

import { Icon, Text } from '@sourcegraph/wildcard'

import { Timeline } from './Timeline'
import { WebStory } from './WebStory'

const decorator: Decorator = story => (
    <div className="container mt-3" style={{ maxWidth: 600 }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/Timeline',
    decorators: [decorator],
}

export default config

export const Basic: StoryFn = () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                stages={[
                    {
                        icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <Icon aria-label="Failed" svgPath={mdiAlertCircle} />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                    },
                    {
                        icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                    },
                    {
                        icon: <Icon aria-label="Failed" svgPath={mdiAlertCircle} />,
                        className: 'bg-danger',
                        text: 'Fourth event description',
                        date: '2020-06-15T14:30:00+00:00',
                    },
                    {
                        icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                        className: 'bg-success',
                        text: 'Fifth event description',
                        date: '2020-06-15T15:35:00+00:00',
                    },
                ]}
            />
        )}
    </WebStory>
)

export const Details: StoryFn = () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                stages={[
                    {
                        icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <Icon aria-label="Failed" svgPath={mdiAlertCircle} />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                        details: (
                            <>
                                <Text>HELLO THERE!</Text>
                                <Text className="m-0">I opened automatically because I'm important.</Text>
                            </>
                        ),
                        expandedByDefault: true,
                    },
                    {
                        icon: <Icon svgPath={mdiCheck} inline={false} aria-label="Success" />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                        details: <Text className="m-0 p-0">HELLO THERE</Text>,
                        expandedByDefault: false,
                    },
                ]}
            />
        )}
    </WebStory>
)

export const DetailsWithoutDurations: StoryFn = () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                showDurations={false}
                stages={[
                    {
                        icon: <Icon svgPath={mdiCheck} inline={false} aria-label="Success" />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <Icon svgPath={mdiAlertCircle} inline={false} aria-label="Failed" />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                        details: <Text className="m-0 p-0">HELLO THERE</Text>,
                    },
                    {
                        icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                        details: <Text className="m-0 p-0">HELLO THERE</Text>,
                    },
                ]}
            />
        )}
    </WebStory>
)

DetailsWithoutDurations.storyName = 'Details, without durations'
