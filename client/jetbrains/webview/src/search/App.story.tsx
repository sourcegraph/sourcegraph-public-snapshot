import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'
import { useDarkMode } from 'storybook-dark-mode'

import { SearchPatternType } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { usePrependStyles } from '@sourcegraph/storybook'

import { callJava } from '../bridge-mock/call-java-mock'

import { App } from './App'

import globalStyles from '../index.scss'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'jetbrains/App',
    decorators: [decorator],
}

window.callJava = callJava

export default config

export const JetBrainsPluginApp: Story = () => {
    usePrependStyles('branded-story-styles', globalStyles)

    return (
        <div className={classNames('theme', useDarkMode() ? 'theme-dark' : 'theme-light')}>
            <div className="d-flex justify-content-center">
                <div className="mx-6">
                    <App
                        isDarkTheme={useDarkMode()}
                        instanceURL="https://sourcegraph.com/"
                        isGlobbingEnabled={false}
                        accessToken=""
                        initialSearch={{
                            query:
                                'repo:^github\\.com/sourcegraph/sourcegraph$@03af036 file:^client/storybook/src/main\\.ts',
                            caseSensitive: false,
                            patternType: SearchPatternType.standard,
                            selectedSearchContextSpec: 'global',
                        }}
                        onOpen={async () => {}}
                        onPreviewChange={async () => {}}
                        onPreviewClear={async () => {}}
                        authenticatedUser={null}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                    />
                </div>
            </div>
        </div>
    )
}

JetBrainsPluginApp.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
        delay: 10000,
    },
}
