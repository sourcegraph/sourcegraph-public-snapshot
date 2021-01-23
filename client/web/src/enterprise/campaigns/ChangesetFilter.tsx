import React from 'react'
import classNames from 'classnames'
import { upperFirst, lowerCase } from 'lodash'

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
}: ChangesetFilterProps<T>): React.ReactElement<ChangesetFilterProps<T>> => (
    <>
        <select
            className={classNames('form-control changeset-filter__dropdown', className)}
            value={selected}
            onChange={event => onChange((event.target.value ?? undefined) as T | undefined)}
        >
            <option value="">{label}</option>
            {values.map(state => (
                <option value={state} key={state}>
                    {upperFirst(lowerCase(state))}
                </option>
            ))}
        </select>
    </>
)
