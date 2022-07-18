import { FunctionComponent, useEffect, useMemo, useState } from 'react'

import { mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import { debounce } from 'lodash'

import { Button, Icon, Input } from '@sourcegraph/wildcard'

import styles from './ReposMatchingPattern.module.scss'

const DEBOUNCED_WAIT = 250

export interface ReposMatchingPatternProps {
    index: number
    pattern: string
    setPattern: (value: string) => void
    onDelete: () => void
    disabled: boolean
    autoFocus?: boolean
}

export const ReposMatchingPattern: FunctionComponent<React.PropsWithChildren<ReposMatchingPatternProps>> = ({
    index,
    pattern,
    setPattern,
    onDelete,
    disabled,
    autoFocus,
}) => {
    const [localPattern, setLocalPattern] = useState('')
    useEffect(() => setLocalPattern(pattern), [pattern])

    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [setPattern])

    return (
        <>
            <div className="d-flex mb-0">
                <Input
                    type="text"
                    inputClassName="text-monospace"
                    value={localPattern}
                    onChange={({ target: { value } }) => {
                        setLocalPattern(value)
                        debouncedSetPattern(value)
                    }}
                    autoFocus={autoFocus}
                    disabled={disabled}
                    required={true}
                    label={`Repository pattern #${index + 1}`}
                    message="Required."
                />
            </div>
            <span className={classNames(styles.button, 'd-none d-md-inline-flex align-items-center')}>
                <Button
                    aria-label="Delete the repository pattern"
                    variant="icon"
                    onClick={() => onDelete()}
                    className="p-0"
                    disabled={disabled}
                >
                    <Icon className="text-danger" aria-hidden={true} svgPath={mdiDelete} />
                </Button>
            </span>
        </>
    )
}
