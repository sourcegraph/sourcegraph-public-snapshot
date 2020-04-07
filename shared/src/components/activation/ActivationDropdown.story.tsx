import * as H from 'history'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { noop } from 'rxjs'
import '../../../../web/src/main.scss'
// import { action } from '@storybook/addon-actions'
// import { boolean } from '@storybook/addon-knobs'
import { ActivationDropdown } from './ActivationDropdown'
import { Activation } from './Activation'

const { add } = storiesOf('ActivationDropdown', module).addDecorator(story => (
    <div className="theme-light container">{story()}</div>
))

const baseActivation: Activation = {
    steps: [
        {
            id: 'ConnectedCodeHost',
            title: 'title1',
            detail: 'detail1',
            onClick: noop,
        },
        {
            id: 'EnabledRepository',
            title: 'title2',
            detail: 'detail2',
        },
    ],
    refetch: noop,
    update: noop,
    completed: undefined,
}
const history = H.createMemoryHistory({ keyLength: 0 })

add('Loading', () => <ActivationDropdown alwaysShow={true} history={history} activation={baseActivation} />)
add('0/2 completed', () => (
    <ActivationDropdown
        alwaysShow={true}
        history={history}
        activation={{
            ...baseActivation,
            completed: {
                ConnectedCodeHost: false,
                EnabledRepository: false,
            },
        }}
    />
))
add('1/2 completed', () => (
    <ActivationDropdown
        alwaysShow={true}
        history={history}
        activation={{ ...baseActivation, completed: { ConnectedCodeHost: true, EnabledRepository: false } }}
    />
))
