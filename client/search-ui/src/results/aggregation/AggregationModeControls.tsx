import { FC, HTMLAttributes } from 'react'

import classNames from 'classnames'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import { SearchAggregationModeAvailability } from '../../graphql-operations'

import styles from './AggregationModeControls.module.scss'

interface AggregationModeControlsProps extends TelemetryProps, HTMLAttributes<HTMLDivElement> {
    mode: SearchAggregationMode | null
    availability?: SearchAggregationModeAvailability[]
    size?: 'sm' | 'lg'
    onModeChange: (nextMode: SearchAggregationMode) => void
}

export const AggregationModeControls: FC<AggregationModeControlsProps> = props => {
    const { mode, availability = [], onModeChange, size, className, telemetryService, ...attributes } = props

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

    const handleModeChange = (aggregationMode: SearchAggregationMode): void => {
        telemetryService.log(`GroupResults${aggregationMode}`)
        onModeChange(aggregationMode)
    }

    const handleModeHover = (aggregationMode: SearchAggregationMode): void => {
        if (!isModeAvailable(aggregationMode)) {
            telemetryService.log(`GroupResults${aggregationMode}DisabledHover`)
        }
    }

    return (
        <div
            {...attributes}
            aria-label="Aggregation mode picker"
            className={classNames(className, styles.aggregationGroup)}
        >
            <div
                // Div onMounterEnter is needed here because button with disabled true doesn't
                // emit any mouse or pointer events.
                onMouseEnter={() => handleModeHover(SearchAggregationMode.REPO)}
            >
                <Tooltip content={availabilityGroups[SearchAggregationMode.REPO]?.reasonUnavailable}>
                    <Button
                        variant="secondary"
                        size={size}
                        outline={mode !== SearchAggregationMode.REPO}
                        data-testid="repo-aggregation-mode"
                        disabled={!isModeAvailable(SearchAggregationMode.REPO)}
                        onClick={() => handleModeChange(SearchAggregationMode.REPO)}
                    >
                        Repository
                    </Button>
                </Tooltip>
            </div>

            <div onMouseEnter={() => handleModeHover(SearchAggregationMode.PATH)}>
                <Tooltip content={availabilityGroups[SearchAggregationMode.PATH]?.reasonUnavailable}>
                    <Button
                        variant="secondary"
                        size={size}
                        outline={mode !== SearchAggregationMode.PATH}
                        disabled={!isModeAvailable(SearchAggregationMode.PATH)}
                        data-testid="file-aggregation-mode"
                        onClick={() => handleModeChange(SearchAggregationMode.PATH)}
                    >
                        File
                    </Button>
                </Tooltip>
            </div>

            <div onMouseEnter={() => handleModeHover(SearchAggregationMode.AUTHOR)}>
                <Tooltip content={availabilityGroups[SearchAggregationMode.AUTHOR]?.reasonUnavailable}>
                    <Button
                        variant="secondary"
                        size={size}
                        outline={mode !== SearchAggregationMode.AUTHOR}
                        disabled={!isModeAvailable(SearchAggregationMode.AUTHOR)}
                        data-testid="author-aggregation-mode"
                        onClick={() => handleModeChange(SearchAggregationMode.AUTHOR)}
                    >
                        Author
                    </Button>
                </Tooltip>
            </div>

            <div onMouseEnter={() => handleModeHover(SearchAggregationMode.CAPTURE_GROUP)}>
                <Tooltip content={availabilityGroups[SearchAggregationMode.CAPTURE_GROUP]?.reasonUnavailable}>
                    <Button
                        variant="secondary"
                        size={size}
                        outline={mode !== SearchAggregationMode.CAPTURE_GROUP}
                        disabled={!isModeAvailable(SearchAggregationMode.CAPTURE_GROUP)}
                        data-testid="captureGroup-aggregation-mode"
                        onClick={() => handleModeChange(SearchAggregationMode.CAPTURE_GROUP)}
                    >
                        Capture group
                    </Button>
                </Tooltip>
            </div>
        </div>
    )
}
