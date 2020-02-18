import React from 'react'
import { IChangesetLabel } from '../../../../../../shared/src/graphql/schema'
import classNames from 'classnames'

interface Props {
    label: IChangesetLabel
}

/**
 * Converts a hex color to a perceived brightness. Algorithm to determine color brightness taken from https://www.w3.org/TR/AERT/#color-contrast
 *
 * @param color The hex encoded RGB color without the #
 *
 * @returns The perceived brightness from 0 to 255
 */
export function colorBrightness(color: string): number {
    const [r, g, b] = [color.substr(0, 2), color.substr(2, 2), color.substr(4, 2)].map(value => parseInt(value, 16))
    return (r * 299 + g * 587 + b * 114) / 1000
}

export const ChangesetLabel: React.FunctionComponent<Props> = ({ label }) => {
    // We use this value to determine the label text color (dark or bright, depending on the colorBrightness of the label)
    const labelBrightness = colorBrightness(label.color)

    return (
        <span
            className={classNames('badge mr-2 badge-secondary', labelBrightness < 127 && 'text-white')}
            // eslint-disable-next-line react/forbid-dom-props
            style={{ backgroundColor: '#' + label.color }}
            data-tooltip={label.description}
        >
            {label.text}
        </span>
    )
}
