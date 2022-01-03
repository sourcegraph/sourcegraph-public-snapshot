import classNames from 'classnames'
import Popper from 'popper.js'
import React, { ReactNode } from 'react'
import { Tooltip as BootstrapTooltip } from 'reactstrap'

import styles from './Tooltip.module.scss'
import { useTooltipState } from './useTooltipState'
import { getTooltipStyle } from './utils'

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

export const Tooltip: React.FunctionComponent<TooltipProps> = ({ className }) => {
    const { subject, content, subjectSeq, placement = 'auto', delay } = useTooltipState()

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
            popperClassName={classNames(styles.tooltip, styles.show, className, getTooltipStyle(placement))}
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
