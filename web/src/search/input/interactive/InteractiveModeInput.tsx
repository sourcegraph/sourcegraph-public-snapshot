import * as React from 'react'
import * as H from 'history'
import { QueryState, submitSearch } from '../../helpers'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { AddFilterRow } from './AddFilterRow'
import { SelectedFiltersRow } from './SelectedFiltersRow'
import { SearchButton } from '../SearchButton'
import { Subscription } from 'rxjs'
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
import { parseSearchURLQuery, InteractiveSearchProps } from '../..'
import { SearchModeToggle } from './SearchModeToggle'

interface InteractiveModeProps
    extends SettingsCascadeProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        ThemeProps,
        ThemePreferenceProps,
        EventLoggerProps,
        ActivationProps,
        Pick<InteractiveSearchProps, 'filtersInQuery' | 'onFiltersInQueryChange' | 'toggleSearchMode'> {
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
}

export default class InteractiveModeInput extends React.Component<InteractiveModeProps> {
    private numFiltersAddedToQuery = 0
    private subscriptions = new Subscription()

    constructor(props: InteractiveModeProps) {
        super(props)

        const searchParams = new URLSearchParams(props.location.search)
        const filtersInQuery: FiltersToTypeAndValue = {}
        for (const t of SuggestionTypeKeys) {
            const itemsOfType = searchParams.getAll(t)
            itemsOfType.map((item, i) => {
                filtersInQuery[`${t}-${i}`] = { type: t, value: item, editable: false }
            })
        }
        this.numFiltersAddedToQuery = Object.keys(filtersInQuery).length
        this.props.onFiltersInQueryChange(filtersInQuery)
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
        this.props.onFiltersInQueryChange({
            ...this.props.filtersInQuery,
            [filterKey]: { type: filterType, value: '', editable: true },
        })
    }

    private onFilterEdited = (filterKey: string, value: string): void => {
        this.props.onFiltersInQueryChange({
            ...this.props.filtersInQuery,
            [filterKey]: {
                ...this.props.filtersInQuery[filterKey],
                value,
            },
        })
    }

    private onFilterDeleted = (filterKey: string): void => {
        const newFiltersInQuery = this.props.filtersInQuery
        delete newFiltersInQuery[filterKey]

        submitSearch(
            this.props.history,
            this.props.navbarSearchState.query,
            'nav',
            this.props.patternType,
            undefined,
            newFiltersInQuery
        )

        this.props.onFiltersInQueryChange(newFiltersInQuery)
    }

    private toggleFilterEditable = (filterKey: string): void => {
        this.props.onFiltersInQueryChange({
            ...this.props.filtersInQuery,
            [filterKey]: {
                ...this.props.filtersInQuery[filterKey],
                editable: !this.props.filtersInQuery[filterKey].editable,
            },
        })
    }

    private onSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()

        submitSearch(
            this.props.history,
            this.props.navbarSearchState.query,
            'nav',
            this.props.patternType,
            undefined,
            this.props.filtersInQuery
        )
    }

    public render(): JSX.Element | null {
        const isSearchHomepage =
            this.props.location.pathname === '/search' && !parseSearchURLQuery(this.props.location.search, true)

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
                <div className={!isSearchHomepage ? 'interactive-mode-input__top-nav' : ''}>
                    {!isSearchHomepage &&
                        (this.props.authRequired ? (
                            <div className={logoLinkClassName}>{logo}</div>
                        ) : (
                            <Link to="/search" className={logoLinkClassName}>
                                {logo}
                            </Link>
                        ))}
                    <div
                        className={`d-none d-sm-flex flex-row ${
                            !isSearchHomepage ? 'interactive-mode-input__search-box-container' : ''
                        }`}
                    >
                        <Form onSubmit={this.onSubmit} className="flex-grow-1">
                            <div className="d-flex align-items-start">
                                <SearchModeToggle {...this.props} interactiveSearchMode={true} />
                                <QueryInput
                                    location={this.props.location}
                                    history={this.props.history}
                                    value={this.props.navbarSearchState}
                                    hasGlobalQueryBehavior={true}
                                    onChange={this.props.onNavbarQueryChange}
                                    patternType={this.props.patternType}
                                    togglePatternType={this.props.togglePatternType}
                                    autoFocus={true}
                                    filterQuery={this.props.filtersInQuery}
                                    withoutSuggestions={true}
                                    withSearchModeToggle={true}
                                />
                                <SearchButton noHelp={true} />
                            </div>
                        </Form>
                    </div>
                    {!this.props.authRequired && !isSearchHomepage && (
                        <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />
                    )}
                </div>
                <div>
                    <SelectedFiltersRow
                        filtersInQuery={this.props.filtersInQuery}
                        navbarQuery={this.props.navbarSearchState}
                        onSubmit={this.onSubmit}
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
