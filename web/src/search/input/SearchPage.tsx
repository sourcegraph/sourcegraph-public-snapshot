import * as H from 'history'
import * as React from 'react'
import {
    parseSearchURLQuery,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
} from '..'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { Notices } from '../../global/Notices'
import { QuickLink, Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { eventLogger, EventLoggerProps } from '../../tracking/eventLogger'
import { ThemePreferenceProps } from '../../theme'
import { limitString } from '../../util'
import { submitSearch, QueryState } from '../helpers'
import { QuickLinks } from '../QuickLinks'
import { QueryInput } from './QueryInput'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { SearchScopes } from './SearchScopes'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { KeyboardShortcutsProps, KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchModeToggle } from './interactive/SearchModeToggle'
import { Link } from '../../../../shared/src/components/Link'
import { BrandLogo } from '../../components/branding/BrandLogo'

interface Props
    extends SettingsCascadeProps,
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
        SmartSearchFieldProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean

    // For NavLinks
    authRequired?: boolean
    showCampaigns: boolean
}

interface State {
    /** The query cursor position and value entered by the user in the query input */
    userQueryState: QueryState
    /** The query that results from combining all values in the query builder form. */
    builderQuery: string
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        const queryFromUrl = parseSearchURLQuery(props.location.search) || ''
        this.state = {
            userQueryState: {
                query: queryFromUrl,
                cursorPosition: queryFromUrl.length,
            },
            builderQuery: '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
    }

    public render(): JSX.Element | null {
        const quickLinks = this.getQuickLinks()
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <BrandLogo className="search-page__logo" isLightTheme={this.props.isLightTheme} />
                <div className="search search-page__container">
                    <div className="d-flex flex-row">
                        {this.props.splitSearchModes && this.props.interactiveSearchMode ? (
                            <InteractiveModeInput
                                {...this.props}
                                navbarSearchState={this.state.userQueryState}
                                onNavbarQueryChange={this.onUserQueryChange}
                                toggleSearchMode={this.props.toggleSearchMode}
                                lowProfile={false}
                            />
                        ) : (
                            <>
                                <Form className="search flex-grow-1" onSubmit={this.onFormSubmit}>
                                    <div className="search-page__input-container">
                                        {this.props.splitSearchModes && (
                                            <SearchModeToggle
                                                {...this.props}
                                                interactiveSearchMode={this.props.interactiveSearchMode}
                                            />
                                        )}

                                        {this.props.smartSearchField ? (
                                            <LazyMonacoQueryInput
                                                {...this.props}
                                                hasGlobalQueryBehavior={true}
                                                queryState={this.state.userQueryState}
                                                onChange={this.onUserQueryChange}
                                                onSubmit={this.onSubmit}
                                                autoFocus={true}
                                            />
                                        ) : (
                                            <QueryInput
                                                {...this.props}
                                                value={this.state.userQueryState}
                                                onChange={this.onUserQueryChange}
                                                autoFocus="cursor-at-end"
                                                hasGlobalQueryBehavior={true}
                                                patternType={this.props.patternType}
                                                setPatternType={this.props.setPatternType}
                                                withSearchModeToggle={this.props.splitSearchModes}
                                                keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                                            />
                                        )}
                                        <SearchButton />
                                    </div>
                                    <div className="search-page__input-sub-container">
                                        {!this.props.splitSearchModes && (
                                            <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                                                Query builder
                                            </Link>
                                        )}
                                        <SearchScopes
                                            history={this.props.history}
                                            query={this.state.userQueryState.query}
                                            authenticatedUser={this.props.authenticatedUser}
                                            settingsCascade={this.props.settingsCascade}
                                            patternType={this.props.patternType}
                                        />
                                    </div>
                                    <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                                    <Notices
                                        className="my-3"
                                        location="home"
                                        settingsCascade={this.props.settingsCascade}
                                        history={this.props.history}
                                    />
                                </Form>
                            </>
                        )}
                    </div>
                </div>
            </div>
        )
    }

    private onUserQueryChange = (userQueryState: QueryState): void => {
        this.setState({ userQueryState })
    }

    private onSubmit = (): void => {
        const query = [this.state.builderQuery, this.state.userQueryState.query].filter(s => !!s).join(' ')
        submitSearch({
            ...this.props,
            query,
            source: 'home',
        })
    }

    private onFormSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        this.onSubmit()
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(this.props.location.search)
        if (query) {
            return `${limitString(this.state.userQueryState.query, 25, true)}`
        }
        return undefined
    }

    private getQuickLinks(): QuickLink[] {
        return (
            (isSettingsValid<Settings>(this.props.settingsCascade) && this.props.settingsCascade.final.quicklinks) || []
        )
    }
}
