import { bottomTickLabelProps } from '@visx/axis/lib/axis/AxisBottom'
import { leftTickLabelProps } from '@visx/axis/lib/axis/AxisLeft'
import { TickLabelProps, TickRendererProps } from '@visx/axis/lib/types'
import { Group } from '@visx/group'
import { Text } from '@visx/text'
import { TextProps } from '@visx/text/lib/Text'
import React from 'react'

import { formatXLabel } from '../../utils/ticks'

export const getTickYProps: TickLabelProps<number> = (value, index, values): Partial<TextProps> => ({
    ...leftTickLabelProps(),
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
})

export const getTickXProps: TickLabelProps<Date> = (value, index, values): Partial<TextProps> => ({
    ...bottomTickLabelProps(),
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${formatXLabel(value)}`,
})

/**
 * Tick component displays tick label for each axis line of chart.
 */
export const Tick: React.FunctionComponent<TickRendererProps> = props => {
    const { formattedValue, ...tickLabelProps } = props

    // Hack with Group + Text (aria hidden)
    // Because the Text component renders text inside of svg element and text element with tspan
    // this makes another nested group for a11y tree. To avoid "group - end group"
    // phrase in voice over we hide nested children from a11y tree and put explicit aria-label
    // on the parent Group element with role text
    return (
        // eslint-disable-next-line jsx-a11y/aria-role
        <Group role="text" aria-label={tickLabelProps['aria-label']}>
            <Text aria-hidden={true} {...tickLabelProps}>
                {formattedValue}
            </Text>
        </Group>
    )
}
