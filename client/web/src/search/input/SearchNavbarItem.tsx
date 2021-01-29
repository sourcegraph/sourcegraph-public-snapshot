import * as H from 'history'
import React, { useCallback } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { Form } from '../../../../branded/src/components/Form'
import { submitSearch, QueryState } from '../helpers'
import { SearchButton } from './SearchButton'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps, OnboardingTourProps } from '..'
import { ThemeProps } from '../../../../shared/src/theme'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { useSearchOnboardingTour } from './SearchOnboardingTour'

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
                autoFocus={true}
            />
            <SearchButton />
        </Form>
    )
}
