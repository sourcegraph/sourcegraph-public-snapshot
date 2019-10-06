import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Form } from '../Form'
import { QueryParameterProps } from '../../util/useQueryParameter'

interface Props extends QueryParameterProps {
    beforeInputFragment?: React.ReactFragment

    /**
     * Update the query immediately after each keystroke instead of waiting for the user to submit
     * the form (by pressing enter).
     */
    instant?: boolean

    className?: string
}

/**
 * The filter control for a {@link ConnectionList}.
 */
export const ConnectionListFilterQueryInput: React.FunctionComponent<Props> = ({
    query,
    onQueryChange,
    beforeInputFragment,
    instant,
    className,
}) => {
    const [uncommittedValue, setUncommittedValue] = useState(query)
    useEffect(() => {
        setUncommittedValue(query)
    }, [instant, query])
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            setUncommittedValue(e.currentTarget.value)
            if (instant) {
                onQueryChange(e.currentTarget.value)
            }
        },
        [instant, onQueryChange]
    )

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        e => {
            e.preventDefault()
            onQueryChange(uncommittedValue)
        },
        [onQueryChange, uncommittedValue]
    )

    const [isFocused, setIsFocused] = useState(false)
    const onFocus = useCallback(() => setIsFocused(true), [])
    const onBlur = useCallback(() => setIsFocused(false), [])

    const prependSearchIcon = !beforeInputFragment

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div
                className={`input-group ${prependSearchIcon ? 'bg-form-control border rounded' : ''} ${
                    prependSearchIcon && isFocused ? 'form-control-focus' : ''
                }`}
            >
                {beforeInputFragment || (
                    <div className="input-group-prepend">
                        <span className="input-group-text border-0 pl-2 pr-1 bg-transparent">
                            <SearchIcon className="icon-inline" />
                        </span>
                    </div>
                )}
                <input
                    type="text"
                    className={`form-control ${prependSearchIcon ? 'shadow-none border-0 rounded-0 pl-1' : ''}`}
                    aria-label="Filter list"
                    autoCapitalize="off"
                    value={uncommittedValue}
                    onChange={onChange}
                    onFocus={onFocus}
                    onBlur={onBlur}
                />
            </div>
        </Form>
    )
}
