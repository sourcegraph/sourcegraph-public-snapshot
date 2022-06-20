import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'

import { Text } from '@sourcegraph/wildcard'

import { Timeline } from './Timeline'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/Timeline', module).addDecorator(story => (
    <div className="container mt-3" style={{ maxWidth: 600 }}>
        {story()}
    </div>
))

add('Basic', () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                stages={[
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <AlertCircleIcon />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                    },
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                    },
                    {
                        icon: <AlertCircleIcon />,
                        className: 'bg-danger',
                        text: 'Fourth event description',
                        date: '2020-06-15T14:30:00+00:00',
                    },
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'Fifth event description',
                        date: '2020-06-15T15:35:00+00:00',
                    },
                ]}
            />
        )}
    </WebStory>
))

add('Details', () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                stages={[
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <AlertCircleIcon />,
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
                        icon: <CheckIcon />,
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
))

add('Details, without durations', () => (
    <WebStory>
        {() => (
            <Timeline
                now={() => parseISO('2020-08-01T16:21:00+00:00')}
                showDurations={false}
                stages={[
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <AlertCircleIcon />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                        details: <Text className="m-0 p-0">HELLO THERE</Text>,
                    },
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                        details: <Text className="m-0 p-0">HELLO THERE</Text>,
                    },
                ]}
            />
        )}
    </WebStory>
))
