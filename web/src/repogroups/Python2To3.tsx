import React, { useEffect, useState } from 'react'
import * as H from 'history'
import { PageTitle } from '../components/PageTitle'
import { BrandLogo } from '../components/branding/BrandLogo'
import { Form } from 'reactstrap'
import { SearchModeToggle } from '../search/input/interactive/SearchModeToggle'
import { VersionContextDropdown } from '../nav/VersionContextDropdown'
import { LazyMonacoQueryInput } from '../search/input/LazyMonacoQueryInput'
import { QueryInput } from '../search/input/QueryInput'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR, KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { SearchButton } from '../search/input/SearchButton'
import { Link } from '../../../shared/src/components/Link'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { Settings } from 'http2'
import { ThemeProps } from '../../../shared/src/theme'
import { ThemePreferenceProps } from '../theme'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    InteractiveSearchProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
} from '../search'
import { EventLoggerProps, eventLogger } from '../tracking/eventLogger'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { VersionContextProps } from '../../../shared/src/search/util'
import { VersionContext } from '../schema/site.schema'
import { submitSearch } from '../search/helpers'
import * as GQL from '../../../shared/src/graphql/schema'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        InteractiveSearchProps,
        SmartSearchFieldProps,
        CopyQueryButtonProps,
        VersionContextProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined

    // For NavLinks
    authRequired?: boolean
    showCampaigns: boolean
}

export const Python2To3: React.FunctionComponent<Props> = (props: Props) => {
    useEffect(() => eventLogger.logViewEvent('Python2To3RepoGroup'))

    const fixedQuery = 'repogroup:python-2-to-3'

    /** The query cursor position and value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: '',
        cursorPosition: 0,
    })

    const onSubmit = React.useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            // False positive
            // eslint-disable-next-line no-unused-expressions
            event?.preventDefault()
            submitSearch({ ...props, query: `${fixedQuery} ${userQueryState.query}`, source: 'home' })
        },
        [props, userQueryState.query]
    )

    return (
        <div className="repogroup-page">
            <PageTitle title="Python 2 to 3 migration" />
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} />
            <div className="repogroup-page__container">
                <div className="d-flex flex-row flex-shrink-past-contents">
                    <>
                        <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                            <div className="search-page__input-container">
                                {props.splitSearchModes && (
                                    <SearchModeToggle {...props} interactiveSearchMode={props.interactiveSearchMode} />
                                )}
                                <VersionContextDropdown
                                    history={props.history}
                                    caseSensitive={props.caseSensitive}
                                    patternType={props.patternType}
                                    navbarSearchQuery={userQueryState.query}
                                    versionContext={props.versionContext}
                                    setVersionContext={props.setVersionContext}
                                    availableVersionContexts={props.availableVersionContexts}
                                />
                                {props.smartSearchField ? (
                                    <LazyMonacoQueryInput
                                        {...props}
                                        hasGlobalQueryBehavior={true}
                                        queryState={userQueryState}
                                        onChange={setUserQueryState}
                                        onSubmit={onSubmit}
                                        autoFocus={true}
                                    />
                                ) : (
                                    <QueryInput
                                        {...props}
                                        value={userQueryState}
                                        onChange={setUserQueryState}
                                        autoFocus="cursor-at-end"
                                        hasGlobalQueryBehavior={true}
                                        patternType={props.patternType}
                                        setPatternType={props.setPatternType}
                                        withSearchModeToggle={props.splitSearchModes}
                                        keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                                    />
                                )}
                                <SearchButton />
                            </div>
                            <div className="search-page__input-sub-container">
                                {!props.splitSearchModes && (
                                    <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                                        Query builder
                                    </Link>
                                )}
                            </div>
                        </Form>
                    </>
                </div>
            </div>
            <div className="repogroup-page__content">
                <div>test</div>
                <div>
                    <div className="repogroup-page__repo-card card">Repositories</div>
                </div>
            </div>
        </div>
    )
}
