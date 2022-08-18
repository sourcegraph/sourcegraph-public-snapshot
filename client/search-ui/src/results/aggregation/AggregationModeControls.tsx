import { FC, HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Button, ButtonGroup } from '@sourcegraph/wildcard'

import { AggregationMode } from './types'

import styles from './AggregationModeControls.module.scss'

interface AggregationModeControlsProps extends HTMLAttributes<HTMLDivElement> {
    mode: AggregationMode
    onModeChange: (nextMode: AggregationMode) => void
    size?: 'sm' | 'lg'
}

export const AggregationModeControls: FC<AggregationModeControlsProps> = props => {
    const { mode, onModeChange, size, className, ...attributes } = props

    return (
        <ButtonGroup
            {...attributes}
            aria-label="Aggregation mode picker"
            className={classNames(className, { [styles.aggregationGroupSm]: size === 'sm' })}
        >
            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationMode.Repository}
                data-testid="repo-aggregation-mode"
                onClick={() => onModeChange(AggregationMode.Repository)}
            >
                Repo
            </Button>

            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationMode.FilePath}
                data-testid="file-aggregation-mode"
                onClick={() => onModeChange(AggregationMode.FilePath)}
            >
                File
            </Button>

            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationMode.Author}
                data-testid="author-aggregation-mode"
                onClick={() => onModeChange(AggregationMode.Author)}
            >
                Author
            </Button>
            <Button
                className={styles.aggregationTypeControl}
                variant="secondary"
                size={size}
                outline={mode !== AggregationMode.CaptureGroups}
                data-testid="captureGroup-aggregation-mode"
                onClick={() => onModeChange(AggregationMode.CaptureGroups)}
            >
                Capture group
            </Button>
        </ButtonGroup>
    )
}
