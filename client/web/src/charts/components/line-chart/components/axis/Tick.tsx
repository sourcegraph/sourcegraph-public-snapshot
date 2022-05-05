import React from 'react'

import { TickLabelProps, TickRendererProps } from '@visx/axis'
import { Group } from '@visx/group'
import { Text, TextProps } from '@visx/text'

import { formatXLabel } from '../../utils'

export const getTickYProps: TickLabelProps<number> = (value, index, values): Partial<TextProps> => ({
    dx: '-0.25em',
    dy: '0.25em',
    fill: '#222',
    fontFamily: 'Arial',
    fontSize: 10,
    textAnchor: 'end',
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${value}`,
})

export const getTickXProps: TickLabelProps<Date> = (value, index, values): Partial<TextProps> => ({
    dy: '0.25em',
    fill: '#222',
    fontFamily: 'Arial',
    fontSize: 10,
    textAnchor: 'middle',
    'aria-label': `Tick axis ${index + 1} of ${values.length}. Value: ${formatXLabel(value)}`,
})

/**
 * Tick component displays tick label for each axis line of chart.
 */
export const Tick: React.FunctionComponent<React.PropsWithChildren<TickRendererProps>> = props => {
    const { formattedValue, ...tickLabelProps } = props

    // Hack with Group + Text (aria hidden)
    // Because the Text component renders text inside svg element and text element with tspan
    // this makes another nested group for a11y tree. To avoid "group - end group"
    // phrase in voice over we hide nested children from a11y tree and put explicit aria-label
    // on the parent Group element with role text
    return (
        // eslint-disable-next-line jsx-a11y/aria-role
        <Group role="text" aria-label={tickLabelProps['aria-label']}>
            <Text aria-hidden={true} {...(tickLabelProps as TextProps)}>
                {formattedValue}
            </Text>
        </Group>
    )
}
