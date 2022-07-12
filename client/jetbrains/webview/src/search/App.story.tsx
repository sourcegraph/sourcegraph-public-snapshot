import { DecoratorFn, Meta, Story } from '@storybook/react'
import { useDarkMode } from 'storybook-dark-mode'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { App } from './App'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'jetbrains/App',
    decorators: [decorator],
}

export default config

export const JetBrainsPluginApp: Story = () => (
    <div>
        <div>
            <div className="d-flex justify-content-center">
                <div className="mx-6">
                    <App
                        isDarkTheme={useDarkMode()}
                        instanceURL="https://sourcegraph.com"
                        isGlobbingEnabled={false}
                        accessToken=""
                        initialSearch={null}
                        onOpen={async () => {}}
                        onPreviewChange={async () => {}}
                        onPreviewClear={async () => {}}
                        authenticatedUser={null}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                </div>
            </div>
        </div>
    </div>
)

JetBrainsPluginApp.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
