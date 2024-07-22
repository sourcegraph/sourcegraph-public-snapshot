import React from 'react'

import { action } from '@storybook/addon-actions'
import type { StoryFn, Meta } from '@storybook/react'
import classNames from 'classnames'
import { flow } from 'lodash'

import '@storybook/addon-designs'

import { H1, H4, Text } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { AlertLink } from '.'
import { Alert } from './Alert'
import { ALERT_VARIANTS } from './constants'

const preventDefault = <E extends React.SyntheticEvent>(event: E): E => {
    event.preventDefault()
    return event
}

const config: Meta = {
    title: 'wildcard/Alert',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
    parameters: {
        component: Alert,

        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A196',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A525',
            },
        ],
    },
}

export default config

export const Alerts: StoryFn = () => (
    <>
        <H1>Alerts</H1>
        <Text>
            Provide contextual feedback messages for typical user actions with the handful of available and flexible
            alert messages.
        </Text>
        <div className="mb-2">
            {ALERT_VARIANTS.map(variant => (
                <Alert key={variant} variant={variant}>
                    <H4>Too many matching repositories</H4>
                    Use a 'repo:' filter to narrow your search.
                </Alert>
            ))}
            <Alert variant="info" className="d-flex align-items-center">
                <div className="flex-grow-1">
                    <H4>Too many matching repositories</H4>
                    Use a 'repo:' filter to narrow your search.
                </div>
                <AlertLink className="mr-2" to="/" onClick={flow(preventDefault, action(classNames('link clicked')))}>
                    Dismiss
                </AlertLink>
            </Alert>

            <Alert variant="secondary" withIcon={false} className="d-flex align-items-center">
                <div className="flex-grow-1">
                    <H4>Too many matching repositories</H4>
                    Use a 'repo:' filter to narrow your search.
                </div>
                <AlertLink className="mr-2" to="/" onClick={flow(preventDefault, action(classNames('link clicked')))}>
                    Dismiss
                </AlertLink>
            </Alert>
        </div>
    </>
)
