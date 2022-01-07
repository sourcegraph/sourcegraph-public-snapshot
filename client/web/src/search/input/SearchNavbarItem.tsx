import { Shortcut } from '@slimsag/react-shortcuts'
import * as H from 'history'
import React, { useCallback, useState, useEffect } from 'react'
import shallow from 'zustand/shallow'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { FuzzyFinder } from '@sourcegraph/web/src/components/fuzzyFinder/FuzzyFinder'

import { PatternTypeProps, SearchContextInputProps, parseSearchURLQuery } from '..'
import { AuthenticatedUser } from '../../auth'
import { KEYBOARD_SHORTCUT_FUZZY_FINDER } from '../../keyboardShortcuts/keyboardShortcuts'
import { useExperimentalFeatures, useNavbarQueryState } from '../../stores'
import { NavbarQueryState } from '../../stores/navbarSearchQueryState'
import { getExperimentalFeatures } from '../../util/get-experimental-features'
import { SubmitSearchParameters } from '../helpers'

import { SearchBox } from './SearchBox'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        SettingsCascadeProps,
        ThemeProps,
        SearchContextInputProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    globbing: boolean
    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
}

const selectQueryState = ({
    queryState,
    setQueryState,
    submitSearch,
    searchCaseSensitivity,
    setSearchCaseSensitivity,
}: NavbarQueryState): Pick<
    NavbarQueryState,
    'queryState' | 'setQueryState' | 'submitSearch' | 'searchCaseSensitivity' | 'setSearchCaseSensitivity'
> => ({ queryState, setQueryState, submitSearch, searchCaseSensitivity, setSearchCaseSensitivity })

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = (props: Props) => {
    const autoFocus = props.isSearchAutoFocusRequired ?? true
    // This uses the same logic as in Layout.tsx until we have a better solution
    // or remove the search help button
    const isSearchPage = props.location.pathname === '/search' && Boolean(parseSearchURLQuery(props.location.search))
    const [isFuzzyFinderVisible, setIsFuzzyFinderVisible] = useState(false)
    const {
        queryState,
        setQueryState,
        submitSearch,
        searchCaseSensitivity,
        setSearchCaseSensitivity,
    } = useNavbarQueryState(selectQueryState, shallow)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            submitSearch({
                history: props.history,
                patternType: props.patternType,
                source: 'nav',
                activation: props.activation,
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                ...parameters,
            })
        },
        [submitSearch, props.history, props.patternType, props.activation, props.selectedSearchContextSpec]
    )

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearchOnChange()
        },
        [submitSearchOnChange]
    )

    const [retainFuzzyFinderCache, setRetainFuzzyFinderCache] = useState(true)
    useEffect(() => {
        if (isSearchPage && isFuzzyFinderVisible) {
            setIsFuzzyFinderVisible(false)
        }
    }, [isSearchPage, isFuzzyFinderVisible])

    const { fuzzyFinder, fuzzyFinderCaseInsensitiveFileCountThreshold } = getExperimentalFeatures(
        props.settingsCascade.final
    )

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onSubmit}
        >
            <SearchBox
                {...props}
                showSearchContext={showSearchContext}
                showSearchContextManagement={showSearchContextManagement}
                caseSensitive={searchCaseSensitivity}
                setCaseSensitivity={setSearchCaseSensitivity}
                queryState={queryState}
                onChange={setQueryState}
                onSubmit={onSubmit}
                submitSearchOnToggle={submitSearchOnChange}
                submitSearchOnSearchContextChange={submitSearchOnChange}
                autoFocus={autoFocus}
                hideHelpButton={isSearchPage}
                onHandleFuzzyFinder={setIsFuzzyFinderVisible}
            />
            <Shortcut
                {...KEYBOARD_SHORTCUT_FUZZY_FINDER.keybindings[0]}
                onMatch={() => {
                    setIsFuzzyFinderVisible(true)
                    setRetainFuzzyFinderCache(true)
                    const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
                    input?.focus()
                    input?.select()
                }}
            />
            {props.isRepositoryRelatedPage && retainFuzzyFinderCache && fuzzyFinder && (
                <FuzzyFinder
                    caseInsensitiveFileCountThreshold={fuzzyFinderCaseInsensitiveFileCountThreshold}
                    setIsVisible={bool => setIsFuzzyFinderVisible(bool)}
                    isVisible={isFuzzyFinderVisible}
                    telemetryService={props.telemetryService}
                    location={props.location}
                    setCacheRetention={bool => setRetainFuzzyFinderCache(bool)}
                />
            )}
        </Form>
    )
}
