import { FC, MouseEvent, useEffect, useMemo } from 'react'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter } from '@sourcegraph/shared/src/search/query/token'
import { Button, ButtonGroup, ButtonProps, Tooltip } from '@sourcegraph/wildcard'

import { SearchBasedInsightSeries } from '../../../../../core'
import { ComputeInsightMap } from '../types'

const TOOLTIP_TEXT = 'Available only for queries with type:commit and type:diff'

export interface ComputeInsightMapPickerProps {
    series: SearchBasedInsightSeries[]
    value: ComputeInsightMap
    onChange: (nextValue: ComputeInsightMap) => void
}

export const ComputeInsightMapPicker: FC<ComputeInsightMapPickerProps> = props => {
    const { series, value, onChange } = props

    const handleOptionClick = (event: MouseEvent<HTMLButtonElement>): void => {
        const target = event.target as HTMLButtonElement
        const pickedValue = target.value as ComputeInsightMap

        onChange(pickedValue)
    }

    const hasTypeDiffOrCommit = useMemo(
        () =>
            series.every(({ query }) => {
                const tokens = scanSearchQuery(query)

                if (tokens.type === 'success') {
                    return tokens.term
                        .filter((token): token is Filter => token.type === 'filter')
                        .some(
                            filter =>
                                resolveFilter(filter.field.value)?.type === FilterType.type &&
                                (filter.value?.value === 'diff' || filter.value?.value === 'commit')
                        )
                }

                return false
            }),
        [series]
    )

    useEffect(() => {
        if (!hasTypeDiffOrCommit && (value === ComputeInsightMap.Author || value === ComputeInsightMap.Date)) {
            onChange(ComputeInsightMap.Repositories)
        }
    }, [hasTypeDiffOrCommit, value, onChange])

    return (
        <ButtonGroup className="mb-3 d-block">
            <OptionButton
                active={value === ComputeInsightMap.Repositories}
                value={ComputeInsightMap.Repositories}
                onClick={handleOptionClick}
            >
                repository
            </OptionButton>

            <OptionButton
                active={value === ComputeInsightMap.Path}
                value={ComputeInsightMap.Path}
                onClick={handleOptionClick}
            >
                path
            </OptionButton>

            <Tooltip content={!hasTypeDiffOrCommit ? TOOLTIP_TEXT : undefined}>
                <OptionButton
                    active={value === ComputeInsightMap.Author}
                    value={ComputeInsightMap.Author}
                    disabled={!hasTypeDiffOrCommit}
                    onClick={handleOptionClick}
                >
                    author
                </OptionButton>
            </Tooltip>

            <Tooltip content={!hasTypeDiffOrCommit ? TOOLTIP_TEXT : undefined}>
                <OptionButton
                    active={value === ComputeInsightMap.Date}
                    value={ComputeInsightMap.Date}
                    disabled={!hasTypeDiffOrCommit}
                    data-tooltip={!hasTypeDiffOrCommit ? TOOLTIP_TEXT : undefined}
                    onClick={handleOptionClick}
                >
                    date
                </OptionButton>
            </Tooltip>
        </ButtonGroup>
    )
}

interface OptionButtonProps extends ButtonProps {
    value: ComputeInsightMap
    active?: boolean
}

const OptionButton: FC<OptionButtonProps> = props => {
    const { children, active, value, ...attributes } = props

    return (
        <Button {...attributes} variant="secondary" outline={!active} value={value}>
            {children}
        </Button>
    )
}
