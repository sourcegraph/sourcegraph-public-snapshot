import * as H from 'history'
import React, { useCallback } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    SearchContextInputProps,
    parseSearchURLQuery,
} from '..'
import { AuthenticatedUser } from '../../auth'
import { VersionContext } from '../../schema/site.schema'
import { submitSearch } from '../helpers'
import { useNavbarQueryState } from '../navbarSearchQueryState'

import { SearchBox } from './SearchBox'

interface Props
    extends ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        ThemeProps,
        SearchContextInputProps,
        VersionContextProps,
        OnboardingTourProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    globbing: boolean
    isSearchAutoFocusRequired?: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
}

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<Props> = (props: Props) => {
    const autoFocus = props.isSearchAutoFocusRequired ?? true
    // This uses the same logic as in Layout.tsx until we have a better solution
    // or remove the search help button
    const isSearchPage = props.location.pathname === '/search' && Boolean(parseSearchURLQuery(props.location.search))

    const { queryState, setQueryState } = useNavbarQueryState()

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearch({ ...props, query: queryState.query, source: 'nav' })
        },
        [props, queryState]
    )

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onSubmit}
        >
            <SearchBox
                {...props}
                hasGlobalQueryBehavior={true}
                queryState={queryState}
                onChange={setQueryState}
                onSubmit={onSubmit}
                autoFocus={autoFocus}
                showSearchContextFeatureTour={true}
                isSearchOnboardingTourVisible={false}
                hideHelpButton={isSearchPage}
            />
        </Form>
    )
}
