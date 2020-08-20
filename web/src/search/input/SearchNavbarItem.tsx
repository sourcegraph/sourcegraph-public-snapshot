import * as H from 'history'
import React, { useCallback, useMemo, useEffect } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    parseSearchURLPatternType,
} from '..'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { QueryInput } from './QueryInput'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'
import Shepherd from 'shepherd.js'
import {
    defaultTourOptions,
    generateStepTooltip,
    HAS_SEEN_TOUR_KEY,
    HAS_CANCELLED_TOUR_KEY,
} from './SearchOnboardingTour'
import { useLocalStorage } from '../../util/useLocalStorage'
import { SearchPatternType } from '../../graphql-operations'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        SmartSearchFieldProps,
        SettingsCascadeProps,
        ThemeProps,
        CopyQueryButtonProps,
        VersionContextProps,
        OnboardingTourProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onChange: (newValue: QueryState) => void
    globbing: boolean
}

function createStructuralSearchTourTooltip(): HTMLElement {
    const listItem = document.createElement('li')
    listItem.className = 'list-group-item p-0 border-0 my-4'
    listItem.textContent = '>'
    const exampleButton = document.createElement('a')
    exampleButton.href = 'https://docs.sourcegraph.com/user/search/structural'
    exampleButton.target = '_blank'
    exampleButton.className = 'btn btn-link test-tour-language-example'
    exampleButton.textContent = 'Structural search documentation'
    listItem.append(exampleButton)
    return listItem
}

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = (props: Props) => {
    const onSubmit = useCallback((): void => {
        submitSearch({ ...props, query: props.navbarSearchState.query, source: 'nav' })
    }, [props])

    const onFormSubmit = useCallback(
        () => (event: React.FormEvent): void => {
            event.preventDefault()
            onSubmit()
        },
        [onSubmit]
    )

    const tour = useMemo(() => new Shepherd.Tour(defaultTourOptions), [])

    useEffect(() => {
        tour.addStep({
            id: 'structural-search-tip',
            text: generateStepTooltip(
                tour,
                'You ran a structural search',
                6,
                `Note that it properly matches the entire code block within the braces.\n
                It is hard to match blocks of code or multiline expressions with regex,\n
                but simple with structural search. Tip: 'my_match' is a name for the\n
                code we matched between code boundries. This is similar to a named capture\n
                group in regex.`,
                createStructuralSearchTourTooltip(),
                true
            ),
            attachTo: {
                element: '.test-structural-search-toggle',
                on: 'bottom',
            },
        })
    }, [tour])

    useEffect(() => {
        const url = new URLSearchParams(props.location.search)
        const isStructuralSearch = parseSearchURLPatternType(props.location.search) === SearchPatternType.structural
        if (url.has('onboardingTour') && isStructuralSearch && props.showOnboardingTour) {
            tour.start()
        }
    }, [tour, props.showOnboardingTour, props.location.search])

    useEffect(
        () => () => {
            // End tour on unmount.
            if (tour.isActive()) {
                tour.complete()
            }
        },
        [tour]
    )

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onFormSubmit}
        >
            {props.smartSearchField ? (
                <LazyMonacoQueryInput
                    {...props}
                    hasGlobalQueryBehavior={true}
                    queryState={props.navbarSearchState}
                    onSubmit={onSubmit}
                    autoFocus={true}
                />
            ) : (
                <QueryInput
                    {...props}
                    value={props.navbarSearchState}
                    autoFocus={props.location.pathname === '/search' ? 'cursor-at-end' : undefined}
                    keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                    hasGlobalQueryBehavior={true}
                />
            )}
            <SearchButton />
        </Form>
    )
}
