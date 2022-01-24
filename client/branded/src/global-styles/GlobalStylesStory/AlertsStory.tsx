import { action } from '@storybook/addon-actions'
import { Story } from '@storybook/react'
import classNames from 'classnames'
import { flow } from 'lodash'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { SEMANTIC_COLORS } from './constants'
import { preventDefault } from './utils'

export const AlertsStory: Story = () => (
    <>
        <h1>Alerts</h1>
        <p>
            Provide contextual feedback messages for typical user actions with the handful of available and flexible
            alert messages.
        </p>
        {SEMANTIC_COLORS.map(semantic => (
            <div key={semantic} className={classNames('alert', `alert-${semantic}`)}>
                <h4>A shiny {semantic} alert - check it out!</h4>
                It can also contain{' '}
                <a href="/" onClick={flow(preventDefault, action('alert link clicked'))}>
                    links like this
                </a>
                .
            </div>
        ))}
        <div className="alert alert-info d-flex align-items-center">
            <div className="flex-grow-1">
                <h4>A shiny info alert with a button - check it out!</h4>
                It can also contain text without links.
            </div>
            <Button onClick={flow(preventDefault, action('alert button clicked'))} variant="info">
                Call to action
            </Button>
        </div>
    </>
)
