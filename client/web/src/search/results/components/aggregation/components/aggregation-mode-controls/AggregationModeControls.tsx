import type { FC, HTMLAttributes } from 'react'

import classNames from 'classnames'
import { useDebouncedCallback } from 'use-debounce'

import { SearchAggregationMode } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../../../../../../featureFlags/useFeatureFlag'
import type { SearchAggregationModeAvailability } from '../../../../../../graphql-operations'

import styles from './AggregationModeControls.module.scss'

interface AggregationModeControlsProps extends HTMLAttributes<HTMLDivElement> {
    mode: SearchAggregationMode | null
    loading: boolean
    availability?: SearchAggregationModeAvailability[]
    size?: 'sm' | 'lg'
    onModeChange: (nextMode: SearchAggregationMode) => void
    onModeHover: (aggregationMode: SearchAggregationMode, available: boolean) => void
}

export const AggregationModeControls: FC<AggregationModeControlsProps> = props => {
    const { mode, loading, availability = [], size, className, onModeChange, onModeHover, ...attributes } = props

    const debouncedOnModeHover = useDebouncedCallback(onModeHover, 1000)
    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', true)

    const availabilityGroups = availability.reduce((store, availability) => {
        store[availability.mode] = availability

        return store
    }, {} as Partial<Record<SearchAggregationMode, SearchAggregationModeAvailability>>)

    const isModeAvailable = (mode: SearchAggregationMode): boolean => {
        if (loading) {
            // Prevent changing aggregation types while data is loading
            // disable all aggregation modes as we fetch the data.
            return false
        }

        const isAvailable = availabilityGroups[mode]?.available

        // Returns true by default because we don't want to disable all modes
        // in case if we don't have availability information from the backend
        // (for example in case of network request failure)
        return isAvailable ?? true
    }

    const handleModeEnter = (aggregationMode: SearchAggregationMode): void => {
        debouncedOnModeHover(aggregationMode, isModeAvailable(aggregationMode))
    }

    const handleMouseLeave = (): void => {
        debouncedOnModeHover.cancel()
    }

    return (
        <div
            {...attributes}
            aria-label="Aggregation mode picker"
            className={classNames(className, styles.aggregationGroup)}
        >
            <div
                // Div onMouseEnter is needed here because button with disabled true doesn't
                // emit any mouse or pointer events.
                onMouseEnter={() => handleModeEnter(SearchAggregationMode.REPO)}
                onMouseLeave={handleMouseLeave}
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
            </div>

            <div onMouseEnter={() => handleModeEnter(SearchAggregationMode.PATH)} onMouseLeave={handleMouseLeave}>
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
            </div>

            <div onMouseEnter={() => handleModeEnter(SearchAggregationMode.AUTHOR)} onMouseLeave={handleMouseLeave}>
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
            </div>

            <div
                onMouseEnter={() => handleModeEnter(SearchAggregationMode.CAPTURE_GROUP)}
                onMouseLeave={handleMouseLeave}
            >
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
            {enableRepositoryMetadata && (
                <div
                    onMouseEnter={() => handleModeEnter(SearchAggregationMode.REPO_METADATA)}
                    onMouseLeave={handleMouseLeave}
                >
                    <Tooltip content={availabilityGroups[SearchAggregationMode.REPO_METADATA]?.reasonUnavailable}>
                        <Button
                            variant="secondary"
                            size={size}
                            outline={mode !== SearchAggregationMode.REPO_METADATA}
                            disabled={!isModeAvailable(SearchAggregationMode.REPO_METADATA)}
                            data-testid="repoMetadata-aggregation-mode"
                            onClick={() => onModeChange(SearchAggregationMode.REPO_METADATA)}
                        >
                            Repo metadata
                        </Button>
                    </Tooltip>
                </div>
            )}
        </div>
    )
}
