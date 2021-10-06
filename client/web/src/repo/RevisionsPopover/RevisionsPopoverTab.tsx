import React from 'react'

import { UseConnectionResult } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionForm,
    ConnectionList,
    ConnectionLoading,
    ShowMoreButton,
    SummaryContainer,
} from '@sourcegraph/web/src/components/FilteredConnection/ui'

interface RevisionsPopoverTabProps extends UseConnectionResult<unknown> {
    inputValue: string
    onInputChange: (value: string) => void
    query: string
    summary?: JSX.Element
}

export const RevisionsPopoverTab: React.FunctionComponent<RevisionsPopoverTabProps> = ({
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
    <ConnectionContainer compact={true} className="connection-popover__content">
        <ConnectionForm
            inputValue={inputValue}
            onInputChange={event => onInputChange(event.target.value)}
            autoFocus={true}
            inputPlaceholder="Find..."
            inputClassName="connection-popover__input"
            compact={true}
        />
        <SummaryContainer compact={true}>{query && summary}</SummaryContainer>
        {error && <ConnectionError errors={[error.message]} compact={true} />}
        <ConnectionList compact={true} className="connection-popover__nodes">
            {children}
        </ConnectionList>
        {loading && <ConnectionLoading compact={true} />}
        {!loading && connection && (
            <SummaryContainer compact={true}>
                {!query && summary}
                {hasNextPage && <ShowMoreButton compact={true} onClick={fetchMore} />}
            </SummaryContainer>
        )}
    </ConnectionContainer>
)
