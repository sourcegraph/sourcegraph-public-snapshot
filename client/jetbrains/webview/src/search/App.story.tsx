import { DecoratorFn, Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H1, H2 } from '@sourcegraph/wildcard'

import { App } from './App'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'extensions/ide/jetbrains/App',
    decorators: [decorator],
}

export default config

export const JetBrainsPluginApp: Story = () => (
    <div>
        <H1 className="text-center mb-3">JetBrains plugin: light and dark</H1>
        <div>
            <div className="d-flex justify-content-center">
                <div className="mx-6">
                    <H2 className="text-center">Light</H2>
                    <App
                        isDarkTheme={false}
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
                <div className="mx-6">
                    <H2 className="text-center">Dark</H2>
                    <App
                        isDarkTheme={true}
                        instanceURL="https://k8s.sgdev.org"
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
