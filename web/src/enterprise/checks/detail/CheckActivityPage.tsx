import HistoryIcon from 'mdi-react/HistoryIcon'
import * as React from 'react'
import { Check } from '../data'

interface Props {
    check: Check
}

/**
 * The activity page for a single check.
 */
export const CheckActivityPage: React.FunctionComponent<Props> = ({ check }) => (
    <div className="check-activity-page">
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
