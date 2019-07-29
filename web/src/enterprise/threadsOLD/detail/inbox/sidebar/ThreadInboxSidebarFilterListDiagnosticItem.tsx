import React, { useMemo } from 'react'
import { Diagnostic } from 'sourcegraph'
import { QueryParameterProps } from '../../../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticSeverityIcon } from '../../../../../diagnostics/components/DiagnosticSeverityIcon'
import { ThreadInboxSidebarFilterListItem } from './ThreadInboxSidebarFilterListItem'

interface Props extends QueryParameterProps {
    diagnostic: Pick<Diagnostic, 'message' | 'severity'> // TODO!(sqs): group by something other than message
    count: number

    className?: string
}

/**
 * A diagnostic group item in the thread inbox sidebar's filter list.
 */
export const ThreadInboxSidebarFilterListDiagnosticItem: React.FunctionComponent<Props> = ({
    diagnostic: { message, severity },
    count,
    ...props
}) => {
    const Icon = useMemo<React.FunctionComponent<{ className?: string }>>(
        () => ({ className }) => <DiagnosticSeverityIcon severity={severity} className={className} />,
        [severity]
    )
    return (
        <ThreadInboxSidebarFilterListItem
            {...props}
            icon={Icon}
            title={message}
            count={count} // TODO!(sqs)
        />
    )
}
