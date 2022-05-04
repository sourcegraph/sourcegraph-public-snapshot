import { FunctionComponent, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { debounce } from 'lodash'
import TrashIcon from 'mdi-react/TrashIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './ReposMatchingPattern.module.scss'

const DEBOUNCED_WAIT = 250

export interface ReposMatchingPatternProps {
    index: number
    pattern: string
    setPattern: (value: string) => void
    onDelete: () => void
    disabled: boolean
}

export const ReposMatchingPattern: FunctionComponent<React.PropsWithChildren<ReposMatchingPatternProps>> = ({
    index,
    pattern,
    setPattern,
    onDelete,
    disabled,
}) => {
    const [localPattern, setLocalPattern] = useState('')
    useEffect(() => setLocalPattern(pattern), [pattern])

    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [setPattern])

    return (
        <>
            <div className="form-group d-flex flex-column mb-0">
                <label htmlFor="repo-pattern">Repository pattern #{index + 1}</label>
                <input
                    type="text"
                    className="form-control text-monospace"
                    value={localPattern}
                    onChange={({ target: { value } }) => {
                        setLocalPattern(value)
                        debouncedSetPattern(value)
                    }}
                    disabled={disabled}
                    required={true}
                />
                <small className="form-text text-muted">Required.</small>
            </div>

            <span className={classNames(styles.button, 'd-none d-md-inline')}>
                <Button onClick={() => onDelete()} className="p-0 m-0 pt-2" disabled={disabled}>
                    <Icon className="text-danger" data-tooltip="Delete the repository pattern" as={TrashIcon} />
                </Button>
            </span>
        </>
    )
}
