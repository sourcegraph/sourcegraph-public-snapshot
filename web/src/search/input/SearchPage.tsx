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
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { SearchScopes } from './SearchScopes'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchModeToggle } from './interactive/SearchModeToggle'
import { Link } from '../../../../shared/src/components/Link'

interface Props
    extends SettingsCascadeProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
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
        that.state = {
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
            (that.props.isLightTheme ? '-light' : '') +
            '-head-logo.svg'
        const { branding } = window.context
        if (branding) {
            if (that.props.isLightTheme) {
                if (branding.light && branding.light.logo) {
                    logoUrl = branding.light.logo
                }
            } else if (branding.dark && branding.dark.logo) {
                logoUrl = branding.dark.logo
            }
        }
        const quickLinks = that.getQuickLinks()
        return (
            <div className="search-page">
                <PageTitle title={that.getPageTitle()} />
                <img className="search-page__logo" src={logoUrl} />
                <div className="search search-page__container">
                    <div className="d-flex flex-row">
                        {that.props.splitSearchModes && that.props.interactiveSearchMode ? (
                            <InteractiveModeInput
                                {...that.props}
                                navbarSearchState={that.state.userQueryState}
                                onNavbarQueryChange={that.onUserQueryChange}
                                toggleSearchMode={that.props.toggleSearchMode}
                            />
                        ) : (
                            <>
                                <Form className="search flex-grow-1" onSubmit={that.onSubmit}>
                                    <div className="search-page__input-container">
                                        {that.props.splitSearchModes && (
                                            <SearchModeToggle
                                                {...that.props}
                                                interactiveSearchMode={that.props.interactiveSearchMode}
                                            />
                                        )}
                                        <QueryInput
                                            {...that.props}
                                            value={that.state.userQueryState}
                                            onChange={that.onUserQueryChange}
                                            autoFocus="cursor-at-end"
                                            hasGlobalQueryBehavior={true}
                                            patternType={that.props.patternType}
                                            setPatternType={that.props.setPatternType}
                                            withSearchModeToggle={that.props.splitSearchModes}
                                        />
                                        <SearchButton />
                                    </div>
                                    <div className="search-page__input-sub-container">
                                        {!that.props.splitSearchModes && (
                                            <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                                                Query builder
                                            </Link>
                                        )}
                                        <SearchScopes
                                            history={that.props.history}
                                            query={that.state.userQueryState.query}
                                            authenticatedUser={that.props.authenticatedUser}
                                            settingsCascade={that.props.settingsCascade}
                                            patternType={that.props.patternType}
                                        />
                                    </div>
                                    <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                                    <Notices
                                        className="my-3"
                                        location="home"
                                        settingsCascade={that.props.settingsCascade}
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
        that.setState({ userQueryState })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [that.state.builderQuery, that.state.userQueryState.query].filter(s => !!s).join(' ')
        submitSearch(that.props.history, query, 'home', that.props.patternType, that.props.activation)
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(that.props.location.search, that.props.interactiveSearchMode)
        if (query) {
            return `${limitString(that.state.userQueryState.query, 25, true)}`
        }
        return undefined
    }

    private getQuickLinks(): QuickLink[] {
        return (
            (isSettingsValid<Settings>(that.props.settingsCascade) && that.props.settingsCascade.final.quicklinks) || []
        )
    }
}
