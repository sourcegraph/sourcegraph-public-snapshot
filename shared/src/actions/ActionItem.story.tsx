import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import * as H from 'history'
import React from 'react'
import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { ActionItem, ActionItemComponentProps } from './ActionItem'
import './ActionItem.scss'

const EXTENSIONS_CONTROLLER: ActionItemComponentProps['extensionsController'] = {
    executeCommand: () => new Promise(resolve => setTimeout(resolve, 750)),
}

const PLATFORM_CONTEXT: ActionItemComponentProps['platformContext'] = {
    forceUpdateTooltip: () => undefined,
}

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const ICON_URL =
    'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=='

const onDidExecute = action('onDidExecute')

const { add } = storiesOf('ActionItem', module)

add('noop action', () => (
    <ActionItem
        action={{ id: 'a', command: undefined, actionItem: { label: 'Hello' } }}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        variant="actionItem"
        location={LOCATION}
        extensionsController={EXTENSIONS_CONTROLLER}
        platformContext={PLATFORM_CONTEXT}
    />
))

add('command action', () => (
    <ActionItem
        action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        disabledDuringExecution={true}
        showLoadingSpinnerDuringExecution={true}
        showInlineError={true}
        onDidExecute={onDidExecute}
        location={LOCATION}
        extensionsController={EXTENSIONS_CONTROLLER}
        platformContext={PLATFORM_CONTEXT}
    />
))

add('link action', () => (
    <ActionItem
        action={{
            id: 'a',
            command: 'open',
            commandArguments: ['javascript:alert("link clicked")'],
            actionItem: { label: 'Hello' },
        }}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        variant="actionItem"
        onDidExecute={onDidExecute}
        location={LOCATION}
        extensionsController={EXTENSIONS_CONTROLLER}
        platformContext={PLATFORM_CONTEXT}
    />
))

add('executing', () => {
    class ActionItemExecuting extends ActionItem {
        constructor(props: ActionItem['props']) {
            super(props)
            this.state.actionOrError = 'loading'
        }
    }
    return (
        <ActionItemExecuting
            action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            disabledDuringExecution={true}
            showLoadingSpinnerDuringExecution={true}
            showInlineError={true}
            location={LOCATION}
            extensionsController={EXTENSIONS_CONTROLLER}
            platformContext={PLATFORM_CONTEXT}
        />
    )
})

add('error', () => {
    class ActionItemWithError extends ActionItem {
        constructor(props: ActionItem['props']) {
            super(props)
            this.state.actionOrError = new Error('e')
        }
    }
    return (
        <ActionItemWithError
            action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            disabledDuringExecution={true}
            showLoadingSpinnerDuringExecution={true}
            showInlineError={true}
            location={LOCATION}
            extensionsController={EXTENSIONS_CONTROLLER}
            platformContext={PLATFORM_CONTEXT}
        />
    )
})
