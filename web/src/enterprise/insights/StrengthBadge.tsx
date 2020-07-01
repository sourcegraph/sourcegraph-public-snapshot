import React, { useState, useCallback, useMemo } from 'react'
import { Dropdown, DropdownToggle, DropdownItem, DropdownMenu } from 'reactstrap'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import AlertIcon from 'mdi-react/AlertIcon'
import { isDefined } from '../../../../shared/src/util/types'

interface Props {
    value: number | undefined
    what: 'Code' | 'Commit'
    className?: string
}

const THRESHOLD = 0.7

const labelForValue = (value: number): string => {
    if (value < THRESHOLD) {
        return 'Weak'
    }
    return 'Strong'
}

const bgClassForValue = (value: number): string =>
    value < THRESHOLD / 2 ? 'danger' : value < THRESHOLD ? 'warning' : 'success'

const StrengthItem: React.FunctionComponent<{ value: number }> = ({ value, children }) => {
    const Icon: React.ComponentType<{ className?: string }> =
        value < THRESHOLD / 2 ? CloseIcon : value < THRESHOLD ? AlertIcon : CheckIcon
    return (
        <DropdownItem tag="div" className="d-flex align-items-center" style={{ cursor: 'pointer' }}>
            <Icon className={`icon-inline mr-2 text-${bgClassForValue(value)}`} />{' '}
            <span className="flex-1">{children}</span>
        </DropdownItem>
    )
}

const fuzz = (value: number, factor: number) => Math.max(Math.min(value * factor, 1), 0)

export const StrengthBadge: React.FunctionComponent<Props> = ({ value: valueInput, what, className = '' }) => {
    const value = useMemo(() => (valueInput === undefined ? Math.random() : valueInput), [])

    type Item = { value: number; weakLabel: string; strongLabel: string }
    const itemValue = value > THRESHOLD ? 99 : value
    const items = useMemo<Item[]>(
        () =>
            [
                what === 'Commit'
                    ? {
                          value: fuzz(itemValue, 0.5 + Math.random()),
                          weakLabel: 'Not reviewed',
                          strongLabel: 'Reviewed',
                      }
                    : undefined,
                what === 'Code'
                    ? {
                          value: fuzz(itemValue, 0.5 + Math.random()),
                          weakLabel: 'Not reviewed in last year',
                          strongLabel: 'Reviewed in last year',
                      }
                    : undefined,
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'No code owners',
                    strongLabel: 'Valid code owners',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Poor test coverage',
                    strongLabel: 'Acceptable test coverage',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'No automatic deployment',
                    strongLabel: 'Continuously deployed',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Not scanned by SAST',
                    strongLabel: 'Passed security scan',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Potential PII issues',
                    strongLabel: 'Passed PII scan',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Licensing issues',
                    strongLabel: 'Passed licensing scan',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'No linter configured',
                    strongLabel: 'Passed lint checks',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Single maintainer',
                    strongLabel: 'Multiple maintainers',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Dependencies are stale',
                    strongLabel: 'Dependencies auto-updated',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'No deprecation SLA',
                    strongLabel: 'Deprecation SLA in place',
                },
                {
                    value: fuzz(itemValue, 0.5 + Math.random()),
                    weakLabel: 'Overdue for reg/compliance check',
                    strongLabel: 'Passed compliance check',
                },
            ].filter(isDefined),

        [value]
    )

    const [isOpen, setIsOpen] = useState(false)
    const toggle = useCallback(() => setIsOpen(prevValue => !prevValue), [])
    return (
        <Dropdown toggle={toggle} isOpen={isOpen}>
            <DropdownToggle
                tag="div"
                className={`badge badge-${bgClassForValue(value)} ${className}`}
                style={{ width: '3rem' }}
            >
                {labelForValue(value)}
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem header={true} className={`font-weight-bold text-${bgClassForValue(value)}`}>
                    {what} is {labelForValue(value).toLowerCase()}
                </DropdownItem>
                <DropdownItem divider={true} />
                {items.map((item, i) => (
                    <StrengthItem key={i} value={item.value}>
                        {item.value < THRESHOLD ? item.weakLabel : item.strongLabel}
                    </StrengthItem>
                ))}
            </DropdownMenu>
        </Dropdown>
    )
}
