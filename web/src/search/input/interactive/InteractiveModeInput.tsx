import * as React from 'react'
import * as H from 'history'
import { QueryState, submitSearch } from '../../helpers'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { AddFilterRow } from './AddFilterRow'
import { SelectedFiltersRow } from './SelectedFiltersRow'
import { SearchButton } from '../SearchButton'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Link } from '../../../../../shared/src/components/Link'
import { NavLinks } from '../../../nav/NavLinks'
import { showDotComMarketing } from '../../../util/features'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { KeyboardShortcutsProps } from '../../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ThemePreferenceProps } from '../../../theme'
import { EventLoggerProps } from '../../../tracking/eventLogger'
import { ActivationProps } from '../../../../../shared/src/components/activation/Activation'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import { SuggestionTypes, suggestionTypeKeys } from '../../../../../shared/src/search/suggestions/util'
import { QueryInput } from '../QueryInput'
import { parseSearchURLQuery, InteractiveSearchProps, PatternTypeProps } from '../..'
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
        PatternTypeProps,
        Pick<InteractiveSearchProps, 'filtersInQuery' | 'onFiltersInQueryChange' | 'toggleSearchMode'> {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onNavbarQueryChange: (userQuery: QueryState) => void

    // For NavLinks
    authRequired?: boolean
    authenticatedUser: GQL.IUser | null
    showCampaigns: boolean
    isSourcegraphDotCom: boolean
}

interface InteractiveModeState {
    /** Count of the total number of filters ever added in this component.*/
    numFiltersAdded: number
}

export class InteractiveModeInput extends React.Component<InteractiveModeProps, InteractiveModeState> {
    constructor(props: InteractiveModeProps) {
        super(props)

        const searchParams = new URLSearchParams(props.location.search)
        const filtersInQuery: FiltersToTypeAndValue = {}
        for (const t of suggestionTypeKeys) {
            const itemsOfType = searchParams.getAll(t)
            itemsOfType.map((item, i) => {
                filtersInQuery[`${t}-${i}`] = { type: t, value: item, editable: false }
            })
        }

        this.state = {
            numFiltersAdded: Object.keys(filtersInQuery).length,
        }

        this.props.onFiltersInQueryChange(filtersInQuery)
    }

    /**
     * Adds a new filter to the top-level filtersInQuery state field.
     * We use the filter name and the number of values added as the key.
     * Keys must begin with the filter name, as defined in `SuggestionTypes`.
     * We use this to identify filter values when building
     * the search URL in {@link interactiveBuildSearchURLQuery}.
     */
    private addNewFilter = (filterType: SuggestionTypes): void => {
        const filterKey = `${filterType}-${this.state.numFiltersAdded}`
        this.setState(state => ({ numFiltersAdded: state.numFiltersAdded + 1 }))
        this.props.onFiltersInQueryChange({
            ...this.props.filtersInQuery,
            [filterKey]: { type: filterType, value: '', editable: true },
        })
    }

    /**
     * onFilterEdited updates the top-level filtersInQuery object with new values
     * when new filter values are submitted by the user.
     *
     * Also conducts a new search with the updated query.
     */
    private onFilterEdited = (filterKey: string, value: string): void => {
        const newFiltersInQuery = {
            ...this.props.filtersInQuery,
            [filterKey]: {
                ...this.props.filtersInQuery[filterKey],
                value,
                editable: false,
            },
        }
        // Update the top-level filtersInQuery with the new values
        this.props.onFiltersInQueryChange(newFiltersInQuery)

        // Submit a search with the new values
        submitSearch(
            this.props.history,
            this.props.navbarSearchState.query,
            'nav',
            this.props.patternType,
            undefined,
            newFiltersInQuery
        )
    }

    private onFilterDeleted = (filterKey: string): void => {
        const newFiltersInQuery = { ...this.props.filtersInQuery }
        delete newFiltersInQuery[filterKey]

        this.props.onFiltersInQueryChange(newFiltersInQuery)

        // Submit a search with the new values
        submitSearch(
            this.props.history,
            this.props.navbarSearchState.query,
            'nav',
            this.props.patternType,
            undefined,
            newFiltersInQuery
        )
    }

    /**
     * toggleFilterEditable updates the top-level filtersInQuery object with
     * the new `editable` state of a single filter when its edit state is
     * being toggled.
     */
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
                                    setPatternType={this.props.setPatternType}
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
