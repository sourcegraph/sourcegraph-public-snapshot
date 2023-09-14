import React from 'react'

import type { UseShowMorePaginationResult } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    ConnectionError,
    ConnectionLoading,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'

import { ConnectionPopoverContainer, ConnectionPopoverForm, ConnectionPopoverList } from './components'

interface RevisionsPopoverTabProps extends UseShowMorePaginationResult<unknown, unknown> {
    inputValue: string
    onInputChange: (value: string) => void
    query: string
    summary?: JSX.Element
    inputAriaLabel: string
}

export const RevisionsPopoverTab: React.FunctionComponent<React.PropsWithChildren<RevisionsPopoverTabProps>> = ({
    children,
    inputValue,
    onInputChange,
    query,
    summary,
    error,
    loading,
    connection,
    hasNextPage,
    fetchMore,
    inputAriaLabel,
}) => (
    <ConnectionPopoverContainer>
        <ConnectionPopoverForm
            inputValue={inputValue}
            onInputChange={event => onInputChange(event.target.value)}
            autoFocus={true}
            inputPlaceholder="Find..."
            compact={true}
            inputAriaLabel={inputAriaLabel}
        />
        <SummaryContainer compact={true}>{query && summary}</SummaryContainer>
        {error && <ConnectionError errors={[error.message]} compact={true} />}
        <ConnectionPopoverList>{children}</ConnectionPopoverList>
        {loading && <ConnectionLoading compact={true} />}
        {!loading && connection && (
            <SummaryContainer compact={true}>
                {!query && summary}
                {hasNextPage && <ShowMoreButton compact={true} onClick={fetchMore} />}
            </SummaryContainer>
        )}
    </ConnectionPopoverContainer>
)
