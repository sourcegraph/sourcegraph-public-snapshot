/**
 * Forked component from @visx/text package.
 * Removed https://github.com/airbnb/visx/issues/1111 when will be resolved
 * */
import React, { ReactElement } from 'react'

import { useText, TextProps as OriginTextProps } from '@visx/text'

const SVG_STYLE = { overflow: 'visible' }

/**
 * Origin text props with changed innerRef is equal to ref of text element
 * origin value = ref of svg element. Because firefox has bug with measurements
 * nested svg element without sizes we have to measure sizes of text element instead.
 * */
export interface TextProps extends Omit<OriginTextProps, 'innerRef'> {
    /** Ref access to text element */
    innerRef?: React.Ref<SVGTextElement>
}

/**
 * Displays svg text element.
 * */
export function Text(props: TextProps): ReactElement {
    const {
        dx: xCoord = 0,
        dy: yCoord = 0,
        textAnchor = 'start',
        innerRef,
        verticalAnchor,
        angle,
        lineHeight = '1em',
        scaleToFit = false,
        capHeight,
        width,
        ...textProps
    } = props

    // eslint-disable-next-line id-length
    const { x = 0, fontSize } = textProps
    const { wordsByLines, startDy, transform } = useText(props as OriginTextProps)

    return (
        // eslint-disable-next-line react/forbid-dom-props
        <svg x={xCoord} y={yCoord} fontSize={fontSize} style={SVG_STYLE}>
            {wordsByLines.length > 0 ? (
                <text ref={innerRef} transform={transform} {...textProps} textAnchor={textAnchor}>
                    {wordsByLines.map((line, index) => (
                        <tspan key={index} x={x} dy={index === 0 ? startDy : lineHeight}>
                            {line.words.join(' ')}
                        </tspan>
                    ))}
                </text>
            ) : null}
        </svg>
    )
}
