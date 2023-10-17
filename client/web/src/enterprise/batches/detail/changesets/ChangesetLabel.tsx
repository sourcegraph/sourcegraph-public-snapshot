import React from 'react'

import classNames from 'classnames'

import { Badge } from '@sourcegraph/wildcard'

import type { ChangesetLabelFields } from '../../../../graphql-operations'

interface Props {
    label: ChangesetLabelFields
}

/**
 * Converts a hex color to a perceived brightness. Algorithm to determine color brightness
 * taken from https://www.w3.org/TR/AERT/#color-contrast
 *
 * @param color The hex encoded RGB color without the #
 * @returns The perceived brightness from 0 to 255
 */
export function colorBrightness(color: string): number {
    const [red, green, blue] = [color.slice(0, 2), color.slice(2, 4), color.slice(4, 6)].map(value =>
        parseInt(value, 16)
    )
    return (red * 299 + green * 587 + blue * 114) / 1000
}

export const ChangesetLabel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ label }) => {
    // We use this value to determine the label text color (dark or bright, depending on the colorBrightness of the label)
    const labelBrightness = colorBrightness(label.color)

    return (
        <Badge
            variant="secondary"
            className={classNames('mr-2', labelBrightness < 127 ? 'text-white' : 'text-dark')}
            style={{ backgroundColor: '#' + label.color }}
            tooltip={label.description || undefined}
            pill={true}
        >
            {label.text}
        </Badge>
    )
}
