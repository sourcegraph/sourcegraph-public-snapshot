import React from 'react'

import { action } from '@storybook/addon-actions'
import { Story, Meta } from '@storybook/react'
import classNames from 'classnames'
import { flow } from 'lodash'
import { animated } from 'react-spring'

import 'storybook-addon-designs'

import { mdiClose } from '@mdi/js'

import { Button, Code, H1, H2, H4, Icon, Text } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { AlertLink, useAnimatedAlert } from '.'
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
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
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

export const Alerts: Story = () => {
    const alertMessage = 'Success! This alert will now be shown for a couple seconds!'
    const manualAlertMessage = 'Warning! This alert will stay visible until it is dismissed.'
    const { isShown, show, ref, style } = useAnimatedAlert({
        autoDuration: 'short',
        ariaAnnouncement: { message: alertMessage },
    })
    const {
        isShown: manualIsShown,
        show: showManual,
        dismiss,
        ref: refManual,
        style: styleManual,
    } = useAnimatedAlert({ ariaAnnouncement: { message: manualAlertMessage } })

    return (
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
                    <AlertLink
                        className="mr-2"
                        to="/"
                        onClick={flow(preventDefault, action(classNames('link clicked')))}
                    >
                        Dismiss
                    </AlertLink>
                </Alert>
                <hr className="my-4" />
                <H2>Auto-dismissing alerts</H2>
                <div className="w-100 d-flex align-items-center justify-content-between mb-3">
                    <Text className="m-0">
                        Alerts can be automatically dismissed with an animation using the <Code>useAnimatedAlert</Code>{' '}
                        hook. Make sure the alert is still properly hidden and visible to screenreaders when it should
                        be by following the example below.
                    </Text>
                    <Button
                        variant="primary"
                        className="ml-2 flex-shrink-0"
                        onClick={() => {
                            show()
                            showManual()
                        }}
                        disabled={isShown}
                    >
                        Show alert
                    </Button>
                </div>
                <animated.div style={style}>
                    <Alert
                        ref={ref}
                        variant="success"
                        className="my-2"
                        // The alert announcement is handled by the useAnimatedAlert hook
                        aria-live="off"
                        aria-hidden={!isShown}
                    >
                        {alertMessage}
                    </Alert>
                </animated.div>
                <Text>I'm some text to demonstrate what happens when alerts appear around me.</Text>
                <animated.div style={styleManual}>
                    <Alert
                        ref={refManual}
                        variant="warning"
                        className="my-2 d-flex align-items-center justify-content-between"
                        // The alert announcement is handled by the useAnimatedAlert hook
                        aria-live="off"
                        aria-hidden={!manualIsShown}
                    >
                        {manualAlertMessage}
                        <Button variant="icon" onClick={dismiss}>
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </Alert>
                </animated.div>
                <Text>I'm some text to demonstrate what happens when an alert appears above me.</Text>
            </div>
        </>
    )
}
