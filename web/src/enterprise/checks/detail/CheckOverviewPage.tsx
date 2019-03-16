import HistoryIcon from 'mdi-react/HistoryIcon'
import * as React from 'react'
import { Check } from '../data'

interface Props {
    check: Check
}

/**
 * The overview page for a single check.
 *
 * TODO(sqs): figure out how this interacts with changes - it seems the check would find multiple
 * hits and you might want to group them arbitrarily into batches that you will address - that is a
 * "change".
 */
export const CheckOverviewPage: React.FunctionComponent<Props> = ({ check }) => (
    <div className="check-overview-page">
        <ul className="list-inline d-flex align-items-center mb-1">
            <li className="list-inline-item">
                <small className="text-muted">
                    <HistoryIcon className="icon-inline" />
                    {check.timeAgo} by {check.author} in <code>{check.commitID}</code>
                </small>
            </li>
        </ul>
    </div>
)
