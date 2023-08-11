import { type FC, type MouseEvent, useEffect, useMemo, forwardRef } from 'react'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter } from '@sourcegraph/shared/src/search/query/token'
import { Button, ButtonGroup, type ButtonProps, Tooltip } from '@sourcegraph/wildcard'

import { GroupByField } from '../../../../../../../graphql-operations'
import type { SearchBasedInsightSeries } from '../../../../../core'

const TOOLTIP_TEXT = 'Available only for queries with type:commit and type:diff'

export interface ComputeInsightMapPickerProps {
    series: SearchBasedInsightSeries[]
    value: GroupByField
    onChange: (nextValue: GroupByField) => void
}

export const ComputeInsightMapPicker: FC<ComputeInsightMapPickerProps> = props => {
    const { series, value, onChange } = props

    const handleOptionClick = (event: MouseEvent<HTMLButtonElement>): void => {
        const target = event.target as HTMLButtonElement
        const pickedValue = target.value as GroupByField

        onChange(pickedValue)
    }

    const hasTypeDiffOrCommit = useMemo(() => {
        if (series.length === 0) {
            return false
        }

        return series.every(({ query }) => {
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
        })
    }, [series])

    useEffect(() => {
        if (!hasTypeDiffOrCommit && (value === GroupByField.AUTHOR || value === GroupByField.DATE)) {
            onChange(GroupByField.REPO)
        }
    }, [hasTypeDiffOrCommit, value, onChange])

    return (
        <ButtonGroup className="mb-3 d-block">
            <OptionButton active={value === GroupByField.REPO} value={GroupByField.REPO} onClick={handleOptionClick}>
                repository
            </OptionButton>

            <OptionButton active={value === GroupByField.PATH} value={GroupByField.PATH} onClick={handleOptionClick}>
                path
            </OptionButton>

            <Tooltip content={!hasTypeDiffOrCommit ? TOOLTIP_TEXT : undefined}>
                <OptionButton
                    active={value === GroupByField.AUTHOR}
                    value={GroupByField.AUTHOR}
                    disabled={!hasTypeDiffOrCommit}
                    onClick={handleOptionClick}
                >
                    author
                </OptionButton>
            </Tooltip>

            <Tooltip content={!hasTypeDiffOrCommit ? TOOLTIP_TEXT : undefined}>
                <OptionButton
                    active={value === GroupByField.DATE}
                    value={GroupByField.DATE}
                    disabled={!hasTypeDiffOrCommit}
                    onClick={handleOptionClick}
                >
                    date
                </OptionButton>
            </Tooltip>
        </ButtonGroup>
    )
}

interface OptionButtonProps extends ButtonProps {
    value: GroupByField
    active?: boolean
}

const OptionButton = forwardRef<HTMLButtonElement, OptionButtonProps>((props, reference) => {
    const { children, active, value, ...attributes } = props

    return (
        <Button ref={reference} {...attributes} variant="secondary" outline={!active} value={value}>
            {children}
        </Button>
    )
})
