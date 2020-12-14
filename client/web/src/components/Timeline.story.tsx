import React from 'react'
import { storiesOf } from '@storybook/react'
import { WebStory } from './WebStory'
import { Timeline } from './Timeline'
import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'

const { add } = storiesOf('web/Timeline', module)

add('Empty', () => <WebStory>{() => <Timeline stages={[]} />}</WebStory>)

add('Basic', () => (
    <WebStory>
        {() => (
            <Timeline
                stages={[
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <ErrorIcon />,
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
                        icon: <ErrorIcon />,
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
                stages={[
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'First event description',
                        date: '2020-06-15T11:15:00+00:00',
                    },
                    {
                        icon: <ErrorIcon />,
                        className: 'bg-danger',
                        text: 'Second event description',
                        date: '2020-06-15T12:20:00+00:00',
                        details: <p>HELLO THERE</p>,
                        expanded: true,
                    },
                    {
                        icon: <CheckIcon />,
                        className: 'bg-success',
                        text: 'Third event description',
                        date: '2020-06-15T13:25:00+00:00',
                        details: <p>HELLO THERE</p>,
                        expanded: false,
                    },
                ]}
            />
        )}
    </WebStory>
))
