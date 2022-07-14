import { DecoratorFn, Meta, Story } from '@storybook/react'
import { useDarkMode } from 'storybook-dark-mode'

import { SearchPatternType } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { usePrependStyles } from '@sourcegraph/storybook'

import { callJava } from '../bridge-mock/call-java-mock'
import { light } from '../bridge-mock/theme-snapshots/light'
import { dark } from '../bridge-mock/theme-snapshots/dark'

import { applyTheme } from '.'

import { App } from './App'

import globalStyles from '../index.scss'
import { useEffect, useRef } from 'react'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'jetbrains/App',
    decorators: [decorator],
}

window.callJava = callJava

export default config

export const JetBrainsPluginApp: Story = () => {
    const rootElementRef = useRef<HTMLDivElement>(null)
    const isDarkTheme = useDarkMode()

    usePrependStyles('branded-story-styles', globalStyles)

    useEffect(() => {
        if (rootElementRef.current == null) {
            return
        }
        applyTheme(isDarkTheme ? dark : light, rootElementRef.current)
    }, [rootElementRef, isDarkTheme])

    return (
        <div ref={rootElementRef}>
            <div className="d-flex justify-content-center">
                <div className="mx-6">
                    <App
                        isDarkTheme={isDarkTheme}
                        instanceURL="https://sourcegraph.com/"
                        isGlobbingEnabled={false}
                        accessToken=""
                        initialSearch={{
                            query: 'repo:^github\\.com/sourcegraph/sourcegraph$@03af036 file:main\\.ts',
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
