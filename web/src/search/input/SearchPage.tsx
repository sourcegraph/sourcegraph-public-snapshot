import * as H from 'history'
import * as React from 'react'
import { parseSearchURLQuery, PatternTypeProps, InteractiveSearchProps } from '..'
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
import { QueryBuilder } from './QueryBuilder'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { ISearchScope, SearchFilterChips } from './SearchFilterChips'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchModeToggle } from './interactive/SearchModeToggle'

interface Props
    extends SettingsCascadeProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        InteractiveSearchProps {
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
        const queryFromUrl = parseSearchURLQuery(props.location.search, props.interactiveSearchMode) || ''
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
        let logoUrl =
            `${window.context.assetsRoot}/img/sourcegraph` +
            (this.props.isLightTheme ? '-light' : '') +
            '-head-logo.svg'
        const { branding } = window.context
        if (branding) {
            if (this.props.isLightTheme) {
                if (branding.light && branding.light.logo) {
                    logoUrl = branding.light.logo
                }
            } else if (branding.dark && branding.dark.logo) {
                logoUrl = branding.dark.logo
            }
        }
        const hasScopes = this.getScopes().length > 0
        const quickLinks = this.getQuickLinks()
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <img className="search-page__logo" src={logoUrl} />
                <div className="search search-page__container">
                    <div className="d-flex flex-row">
                        {this.props.splitSearchModes && this.props.interactiveSearchMode ? (
                            <InteractiveModeInput
                                {...this.props}
                                navbarSearchState={this.state.userQueryState}
                                onNavbarQueryChange={this.onUserQueryChange}
                                toggleSearchMode={this.props.toggleSearchMode}
                            />
                        ) : (
                            <>
                                <Form className="search flex-grow-1" onSubmit={this.onSubmit}>
                                    <div className="search-page__input-container">
                                        {this.props.splitSearchModes && (
                                            <SearchModeToggle
                                                {...this.props}
                                                interactiveSearchMode={this.props.interactiveSearchMode}
                                            />
                                        )}
                                        <QueryInput
                                            {...this.props}
                                            value={this.state.userQueryState}
                                            onChange={this.onUserQueryChange}
                                            autoFocus="cursor-at-end"
                                            hasGlobalQueryBehavior={true}
                                            patternType={this.props.patternType}
                                            setPatternType={this.props.setPatternType}
                                            withSearchModeToggle={this.props.splitSearchModes}
                                        />
                                        <SearchButton />
                                    </div>
                                    {hasScopes ? (
                                        <>
                                            <div className="search-page__input-sub-container">
                                                <SearchFilterChips
                                                    location={this.props.location}
                                                    history={this.props.history}
                                                    query={this.state.userQueryState.query}
                                                    authenticatedUser={this.props.authenticatedUser}
                                                    settingsCascade={this.props.settingsCascade}
                                                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                                    patternType={this.props.patternType}
                                                />
                                            </div>
                                            <QuickLinks
                                                quickLinks={quickLinks}
                                                className="search-page__input-sub-container"
                                            />
                                            <QueryBuilder
                                                onFieldsQueryChange={this.onBuilderQueryChange}
                                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                                patternType={this.props.patternType}
                                            />
                                        </>
                                    ) : (
                                        <>
                                            <QueryBuilder
                                                onFieldsQueryChange={this.onBuilderQueryChange}
                                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                                patternType={this.props.patternType}
                                            />
                                            <QuickLinks
                                                quickLinks={quickLinks}
                                                className="search-page__input-sub-container"
                                            />
                                            <div className="search-page__input-sub-container">
                                                <SearchFilterChips
                                                    location={this.props.location}
                                                    history={this.props.history}
                                                    query={this.state.userQueryState.query}
                                                    authenticatedUser={this.props.authenticatedUser}
                                                    settingsCascade={this.props.settingsCascade}
                                                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                                    patternType={this.props.patternType}
                                                />
                                            </div>
                                        </>
                                    )}
                                    <Notices
                                        className="my-3"
                                        location="home"
                                        settingsCascade={this.props.settingsCascade}
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

    private onBuilderQueryChange = (builderQuery: string): void => {
        this.setState({ builderQuery })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.builderQuery, this.state.userQueryState.query].filter(s => !!s).join(' ')
        submitSearch(this.props.history, query, 'home', this.props.patternType, this.props.activation)
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(this.props.location.search, this.props.interactiveSearchMode)
        if (query) {
            return `${limitString(this.state.userQueryState.query, 25, true)}`
        }
        return undefined
    }

    private getScopes(): ISearchScope[] {
        return (
            (isSettingsValid<Settings>(this.props.settingsCascade) &&
                this.props.settingsCascade.final['search.scopes']) ||
            []
        )
    }

    private getQuickLinks(): QuickLink[] {
        return (
            (isSettingsValid<Settings>(this.props.settingsCascade) && this.props.settingsCascade.final.quicklinks) || []
        )
    }
}
