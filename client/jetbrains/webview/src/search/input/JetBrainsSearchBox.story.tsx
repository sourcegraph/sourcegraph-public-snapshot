import { useEffect, useRef } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { BrowserRouter } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'
import { EMPTY, NEVER } from 'rxjs'
import { useDarkMode } from 'storybook-dark-mode'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
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
            <BrowserRouter>
                <CompatRouter>
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
                                    globbing={false}
                                    isLightTheme={!isDarkTheme}
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                    platformContext={{ requestGraphQL: () => EMPTY }}
                                    className=""
                                    containerClassName=""
                                    autoFocus={true}
                                    hideHelpButton={true}
                                />
                            </div>
                        </div>
                    </div>
                </CompatRouter>
            </BrowserRouter>
        </WildcardThemeContext.Provider>
    )
}

JetBrainsSearchBoxStory.parameters = {
    chromatic: {
        disableSnapshot: false,
    },
}
