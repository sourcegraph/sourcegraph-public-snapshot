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

    const isModeAvailable = (mode: SearchAggregationMode): boolean => {
        const isAvailable = availabilityGroups[mode]?.available

        // Returns true by default because we don't want to disable all modes
        // in case if we don't have availability information from the backend
        // (for example in case of network request failure)
        return isAvailable ?? true
    }

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
                    disabled={!isModeAvailable(SearchAggregationMode.REPO)}
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
                    disabled={!isModeAvailable(SearchAggregationMode.PATH)}
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
                    disabled={!isModeAvailable(SearchAggregationMode.AUTHOR)}
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
                    disabled={!isModeAvailable(SearchAggregationMode.CAPTURE_GROUP)}
                    data-testid="captureGroup-aggregation-mode"
                    onClick={() => onModeChange(SearchAggregationMode.CAPTURE_GROUP)}
                >
                    Capture group
                </Button>
            </Tooltip>
        </div>
    )
}
