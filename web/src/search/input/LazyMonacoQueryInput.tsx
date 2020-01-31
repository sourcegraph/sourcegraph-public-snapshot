import React, { Suspense } from 'react'
import { MonacoQueryInputProps } from './MonacoQueryInput'
import { lazyComponent } from '../../util/lazyComponent'
import { CaseSensitivityToggle } from './CaseSensitivityToggle'
import { RegexpToggle } from './RegexpToggle'

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')

/**
 * A plain query input displayed during lazy-loading of the MonacoQueryInput.
 * It has no suggestions, but still allows to type in and submit queries.
 */
const PlainQueryInput: React.FunctionComponent<MonacoQueryInputProps> = ({
    queryState,
    autoFocus,
    onChange,
    ...props
}) => {
    const onInputChange = React.useCallback(
        (e: React.ChangeEvent<HTMLInputElement>) => {
            // cursorPosition is only used for legacy suggestions, it's OK to set it to 0 here.
            onChange({ query: e.target.value, cursorPosition: 0 })
        },
        [onChange]
    )
    return (
        <div className="query-input2 d-flex">
            <input
                type="text"
                autoFocus={autoFocus}
                className="form-control query-input2__input e2e-query-input"
                value={queryState.query}
                onChange={onInputChange}
                spellCheck={false}
            />
            <div className="query-input2__toggle-container">
                <CaseSensitivityToggle {...props} navbarSearchQuery={queryState.query}></CaseSensitivityToggle>
                <RegexpToggle {...props} navbarSearchQuery={queryState.query}></RegexpToggle>
            </div>
        </div>
    )
}

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<MonacoQueryInputProps> = props => (
    <Suspense fallback={<PlainQueryInput {...props} />}>
        <MonacoQueryInput {...props}></MonacoQueryInput>
    </Suspense>
)
