import { FC } from 'react'

import { TickRendererProps } from '@visx/axis'
import { Group } from '@visx/group'
import { TextProps } from '@visx/text'
import classNames from 'classnames'

import styles from './Tick.module.scss'

export interface TickProps extends TickRendererProps {
    getTruncatedTick?: (lable: string) => string
}

/** Tick component displays tick label for each axis line of chart. */
export const Tick: FC<TickProps> = props => {
    const {
        children,
        formattedValue = '',
        className,
        'aria-label': ariaLabel,
        getTruncatedTick,
        ...tickLabelProps
    } = props

    // Formatted and truncated tick value
    const tickValue = getTruncatedTick ? getTruncatedTick(formattedValue) : formattedValue

    // Empty tick value breaks the tick axis container measurement
    // And therefore breaks the whole chart content calculation
    // see https://github.com/sourcegraph/sourcegraph/issues/41158
    const sanitizedTickValue = tickValue.trim() === '' ? '&nbsp;' : tickValue

    // Hack with Group + Text (aria hidden)
    // Because the Text component renders text inside svg element and text element with tspan
    // this makes another nested group for a11y tree. To avoid "group - end group"
    // phrase in voice over we hide nested children from a11y tree and put explicit aria-label
    // on the parent Group element with role text
    return (
        // eslint-disable-next-line jsx-a11y/aria-role
        <Group role="text" aria-label={ariaLabel}>
            <text
                aria-hidden={true}
                className={classNames(styles.tick, className)}
                {...(tickLabelProps as TextProps)}
                dangerouslySetInnerHTML={{ __html: sanitizedTickValue }}
            />
        </Group>
    )
}

/**
 * Text (labels) ticks measure helper. Since there is no way to measure text
 * before rendering inside React tree we have to conduct pre-rendering measurements
 * for ticks labels.
 *
 * It renders each labels (text tick) inside selection element with SVG text element
 * and measures its sizes.
 */
export const getMaxTickWidth = (selection: Element, labels: string[]): number => {
    const tester = document.createElementNS('http://www.w3.org/2000/svg', 'text')

    // In order to sync Tick component and pre-rendering text styles which is vital for
    // text measurements
    tester.classList.add(styles.tick)
    selection.append(tester)

    const boundingBoxes = labels.map(label => {
        tester.textContent = label

        return tester.getBBox()
    })

    tester.remove()

    return Math.max(...boundingBoxes.map(b => b.width))
}
