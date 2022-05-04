import React, { ReactNode, useMemo } from 'react'

import classNames from 'classnames'
import Popper from 'popper.js'
// eslint-disable-next-line no-restricted-imports
import { Tooltip as BootstrapTooltip } from 'reactstrap'

import { useTooltipState } from './useTooltipState'
import { getTooltipStyle } from './utils'

import styles from './Tooltip.module.scss'

interface TooltipProps {
    className?: string
    children?: ReactNode
}

const TOOLTIP_MODIFIERS: Popper.Modifiers = {
    flip: {
        enabled: false,
    },
    preventOverflow: {
        boundariesElement: 'window',
    },
}

/**
 * Renders a Tooltip that can be positioned relative to a target element.
 *
 * This component should typically only need to be rendered once in a React tree.
 * If you need to attach a tooltip to an specific element, simply add the `data-tooltip` attribute to that element.
 */
export const Tooltip: React.FunctionComponent<React.PropsWithChildren<TooltipProps>> = ({ className }) => {
    const { subject, content, subjectSeq, placement = 'auto', delay } = useTooltipState()

    const tooltipStyle = useMemo(() => getTooltipStyle(placement), [placement])

    if (!subject || !content) {
        return null
    }

    return (
        <BootstrapTooltip
            // Set key prop to work around a bug where quickly mousing between 2 elements with tooltips
            // displays the 2nd element's tooltip as still pointing to the first.
            key={subjectSeq}
            isOpen={true}
            target={subject}
            placement={placement}
            // in order to add our own placement classes we need to set the popperClassNames
            // here is where bootstrap injects it's placement classes such as 'bs-tooltip-auto' automatically.
            popperClassName={classNames(styles.tooltip, styles.show, className, tooltipStyle)}
            arrowClassName={styles.arrow}
            innerClassName={styles.tooltipInner}
            // This is a workaround to an issue with tooltips in reactstrap that causes the entire page to freeze.
            // Remove when https://github.com/reactstrap/reactstrap/issues/1482 is fixed.
            modifiers={TOOLTIP_MODIFIERS}
            delay={delay}
        >
            {content}
        </BootstrapTooltip>
    )
}
