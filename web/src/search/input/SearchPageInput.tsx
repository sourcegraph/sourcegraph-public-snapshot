import * as H from 'history'
import * as GQL from '../../../../shared/src/graphql/schema'
import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { Form } from 'reactstrap'
import { SearchModeToggle } from './interactive/SearchModeToggle'
import { VersionContextDropdown } from '../../nav/VersionContextDropdown'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { QueryInput } from './QueryInput'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR, KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { SearchButton } from './SearchButton'
import { Link } from '../../../../shared/src/components/Link'
import { SearchScopes } from './SearchScopes'
import { QuickLinks } from '../QuickLinks'
import { Notices } from '../../global/Notices'
import { SettingsCascadeProps, isSettingsValid } from '../../../../shared/src/settings/settings'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { ThemePreferenceProps } from '../../theme'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    InteractiveSearchProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    parseSearchURLQuery,
} from '..'
import { EventLoggerProps } from '../../tracking/eventLogger'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { VersionContext } from '../../schema/site.schema'
import { submitSearch, SubmitSearchParams } from '../helpers'
import {
    generateStepTooltip,
    createStep1Tooltip,
    stepCallbacks,
    HAS_SEEN_TOUR_KEY,
    HAS_CANCELLED_TOUR_KEY,
} from './SearchOnboardingTour'
import { useLocalStorage } from '../../util/useLocalStorage'
import Shepherd from 'shepherd.js'

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
        Pick<SubmitSearchParams, 'source'>,
        VersionContextProps,
        OnboardingTourProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined
    /** Whether globbing is enabled for filters. */
    globbing: boolean
    /** Whether to display the interactive mode input centered on the page, as on the search homepage. */
    interactiveModeHomepageMode?: boolean
    /** A query fragment to appear at the beginning of the input. */
    queryPrefix?: string
    autoFocus?: boolean

    // For NavLinks
    authRequired?: boolean
    showCampaigns: boolean
}

export const SearchPageInput: React.FunctionComponent<Props> = (props: Props) => {
    /** The query cursor position and value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
        cursorPosition: props.queryPrefix ? props.queryPrefix.length : 0,
    })

    useEffect(() => {
        setUserQueryState({ query: props.queryPrefix || '', cursorPosition: props.queryPrefix?.length || 0 })
    }, [props.queryPrefix])

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            // False positive
            // eslint-disable-next-line no-unused-expressions
            event?.preventDefault()

            submitSearch({ ...props, query: userQueryState.query, source: 'home' })
        },
        [props, userQueryState.query]
    )

    const [hasSeenTour, setHasSeenTour] = useLocalStorage(HAS_SEEN_TOUR_KEY, false)
    const [hasCancelledTour, setHasCancelledTour] = useLocalStorage(HAS_CANCELLED_TOUR_KEY, false)

    const isHomepage = useMemo(
        () => props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search),
        [props.location.pathname, props.location.search]
    )

    const showOnboardingTour = props.showOnboardingTour && isHomepage

    const tour = useMemo(
        () =>
            new Shepherd.Tour({
                useModalOverlay: true,
                defaultStepOptions: {
                    arrow: true,
                    classes: 'web-content tour-card card py-4 px-3',
                    popperOptions: {
                        // Removes default behavior of autofocusing steps
                        modifiers: [
                            {
                                name: 'focusAfterRender',
                                enabled: false,
                            },
                        ],
                    },
                    attachTo: { on: 'bottom' },
                    scrollTo: false,
                },
            }),
        []
    )

    useEffect(() => {
        if (showOnboardingTour) {
            tour.addSteps([
                {
                    id: 'step-1',
                    text: createStep1Tooltip(
                        tour,
                        () => {
                            setUserQueryState({ query: 'lang:', cursorPosition: 'lang:'.length })
                            tour.show('step-2-lang')
                        },
                        () => {
                            setUserQueryState({ query: 'repo:', cursorPosition: 'repo:'.length })
                            tour.show('step-2-repo')
                        }
                    ),
                    attachTo: {
                        element: '.search-page__input-container',
                        on: 'bottom',
                    },
                },
                {
                    id: 'step-2-lang',
                    text: generateStepTooltip(tour, 'Type to filter the language autocomplete', 2),
                    attachTo: {
                        element: '.search-page__input-container',
                        on: 'top',
                    },
                },
                {
                    id: 'step-2-repo',
                    text: generateStepTooltip(
                        tour,
                        "Type the name of a repository you've used recently to filter the autocomplete list",
                        2
                    ),
                    attachTo: {
                        element: '.search-page__input-container',
                        on: 'top',
                    },
                },
                // This step requires examples to be generated based on the language selected by the user,
                // so the text is generated by the callback called when the previous is completed.
                {
                    id: 'step-3',
                    attachTo: {
                        element: '.search-page__input-container',
                        on: 'bottom',
                    },
                },
                {
                    id: 'step-4',
                    text: generateStepTooltip(tour, 'Review the search reference', 4),
                    attachTo: {
                        element: '.search-help-dropdown-button',
                        on: 'bottom',
                    },
                    advanceOn: { selector: '.search-help-dropdown-button', event: 'click' },
                },
                {
                    id: 'final-step',
                    text: generateStepTooltip(tour, "Use the 'return' key or the search button to run your search", 5),
                    attachTo: {
                        element: '.search-button',
                        on: 'top',
                    },
                    advanceOn: { selector: '.search-button__btn', event: 'click' },
                },
            ])
        }
    }, [tour, showOnboardingTour])

    useEffect(() => {
        if (showOnboardingTour && !hasCancelledTour && !hasSeenTour) {
            tour.start()
        }
        return
    }, [tour, showOnboardingTour, hasCancelledTour, hasSeenTour])

    useEffect(
        () => () => {
            // End tour on unmount.
            if (tour.isActive()) {
                tour.complete()
            }
        },
        [tour]
    )

    useMemo(() => {
        tour.on('complete', () => {
            setHasSeenTour(true)
        })
        tour.on('cancel', () => {
            setHasCancelledTour(true)
        })
    }, [tour, setHasSeenTour, setHasCancelledTour])

    return (
        <div className="d-flex flex-row flex-shrink-past-contents">
            {props.splitSearchModes && props.interactiveSearchMode ? (
                <InteractiveModeInput
                    {...props}
                    navbarSearchState={userQueryState}
                    onNavbarQueryChange={setUserQueryState}
                    toggleSearchMode={props.toggleSearchMode}
                    lowProfile={false}
                    homepageMode={props.interactiveModeHomepageMode}
                />
            ) : (
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
                                    autoFocus={props.autoFocus !== false}
                                    tour={showOnboardingTour ? tour : undefined}
                                    tourAdvanceStepCallbacks={stepCallbacks}
                                />
                            ) : (
                                <QueryInput
                                    {...props}
                                    value={userQueryState}
                                    onChange={setUserQueryState}
                                    // We always want to set this to 'cursor-at-end' when true.
                                    autoFocus={props.autoFocus ? 'cursor-at-end' : props.autoFocus}
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
                            <SearchScopes
                                history={props.history}
                                query={userQueryState.query}
                                authenticatedUser={props.authenticatedUser}
                                settingsCascade={props.settingsCascade}
                                patternType={props.patternType}
                                versionContext={props.versionContext}
                            />
                        </div>
                        <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                        <Notices
                            className="my-3"
                            location="home"
                            settingsCascade={props.settingsCascade}
                            history={props.history}
                        />
                    </Form>
                </>
            )}
        </div>
    )
}
