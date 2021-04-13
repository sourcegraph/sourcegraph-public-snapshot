import * as H from 'history'
import React, { useCallback } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import {
    PatternTypeProps,
    CaseSensitivityProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    SearchContextProps,
} from '..'
import { submitSearch, QueryState } from '../helpers'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { useSearchOnboardingTour } from './SearchOnboardingTour'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ThemeProps,
        CopyQueryButtonProps,
        SearchContextProps,
        VersionContextProps,
        OnboardingTourProps {
    location: H.Location
    history: H.History
    navbarSearchState: QueryState
    onChange: (newValue: QueryState) => void
    globbing: boolean
    enableSmartQuery: boolean
    isSearchAutoFocusRequired?: boolean
}

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = (props: Props) => {
    const autoFocus = props.isSearchAutoFocusRequired ?? true

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearch({ ...props, query: props.navbarSearchState.query, source: 'nav' })
        },
        [props]
    )
    const onboardingTourQueryInputProps = useSearchOnboardingTour({
        ...props,
        inputLocation: 'global-navbar',
        queryState: props.navbarSearchState,
        setQueryState: props.onChange,
    })

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onSubmit}
        >
            <LazyMonacoQueryInput
                {...props}
                {...onboardingTourQueryInputProps}
                hasGlobalQueryBehavior={true}
                queryState={props.navbarSearchState}
                onSubmit={onSubmit}
                autoFocus={autoFocus}
                showSearchContextHighlightTourStep={true}
            />
            <SearchButton />
        </Form>
    )
}
