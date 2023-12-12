import { useEffect, useRef } from 'react'

import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { BrowserRouter } from 'react-router-dom'
import { EMPTY, NEVER } from 'rxjs'
import { useDarkMode } from 'storybook-dark-mode'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeContext, ThemeSetting } from '@sourcegraph/shared/src/theme'
import { WildcardThemeContext } from '@sourcegraph/wildcard'
import { usePrependStyles } from '@sourcegraph/wildcard/src/stories'

import { applyTheme } from '..'
import { dark } from '../../bridge-mock/theme-snapshots/dark'
import { light } from '../../bridge-mock/theme-snapshots/light'
import { SearchPatternType } from '../../graphql-operations'

import { JetBrainsSearchBox } from './JetBrainsSearchBox'

import globalStyles from '../../index.scss'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'jetbrains/JetBrainsSearchBox',
    decorators: [decorator],
}

export default config

export const JetBrainsSearchBoxStory: Story = () => {
    const rootElementRef = useRef<HTMLDivElement>(null)
    const isDarkTheme = useDarkMode()

    usePrependStyles('branded-story-styles', globalStyles)

    useEffect(() => {
        if (rootElementRef.current === null) {
            return
        }
        applyTheme(isDarkTheme ? dark : light, rootElementRef.current)
    }, [rootElementRef, isDarkTheme])

    return (
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <ThemeContext.Provider value={{ themeSetting: !isDarkTheme ? ThemeSetting.Light : ThemeSetting.Dark }}>
                <BrowserRouter>
                    <div ref={rootElementRef}>
                        <div className="d-flex justify-content-center">
                            <div className="mx-6">
                                <JetBrainsSearchBox
                                    caseSensitive={true}
                                    setCaseSensitivity={() => {}}
                                    patternType={SearchPatternType.regexp}
                                    setPatternType={() => {}}
                                    isSourcegraphDotCom={false}
                                    structuralSearchDisabled={false}
                                    queryState={{ query: 'type:file test AND test repo:contains.file(CHANGELOG)' }}
                                    onChange={() => {}}
                                    onSubmit={() => {}}
                                    authenticatedUser={null}
                                    searchContextsEnabled={true}
                                    showSearchContext={true}
                                    showSearchContextManagement={false}
                                    setSelectedSearchContextSpec={() => {}}
                                    selectedSearchContextSpec={undefined}
                                    fetchSearchContexts={() => {
                                        throw new Error('fetchSearchContexts')
                                    }}
                                    getUserSearchContextNamespaces={() => []}
                                    fetchStreamSuggestions={() => NEVER}
                                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                    telemetryRecorder={noOpTelemetryRecorder}
                                    platformContext={{ requestGraphQL: () => EMPTY }}
                                    className=""
                                    containerClassName=""
                                    autoFocus={true}
                                    hideHelpButton={true}
                                />
                            </div>
                        </div>
                    </div>
                </BrowserRouter>
            </ThemeContext.Provider>
        </WildcardThemeContext.Provider>
    )
}

JetBrainsSearchBoxStory.parameters = {
    chromatic: {
        disableSnapshot: false,
    },
}
