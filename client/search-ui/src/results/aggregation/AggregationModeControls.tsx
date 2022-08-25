import { FC, HTMLAttributes } from 'react'

import classNames from 'classnames'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import { SearchAggregationModeAvailability } from '../../graphql-operations'

import styles from './AggregationModeControls.module.scss'

interface AggregationModeControlsProps extends HTMLAttributes<HTMLDivElement> {
    mode: SearchAggregationMode | null
    availability?: SearchAggregationModeAvailability[]
    size?: 'sm' | 'lg'
    onModeChange: (nextMode: SearchAggregationMode) => void
}

export const AggregationModeControls: FC<AggregationModeControlsProps> = props => {
    const { mode, availability = [], onModeChange, size, className, ...attributes } = props

    const availabilityGroups = availability.reduce((store, availability) => {
        store[availability.mode] = availability

        return store
    }, {} as Partial<Record<SearchAggregationMode, SearchAggregationModeAvailability>>)

    return (
        <div
            {...attributes}
            aria-label="Aggregation mode picker"
            className={classNames(className, styles.aggregationGroup)}
        >
            <Tooltip content={availabilityGroups[SearchAggregationMode.REPO]?.reasonUnavailable}>
                <Button
                    variant="secondary"
                    size={size}
                    outline={mode !== SearchAggregationMode.REPO}
                    data-testid="repo-aggregation-mode"
                    disabled={!availabilityGroups[SearchAggregationMode.REPO]?.available}
                    onClick={() => onModeChange(SearchAggregationMode.REPO)}
                >
                    Repository
                </Button>
            </Tooltip>

            <Tooltip content={availabilityGroups[SearchAggregationMode.PATH]?.reasonUnavailable}>
                <Button
                    variant="secondary"
                    size={size}
                    outline={mode !== SearchAggregationMode.PATH}
                    disabled={!availabilityGroups[SearchAggregationMode.PATH]?.available}
                    data-testid="file-aggregation-mode"
                    onClick={() => onModeChange(SearchAggregationMode.PATH)}
                >
                    File
                </Button>
            </Tooltip>

            <Tooltip content={availabilityGroups[SearchAggregationMode.AUTHOR]?.reasonUnavailable}>
                <Button
                    variant="secondary"
                    size={size}
                    outline={mode !== SearchAggregationMode.AUTHOR}
                    disabled={!availabilityGroups[SearchAggregationMode.AUTHOR]?.available}
                    data-testid="author-aggregation-mode"
                    onClick={() => onModeChange(SearchAggregationMode.AUTHOR)}
                >
                    Author
                </Button>
            </Tooltip>

            <Tooltip content={availabilityGroups[SearchAggregationMode.CAPTURE_GROUP]?.reasonUnavailable}>
                <Button
                    variant="secondary"
                    size={size}
                    outline={mode !== SearchAggregationMode.CAPTURE_GROUP}
                    disabled={!availabilityGroups[SearchAggregationMode.CAPTURE_GROUP]?.available}
                    data-testid="captureGroup-aggregation-mode"
                    onClick={() => onModeChange(SearchAggregationMode.CAPTURE_GROUP)}
                >
                    Capture group
                </Button>
            </Tooltip>
        </div>
    )
}
