import React, { Suspense } from 'react'

import { lazyComponent } from '../../util/lazyComponent'
import { submitSearch } from '../helpers'

import { MonacoQueryInputProps } from './MonacoQueryInput'
import { SearchContextDropdown } from './SearchContextDropdown'
import { Toggles } from './toggles/Toggles'

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')

/**
 * A plain query input displayed during lazy-loading of the MonacoQueryInput.
 * It has no suggestions, but still allows to type in and submit queries.
 */
export const PlainQueryInput: React.FunctionComponent<MonacoQueryInputProps> = ({
    queryState,
    autoFocus,
    onChange,
    ...props
}) => {
    const onInputChange = React.useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            onChange({ query: event.target.value })
        },
        [onChange]
    )
    return (
        <div className="query-input2 d-flex">
            {props.showSearchContext && (
                <div className="query-input2__search-context-dropdown-container">
                    <SearchContextDropdown {...props} submitSearch={submitSearch} query={queryState.query} />
                </div>
            )}
            <input
                type="text"
                autoFocus={autoFocus}
                className="form-control text-code lazy-monaco-query-input--intermediate-input"
                value={queryState.query}
                onChange={onInputChange}
                spellCheck={false}
            />
            <div className="query-input2__toggle-container">
                <Toggles {...props} navbarSearchQuery={queryState.query} />
            </div>
        </div>
    )
}

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<MonacoQueryInputProps> = props => (
    <Suspense fallback={<PlainQueryInput {...props} />}>
        <MonacoQueryInput {...props} />
    </Suspense>
)
