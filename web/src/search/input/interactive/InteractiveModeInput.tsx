import * as React from 'react'
import * as H from 'history'
import { QueryState, submitSearch } from '../../helpers'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { AddFilterRow } from './AddFilterRow'
import { SelectedFiltersRow } from './SelectedFiltersRow'
import { SearchButton } from '../SearchButton'
import { Subscription, Subject } from 'rxjs'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Link } from '../../../../../shared/src/components/Link'
import { NavLinks } from '../../../nav/NavLinks'
import { showDotComMarketing } from '../../../util/features'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { KeyboardShortcutsProps } from '../../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ThemePreferenceProps } from '../../theme'
import { EventLoggerProps } from '../../../tracking/eventLogger'
import { ActivationProps } from '../../../../../shared/src/components/activation/Activation'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import { SuggestionTypes, SuggestionTypeKeys } from '../../../../../shared/src/search/suggestions/util'
import { QueryInput } from '../QueryInput'
import { parseSearchURLQuery } from '../..'

interface InteractiveModeProps
    extends SettingsCascadeProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        ThemeProps,
        ThemePreferenceProps,
        EventLoggerProps,
        ActivationProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onNavbarQueryChange: (userQuery: QueryState) => void
    patternType: GQL.SearchPatternType
    togglePatternType: () => void

    // For NavLinks
    authRequired?: boolean
    authenticatedUser: GQL.IUser | null
    showDotComMarketing: boolean
    showCampaigns: boolean
    isSourcegraphDotCom: boolean

    toggleSearchMode: (e: React.MouseEvent<HTMLAnchorElement>) => void
}

interface InteractiveInputState {
    /**
     * filtersInQuery is the source of truth for the filter values currently in the query.
     *
     * The data structure is a map, where the key is a uniquely assigned string in the form `repoType-numberOfFilterAdded`.
     * The value is a data structure containing the fields {`type`, `value`, `editable`}.
     * `type` is the field type of the filter (repo, file, etc.) `value` is the current value for that particular filter,
     * and `editable` is whether the corresponding filter input is currently editable in the UI.
     * */
    filtersInQuery: FiltersToTypeAndValue
}

export default class InteractiveModeInput extends React.Component<InteractiveModeProps, InteractiveInputState> {
    private numFiltersAddedToQuery = 0
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<InteractiveModeProps>()

    constructor(props: InteractiveModeProps) {
        super(props)

        this.state = {
            filtersInQuery: {},
        }
        this.subscriptions.add(
            this.componentUpdates.subscribe(props => {
                const searchParams = new URLSearchParams(props.location.search)
                const filtersInQuery: FiltersToTypeAndValue = {}
                for (const t of SuggestionTypeKeys) {
                    const itemsOfType = searchParams.getAll(t)
                    itemsOfType.map((item, i) => {
                        filtersInQuery[`${t}-${i}`] = { type: t, value: item, editable: false }
                    })
                }
                this.numFiltersAddedToQuery = Object.keys(filtersInQuery).length
                this.setState({ filtersInQuery })
            })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    /**
     * Adds a new filter to the filtersInQuery state field.
     * We use the filter name and the number of values added as the key.
     * Keys must begin with the filter name, as defined in `SuggestionTypes`.
     * We use this to identify filter values when building
     * the search URL in {@link interactiveBuildSearchURLQuery}.
     */
    private addNewFilter = (filterType: SuggestionTypes): void => {
        const filterKey = `${filterType}-${this.numFiltersAddedToQuery}`
        this.numFiltersAddedToQuery++
        this.setState(state => ({
            filtersInQuery: { ...state.filtersInQuery, [filterKey]: { type: filterType, value: '', editable: true } },
        }))
    }

    private onFilterEdited = (filterKey: string, value: string): void => {
        this.setState(state => ({
            filtersInQuery: {
                ...state.filtersInQuery,
                [filterKey]: {
                    ...state.filtersInQuery[filterKey],
                    value,
                },
            },
        }))
    }

    private onFilterDeleted = (filterKey: string): void => {
        this.setState(state => {
            const newState = state.filtersInQuery
            delete newState[filterKey]
            return { filtersInQuery: newState }
        })
    }

    private toggleFilterEditable = (filterKey: string): void => {
        this.setState(state => ({
            filtersInQuery: {
                ...state.filtersInQuery,
                [filterKey]: {
                    ...state.filtersInQuery[filterKey],
                    editable: !state.filtersInQuery[filterKey].editable,
                },
            },
        }))
    }

    private onSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()

        submitSearch(
            this.props.history,
            this.props.navbarSearchState.query,
            'nav',
            this.props.patternType,
            undefined,
            this.state.filtersInQuery
        )
    }

    public render(): JSX.Element | null {
        const isSearchHomepage =
            this.props.location.pathname === '/search' && !parseSearchURLQuery(this.props.location.search, true)
        console.log('isSearchHomepage', isSearchHomepage)

        let logoSrc = '/.assets/img/sourcegraph-mark.svg'
        let logoLinkClassName = 'global-navbar__logo-link global-navbar__logo-animated'

        const { branding } = window.context
        if (branding) {
            if (this.props.isLightTheme) {
                if (branding.light && branding.light.symbol) {
                    logoSrc = branding.light.symbol
                }
            } else if (branding.dark && branding.dark.symbol) {
                logoSrc = branding.dark.symbol
            }
            if (branding.disableSymbolSpin) {
                logoLinkClassName = 'global-navbar__logo-link'
            }
        }

        const logo = <img className="global-navbar__logo" src={logoSrc} />

        return (
            <div className="interactive-mode-input e2e-interactive-mode-input">
                <div className="interactive-mode-input__top-nav">
                    {!isSearchHomepage &&
                        (this.props.authRequired ? (
                            <div className={logoLinkClassName}>{logo}</div>
                        ) : (
                            <Link to="/search" className={logoLinkClassName}>
                                {logo}
                            </Link>
                        ))}
                    <div className="global-navbar__search-box-container d-none d-sm-flex">
                        <Form onSubmit={this.onSubmit}>
                            <div className="d-flex align-items-start">
                                <QueryInput
                                    location={this.props.location}
                                    history={this.props.history}
                                    value={this.props.navbarSearchState}
                                    hasGlobalQueryBehavior={true}
                                    onChange={this.props.onNavbarQueryChange}
                                    patternType={this.props.patternType}
                                    togglePatternType={this.props.togglePatternType}
                                    autoFocus={true}
                                    filterQuery={this.state.filtersInQuery}
                                    withoutSuggestions={true}
                                />
                                <SearchButton />
                            </div>
                        </Form>
                    </div>
                    {!this.props.authRequired && !isSearchHomepage && (
                        <NavLinks
                            {...this.props}
                            interactiveSearchMode={true}
                            showInteractiveSearchMode={true}
                            showDotComMarketing={showDotComMarketing}
                        />
                    )}
                </div>
                <div>
                    <SelectedFiltersRow
                        filtersInQuery={this.state.filtersInQuery}
                        navbarQuery={this.props.navbarSearchState}
                        onFilterEdited={this.onFilterEdited}
                        onFilterDeleted={this.onFilterDeleted}
                        toggleFilterEditable={this.toggleFilterEditable}
                        isHomepage={isSearchHomepage}
                    />
                    <AddFilterRow onAddNewFilter={this.addNewFilter} isHomepage={isSearchHomepage} />
                </div>
            </div>
        )
    }
}
