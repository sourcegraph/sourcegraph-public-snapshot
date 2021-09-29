import classNames from 'classnames'
import * as H from 'history'
import React, { useState, useCallback, useEffect } from 'react'
import { Form } from 'reactstrap'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../auth'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import { VersionContext } from '../../schema/site.schema'
import { PatternTypeProps, CaseSensitivityProps, ParsedSearchQueryProps, SearchContextInputProps } from '../../search'
import { submitSearch, SubmitSearchParameters } from '../../search/helpers'
import { SearchBox } from '../../search/input/SearchBox'
import { ThemePreferenceProps } from '../../theme'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        Pick<SubmitSearchParameters, 'source'>,
        VersionContextProps,
        SearchContextInputProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
    /** Whether globbing is enabled for filters. */
    globbing: boolean
    /** A query fragment to appear at the beginning of the input. */
    queryPrefix?: string
    /** A query fragment to be prepended to queries. This will not appear in the input until a search is submitted. */
    hiddenQueryPrefix?: string
    className?: string
}

export const TreePageSearchInput: React.FunctionComponent<Props> = (props: Props) => {
    /** The value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
    })

    useEffect(() => {
        setUserQueryState({ query: props.queryPrefix || '' })
    }, [props.queryPrefix])

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            event?.preventDefault()
            submitSearch({
                ...props,
                query: props.hiddenQueryPrefix
                    ? `${props.hiddenQueryPrefix} ${userQueryState.query}`
                    : userQueryState.query,
                source: 'repo',
            })
        },
        [props, userQueryState.query]
    )

    return (
        <div className={classNames('d-flex flex-row flex-shrink-past-contents', props.className)}>
            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                <div className="d-flex w-100">
                    <SearchBox
                        showSearchContextFeatureTour={false}
                        {...props}
                        submitSearchOnSearchContextChange={false}
                        hasGlobalQueryBehavior={false}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        autoFocus={false}
                        isSearchOnboardingTourVisible={false}
                        hideVersionContexts={true}
                        hideHelpButton={true}
                    />
                </div>
            </Form>
        </div>
    )
}
