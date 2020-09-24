import * as H from 'history'
import React, { useCallback, useMemo, useEffect } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    parseSearchURLPatternType,
} from '..'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'
import Shepherd from 'shepherd.js'
import { defaultTourOptions, generateStepTooltip, createStructuralSearchTourTooltip } from './SearchOnboardingTour'
import { SearchPatternType } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
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

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = (props: Props) => {
    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearch({ ...props, query: props.navbarSearchState.query, source: 'nav' })
        },
        [props]
    )

    const tour = useMemo(() => new Shepherd.Tour(defaultTourOptions), [])

    useEffect(() => {
        tour.addSteps([
            {
                id: 'structural-search-tip',
                text: generateStepTooltip(
                    tour,
                    'You ran a structural search',
                    5,
                    6,
                    `Note that it properly matches the entire code block within the braces.\n
                It is hard to match blocks of code or multiline expressions with regex,\n
                but simple with structural search. Tip: 'my_match' is a name for the\n
                code we matched between code boundries. This is similar to a named capture\n
                group in regex.`,
                    createStructuralSearchTourTooltip(tour)
                ),
                when: {
                    show() {
                        eventLogger.log('ViewedOnboardingTourStructuralSearchStep')
                    },
                },
                attachTo: {
                    element: '.test-structural-search-toggle',
                    on: 'bottom',
                },
            },
            {
                id: 'view-search-reference',
                text: generateStepTooltip(tour, 'Review the search reference', 5, 5),
                attachTo: {
                    element: '.search-help-dropdown-button',
                    on: 'bottom',
                },
                when: {
                    show() {
                        eventLogger.log('ViewedOnboardingTourSearchReferenceStep')
                    },
                },
                advanceOn: { selector: '.search-help-dropdown-button', event: 'click' },
            },
        ])
    }, [tour])

    useEffect(() => {
        const url = new URLSearchParams(props.location.search)
        const isStructuralSearch = parseSearchURLPatternType(props.location.search) === SearchPatternType.structural
        if (url.has('onboardingTour') && props.showOnboardingTour) {
            if (isStructuralSearch) {
                tour.show('structural-search-tip')
            } else {
                tour.show('view-search-reference')
            }
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
            onSubmit={onSubmit}
        >
            <LazyMonacoQueryInput
                {...props}
                hasGlobalQueryBehavior={true}
                queryState={props.navbarSearchState}
                onSubmit={onSubmit}
                autoFocus={true}
            />
            <SearchButton />
        </Form>
    )
}
