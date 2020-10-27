import React, { FunctionComponent } from 'react'
import { DateTime } from '../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../components/time/Timestamp'

export interface CodeIntelOptionalTimestampProps {
    date: DateTime | null
    fallbackText: string
    now?: () => Date
}

export const CodeIntelOptionalTimestamp: FunctionComponent<CodeIntelOptionalTimestampProps> = ({
    date,
    fallbackText,
    now,
}) => (date ? <Timestamp date={date} now={now} noAbout={true} /> : <span className="text-muted">{fallbackText}</span>)
