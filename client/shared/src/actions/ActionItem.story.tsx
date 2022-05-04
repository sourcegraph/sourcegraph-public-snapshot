import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'
import { NEVER } from 'rxjs'

import { subtypeOf } from '@sourcegraph/common'
// eslint-disable-next-line no-restricted-imports
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'

import { ActionItem, ActionItemComponentProps, ActionItemProps } from './ActionItem'

const EXTENSIONS_CONTROLLER: ActionItemComponentProps['extensionsController'] = {
    executeCommand: () => new Promise(resolve => setTimeout(resolve, 750)),
}

const PLATFORM_CONTEXT: ActionItemComponentProps['platformContext'] = {
    forceUpdateTooltip: () => undefined,
    settings: NEVER,
}

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const ICON_URL =
    'data:image/svg+xml,' +
    encodeURIComponent(
        '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><text y=".9em" font-size="90">🚀</text></svg>'
    )

const onDidExecute = action('onDidExecute')

const commonProps = subtypeOf<Partial<ActionItemProps>>()({
    location: LOCATION,
    extensionsController: EXTENSIONS_CONTROLLER,
    platformContext: PLATFORM_CONTEXT,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    iconClassName: 'icon-inline',
    active: true,
})

const decorator: DecoratorFn = story => <WebStory>{() => <div className="p-4">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'shared/ActionItem',
    decorators: [decorator],
}
export default config

export const NoopAction: Story = () => (
    <ActionItem
        {...commonProps}
        action={{ id: 'a', command: undefined, actionItem: { label: 'Hello' } }}
        variant="actionItem"
    />
)

NoopAction.storyName = 'Noop action'

export const CommandAction: Story = () => (
    <ActionItem
        {...commonProps}
        action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        disabledDuringExecution={true}
        showLoadingSpinnerDuringExecution={true}
        showInlineError={true}
        onDidExecute={onDidExecute}
    />
)

CommandAction.storyName = 'Command action'
CommandAction.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const LinkAction: Story = () => (
    <ActionItem
        {...commonProps}
        action={{
            id: 'a',
            command: 'open',
            commandArguments: ['javascript:alert("link clicked")'],
            actionItem: { label: 'Hello' },
        }}
        variant="actionItem"
        onDidExecute={onDidExecute}
    />
)

LinkAction.storyName = 'Link action'

export const Executing: Story = () => {
    class ActionItemExecuting extends ActionItem {
        constructor(props: ActionItem['props']) {
            super(props)
            // eslint-disable-next-line react/no-this-in-sfc
            this.state.actionOrError = 'loading'
        }
    }
    return (
        <ActionItemExecuting
            {...commonProps}
            action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
            disabledDuringExecution={true}
            showLoadingSpinnerDuringExecution={true}
            showInlineError={true}
        />
    )
}

export const _Error: Story = () => {
    class ActionItemWithError extends ActionItem {
        constructor(props: ActionItem['props']) {
            super(props)
            // eslint-disable-next-line react/no-this-in-sfc
            this.state.actionOrError = new Error('e')
        }
    }
    return (
        <ActionItemWithError
            {...commonProps}
            action={{ id: 'a', command: 'c', title: 'Hello', iconURL: ICON_URL }}
            disabledDuringExecution={true}
            showLoadingSpinnerDuringExecution={true}
            showInlineError={true}
        />
    )
}
