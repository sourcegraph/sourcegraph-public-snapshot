import React from 'react'

import { UseConnectionResult } from '../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionError,
    ConnectionLoading,
    ShowMoreButton,
    SummaryContainer,
} from '../../components/FilteredConnection/ui'

import { ConnectionPopoverContainer, ConnectionPopoverForm, ConnectionPopoverList } from './components'

interface RevisionsPopoverTabProps extends UseConnectionResult<unknown> {
    inputValue: string
    onInputChange: (value: string) => void
    query: string
    summary?: JSX.Element
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
}) => (
    <ConnectionPopoverContainer>
        <ConnectionPopoverForm
            inputValue={inputValue}
            onInputChange={event => onInputChange(event.target.value)}
            autoFocus={true}
            inputPlaceholder="Find..."
            compact={true}
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
