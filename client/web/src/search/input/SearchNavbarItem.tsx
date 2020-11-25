import * as H from 'history'
import React, { useCallback, useMemo, useEffect } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../../../branded/src/components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps, OnboardingTourProps } from '..'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'
import Shepherd from 'shepherd.js'
import { defaultTourOptions, generateStepTooltip, HAS_CANCELLED_TOUR_KEY } from './SearchOnboardingTour'
import { eventLogger } from '../../tracking/eventLogger'
import { useLocalStorage } from '../../util/useLocalStorage'

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
    enableSmartQuery: boolean
}

const HAS_COMPLETED_TOUR_KEY = 'has-completed-onboarding-tour'
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
    const [hasCancelledTour, setHasCancelledTour] = useLocalStorage(HAS_CANCELLED_TOUR_KEY, false)
    const [hasCompletedTour, setHasCompletedTour] = useLocalStorage(HAS_COMPLETED_TOUR_KEY, false)

    useEffect(() => {
        tour.addSteps([
            {
                id: 'view-search-reference',
                text: generateStepTooltip({
                    tour,
                    dangerousTitleHtml: 'Review the search reference',
                    stepNumber: 5,
                    totalStepCount: 5,
                }),
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
        if (
            url.has('onboardingTour') &&
            props.showOnboardingTour &&
            hasCancelledTour !== true &&
            hasCompletedTour !== true
        ) {
            tour.show('view-search-reference')
        }
    }, [tour, props.showOnboardingTour, props.location.search, hasCancelledTour, hasCompletedTour])

    useEffect(() => {
        const onCancelled = (): void => setHasCancelledTour(true)
        const onCompleted = (): void => setHasCompletedTour(true)
        tour.on('cancel', onCancelled)
        tour.on('complete', onCompleted)
        return () => {
            tour.off('cancel', onCancelled)
            tour.off('complete', onCompleted)
        }
    }, [tour, props.location, props.history, setHasCompletedTour, setHasCancelledTour])

    useEffect(() => {
        if (hasCancelledTour || hasCompletedTour) {
            const queryParameters = new URLSearchParams(props.location.search)
            if (queryParameters.has('onboardingTour')) {
                queryParameters.delete('onboardingTour')
                props.history.replace({
                    search: queryParameters.toString(),
                    hash: props.history.location.hash,
                })
            }
        }
    }, [hasCancelledTour, hasCompletedTour, props.history, props.location.search])

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
