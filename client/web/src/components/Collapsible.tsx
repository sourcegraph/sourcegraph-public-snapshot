import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon } from '@sourcegraph/wildcard'

import styles from './Collapsible.module.scss'

interface Props {
    /**
     * Content in the always-visible title bar.
     */
    title: React.ReactNode

    /**
     * Sub-content always visible in the title bar.
     */
    detail?: string | React.ReactElement

    /**
     * Optional children that appear below the title bar that can be expanded/collapsed. If present,
     * a button that expands or collapses the children will be shown.
     */
    children?: React.ReactNode

    /**
     * Whether the children are expanded and visible by default.
     */
    defaultExpanded?: boolean

    className?: string
    titleClassName?: string
    buttonClassName?: string
    expandedButtonClassName?: string
    detailClassName?: string

    /**
     * Whether the whole title section should be clickable to expand the content
     */
    wholeTitleClickable?: boolean

    /**
     * Whether the title should be placed before the chevron icon.
     */
    titleAtStart?: true
}

/**
 * Collapsible is an element with a title that is always displayed and children that are displayed
 * only when expanded.
 */
export const Collapsible: React.FunctionComponent<Props> = ({
    title,
    detail,
    children,
    titleAtStart = false,
    defaultExpanded = false,
    className = '',
    titleClassName = '',
    detailClassName = '',
    buttonClassName = '',
    expandedButtonClassName = '',
    wholeTitleClickable = true,
    ...rest
}) => {
    const [isExpanded, setIsExpanded] = useState(defaultExpanded)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const titleNode = detail ? (
        <div className="d-flex flex-column">
            <span className={titleClassName}>{title}</span>
            {detail && <div className={detailClassName}>{detail}</div>}
        </div>
    ) : (
        <span className={titleClassName}>{title}</span>
    )

    return (
        <div className={className} {...rest}>
            <div
                className={classNames(
                    'd-flex justify-content-between align-items-center position-relative',
                    isExpanded && expandedButtonClassName,
                    buttonClassName
                )}
            >
                {titleAtStart && titleNode}
                <Button
                    variant="icon"
                    className={classNames('d-flex', styles.expandBtn, wholeTitleClickable && 'stretched-link')}
                    aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                    onClick={toggleIsExpanded}
                >
                    <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                </Button>
                {!titleAtStart && titleNode}
            </div>
            {isExpanded && children}
        </div>
    )
}
