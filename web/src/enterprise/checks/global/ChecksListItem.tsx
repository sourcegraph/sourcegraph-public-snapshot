import H from 'history'
import CheckboxMultipleBlankOutlineIcon from 'mdi-react/CheckboxMultipleBlankOutlineIcon'
import MessageOutlineIcon from 'mdi-react/MessageOutlineIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Check } from '../data'

interface Props {
    check: Check
    location: H.Location
}

/**
 * A list item for a check in {@link ChecksList}.
 */
export const ChecksListItem: React.FunctionComponent<Props> = ({ check, location }) => (
    <li className="list-group-item p-2">
        <div className="d-flex align-items-start">
            <div
                className="form-check mx-2"
                /* tslint:disable-next-line:jsx-ban-props */
                style={{ marginTop: '2px' /* stylelint-disable-line declaration-property-unit-whitelist */ }}
            >
                <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
            </div>
            <CheckboxMultipleBlankOutlineIcon
                className={`icon-inline small mr-2 mt-1 ${checkIconColorClass(check)}`}
                data-tooltip={checkIconTooltip(check)}
            />
            <div className="flex-1">
                <h3 className="d-flex align-items-center mb-0">
                    {/* tslint:disable-next-line:jsx-ban-props */}
                    <Link to={`${location.pathname}/${check.id}`} style={{ color: 'var(--body-color)' }}>
                        {check.title}
                    </Link>
                    {check.count > 1 && <span className="badge badge-secondary ml-1">{check.count}</span>}
                </h3>

                {check.labels && (
                    <div>
                        {check.labels.map((label, i) => (
                            <span key={i} className={`badge mr-1 ${badgeColorClass(label)}`}>
                                {label}
                            </span>
                        ))}
                    </div>
                )}
            </div>
            <div>
                <ul className="list-inline d-flex align-items-center">
                    {check.messageCount > 0 && (
                        <li className="list-inline-item">
                            <small className="text-muted">
                                <MessageOutlineIcon className="icon-inline" /> {check.messageCount}
                            </small>
                        </li>
                    )}
                </ul>
            </div>
        </div>
    </li>
)

function checkIconColorClass({ status }: Pick<Check, 'status'>): string {
    switch (status) {
        case 'open':
            return 'text-danger'
        case 'closed':
            return 'text-success'
        case 'disabled':
            return 'text-muted'
    }
}

function checkIconTooltip({ status }: Pick<Check, 'status'>): string {
    switch (status) {
        case 'open':
            return 'Open check (needs attention)'
        case 'closed':
            return 'Closed check (no action needed)'
        case 'disabled':
            return 'Disabled check'
    }
}

function badgeColorClass(label: string): string {
    if (label === 'security' || label.endsWith('sec')) {
        return 'badge-danger'
    }
    const CLASSES = ['badge-primary', 'badge-warning', 'badge-info', 'badge-success']
    const k = label.split('').reduce((sum, c) => (sum += c.charCodeAt(0)), 0)
    return CLASSES[k % (CLASSES.length - 1)]
}
