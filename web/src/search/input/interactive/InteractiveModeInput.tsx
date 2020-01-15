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
import { uniqueId } from 'lodash'

interface InteractiveModeProps
    extends SettingsCascadeProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
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
    /** Count of the total number of filters ever added in that component.*/
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
                filtersInQuery[uniqueId(t)] = { type: t, value: item, editable: false }
            })
        }

        that.props.onFiltersInQueryChange(filtersInQuery)
    }

    /**
     * Adds a new filter to the top-level filtersInQuery state field.
     */
    private addNewFilter = (filterType: SuggestionTypes): void => {
        const filterKey = uniqueId(filterType)
        that.props.onFiltersInQueryChange({
            ...that.props.filtersInQuery,
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
            ...that.props.filtersInQuery,
            [filterKey]: {
                ...that.props.filtersInQuery[filterKey],
                value,
                editable: false,
            },
        }
        // Update the top-level filtersInQuery with the new values
        that.props.onFiltersInQueryChange(newFiltersInQuery)

        // Submit a search with the new values
        submitSearch(
            that.props.history,
            that.props.navbarSearchState.query,
            'nav',
            that.props.patternType,
            undefined,
            newFiltersInQuery
        )
    }

    private onFilterDeleted = (filterKey: string): void => {
        const newFiltersInQuery = { ...that.props.filtersInQuery }
        delete newFiltersInQuery[filterKey]

        that.props.onFiltersInQueryChange(newFiltersInQuery)

        // Submit a search with the new values
        submitSearch(
            that.props.history,
            that.props.navbarSearchState.query,
            'nav',
            that.props.patternType,
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
        that.props.onFiltersInQueryChange({
            ...that.props.filtersInQuery,
            [filterKey]: {
                ...that.props.filtersInQuery[filterKey],
                editable: !that.props.filtersInQuery[filterKey].editable,
            },
        })
    }

    private onSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()

        submitSearch(
            that.props.history,
            that.props.navbarSearchState.query,
            'nav',
            that.props.patternType,
            undefined,
            that.props.filtersInQuery
        )
    }

    public render(): JSX.Element | null {
        const isSearchHomepage =
            that.props.location.pathname === '/search' && !parseSearchURLQuery(that.props.location.search, true)

        let logoSrc = '/.assets/img/sourcegraph-mark.svg'
        let logoLinkClassName = 'global-navbar__logo-link global-navbar__logo-animated'

        const { branding } = window.context
        if (branding) {
            if (that.props.isLightTheme) {
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
                        (that.props.authRequired ? (
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
                        <Form onSubmit={that.onSubmit} className="flex-grow-1">
                            <div className="d-flex align-items-start">
                                <SearchModeToggle {...that.props} interactiveSearchMode={true} />
                                <QueryInput
                                    location={that.props.location}
                                    history={that.props.history}
                                    value={that.props.navbarSearchState}
                                    hasGlobalQueryBehavior={true}
                                    onChange={that.props.onNavbarQueryChange}
                                    patternType={that.props.patternType}
                                    setPatternType={that.props.setPatternType}
                                    autoFocus={true}
                                    filterQuery={that.props.filtersInQuery}
                                    withoutSuggestions={true}
                                    withSearchModeToggle={true}
                                />
                                <SearchButton noHelp={true} />
                            </div>
                        </Form>
                    </div>
                    {!that.props.authRequired && !isSearchHomepage && (
                        <NavLinks {...that.props} showDotComMarketing={showDotComMarketing} />
                    )}
                </div>
                <div>
                    <SelectedFiltersRow
                        filtersInQuery={that.props.filtersInQuery}
                        navbarQuery={that.props.navbarSearchState}
                        onSubmit={that.onSubmit}
                        onFilterEdited={that.onFilterEdited}
                        onFilterDeleted={that.onFilterDeleted}
                        toggleFilterEditable={that.toggleFilterEditable}
                        isHomepage={isSearchHomepage}
                    />
                    <AddFilterRow onAddNewFilter={that.addNewFilter} isHomepage={isSearchHomepage} />
                </div>
            </div>
        )
    }
}
