import { FC, HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup } from '@sourcegraph/wildcard'

import { AggregationModes } from './types'

import styles from './AggregationModeControls.module.scss'

interface AggregationModeControlsProps extends HTMLAttributes<HTMLDivElement> {
    mode: AggregationModes
    onModeChange: (nextMode: AggregationModes) => void
    size?: 'sm' | 'lg'
}

export const AggregationModeControls: FC<AggregationModeControlsProps> = props => {
    const { mode, onModeChange, size, className, ...attributes } = props

    return (
        <ButtonGroup {...attributes} className={classNames(className, { [styles.aggregationGroupSm]: size === 'sm' })}>
            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationModes.Repository}
                onClick={() => onModeChange(AggregationModes.Repository)}
            >
                Repo
            </Button>

            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationModes.FilePath}
                onClick={() => onModeChange(AggregationModes.FilePath)}
            >
                File
            </Button>

            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationModes.Author}
                onClick={() => onModeChange(AggregationModes.Author)}
            >
                Author
            </Button>
            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationModes.CaptureGroups}
                onClick={() => onModeChange(AggregationModes.CaptureGroups)}
            >
                Capture group
            </Button>
        </ButtonGroup>
    )
}
