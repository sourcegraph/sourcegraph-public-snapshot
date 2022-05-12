import React from 'react'

import { action } from '@storybook/addon-actions'
import { Story, Meta } from '@storybook/react'
import classNames from 'classnames'
import { flow } from 'lodash'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import 'storybook-addon-designs'

import { Typography } from '..'

import { Alert } from './Alert'
import { ALERT_VARIANTS } from './constants'

import { AlertLink } from '.'

const preventDefault = <E extends React.SyntheticEvent>(event: E): E => {
    event.preventDefault()
    return event
}

const config: Meta = {
    title: 'wildcard/Alert',
    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        component: Alert,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url:
                    'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A196',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url:
                    'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A525',
            },
        ],
    },
}

export default config

export const Alerts: Story = () => (
    <>
        <Typography.H1>Alerts</Typography.H1>
        <p>
            Provide contextual feedback messages for typical user actions with the handful of available and flexible
            alert messages.
        </p>
        <div className="mb-2">
            {ALERT_VARIANTS.map(variant => (
                <Alert key={variant} variant={variant}>
                    <Typography.H4>Too many matching repositories</Typography.H4>
                    Use a 'repo:' or 'repogroup:' filter to narrow your search.
                </Alert>
            ))}
            <Alert variant="info" className="d-flex align-items-center">
                <div className="flex-grow-1">
                    <Typography.H4>Too many matching repositories</Typography.H4>
                    Use a 'repo:' or 'repogroup:' filter to narrow your search.
                </div>
                <AlertLink className="mr-2" to="/" onClick={flow(preventDefault, action(classNames('link clicked')))}>
                    Dismiss
                </AlertLink>
            </Alert>
        </div>
    </>
)
