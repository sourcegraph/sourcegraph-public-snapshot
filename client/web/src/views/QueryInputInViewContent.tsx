import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { CaseSensitivityProps, PatternTypeProps, CopyQueryButtonProps, SearchContextProps } from '../search'
import { submitSearch } from '../search/helpers'
import { SearchButton } from '../search/input/SearchButton'

interface Props
    extends SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    implicitQueryPrefix: string

    location: H.Location
    history: H.History

    /** Whether globbing is enabled for filters. */
    globbing: boolean
}

/**
 * A query input rendered in a view from an extension.
 */
export const QueryInputInViewContent: React.FunctionComponent<Props> = ({
    implicitQueryPrefix,
    settingsCascade,
    ...props
}) => {
    const [query, setQuery] = useState<string>('')
    const onQueryChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>): void => {
            setQuery(event.target.value)
        },
        [setQuery]
    )
    const onSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            submitSearch({
                ...props,
                query: `${implicitQueryPrefix} ${query}`,
                source: 'scopePage',
            })
        },
        [implicitQueryPrefix, props, query]
    )
    return (
        <Form className="d-flex" onSubmit={onSubmit}>
            <input type="text" value={query} onChange={onQueryChange} />
            <SearchButton />
        </Form>
    )
}
