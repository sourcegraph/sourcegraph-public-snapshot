import React, { useCallback } from 'react'

import classNames from 'classnames'
import { upperFirst, lowerCase } from 'lodash'

import { Input } from '@sourcegraph/wildcard'

export interface ChangesetFilterProps<T extends string> {
    label: string
    values: T[]
    selected: T | undefined
    onChange: (value: T | undefined) => void
    className?: string
}

export const ChangesetFilter = <T extends string>({
    label,
    values,
    selected,
    onChange,
    className,
}: ChangesetFilterProps<T>): React.ReactElement<ChangesetFilterProps<T>> => {
    const innerOnChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            onChange((event.target.value ?? undefined) as T | undefined)
        },
        [onChange]
    )

    return (
        <>
            <Input
                aria-label=""
                id=""
                selectClassName={classNames('changeset-filter__dropdown', className)}
                className="mb-0"
                value={selected}
                onChange={innerOnChange}
            >
                <option value="">{label}</option>
                {values.map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </Input>
        </>
    )
}
