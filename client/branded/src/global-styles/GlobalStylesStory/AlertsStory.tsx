import { action } from '@storybook/addon-actions'
import { Story } from '@storybook/react'
import { flow } from 'lodash'
import React from 'react'

import { Button, Alert } from '@sourcegraph/wildcard'

import { SEMANTIC_COLORS } from './constants'
import { preventDefault } from './utils'

// TODO: Remove story?

export const AlertsStory: Story = () => (
    <>
        <h1>Alerts</h1>
        <p>
            Provide contextual feedback messages for typical user actions with the handful of available and flexible
            alert messages.
        </p>
        {SEMANTIC_COLORS.map(semantic => (
            <Alert key={semantic} variant={semantic}>
                <h4>A shiny {semantic} alert - check it out!</h4>
                It can also contain{' '}
                <a href="/" onClick={flow(preventDefault, action('alert link clicked'))}>
                    links like this
                </a>
                .
            </Alert>
        ))}
        <Alert className="d-flex align-items-center" variant="info">
            <div className="flex-grow-1">
                <h4>A shiny info alert with a button - check it out!</h4>
                It can also contain text without links.
            </div>
            <Button onClick={flow(preventDefault, action('alert button clicked'))} variant="info">
                Call to action
            </Button>
        </Alert>
    </>
)
