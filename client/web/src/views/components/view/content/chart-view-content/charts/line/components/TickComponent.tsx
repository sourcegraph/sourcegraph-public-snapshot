import React from 'react'

import { TickLabelProps, TickRendererProps } from '@visx/axis/lib/types'
import { Group } from '@visx/group'
import { Text } from '@visx/text'
import { TextProps } from '@visx/text/lib/Text'
import { format } from 'd3-format'
import { timeFormat } from 'd3-time-format'

// Date formatters
const SI_PREFIX_FORMATTER = format('~s')
export const numberFormatter = (number: number): string => {
    if (!Number.isInteger(number)) {
        return number.toString()
    }

    return SI_PREFIX_FORMATTER(number)
}

// Number of month day + short name of month
export const dateTickFormatter = timeFormat('%d %b')
// Year + full name of month + full name of week day
export const dateLabelFormatter = timeFormat('%d %B %A')

// Label props generators for x and y axes.
// We need separate x and y generators because we need formatted value
// depend on for which axis we generate label props
export const getTickYProps: TickLabelProps<number> = (value, index, values): Partial<TextProps> => ({
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
})
export const getTickXProps: TickLabelProps<Date> = (value, index, values): Partial<TextProps> => ({
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${dateLabelFormatter(value)}`,
})

/** Tick component displays tick label for each axis line of chart */
export const Tick: React.FunctionComponent<TickRendererProps> = props => {
    const { to, formattedValue, x: xPosition, y: yPosition, ...tickLabelProps } = props

    // Hack with Group + Text (aria hidden)
    // Because Text renders text inside another svg element and text with tspan
    // that makes another nested group for accessibility tree. To avoid group - end group
    // phrases of voice over we hide nested children from a11y tree and put explicit aria-label
    // on parent Group element with role text
    return (
        // eslint-disable-next-line jsx-a11y/aria-role
        <Group role="text" aria-label={tickLabelProps['aria-label']}>
            <Text aria-hidden={true} x={xPosition} y={yPosition} {...tickLabelProps}>
                {formattedValue}
            </Text>
        </Group>
    )
}
