import React from 'react'
import { IChangesetLabel } from '../../../../../../shared/src/graphql/schema'
import classNames from 'classnames'

interface Props {
    label: IChangesetLabel
}

export const ChangesetLabel: React.FunctionComponent<Props> = ({ label }) => {
    const [r, g, b] = [label.color.substr(0, 2), label.color.substr(2, 2), label.color.substr(4, 2)].map(value =>
        parseInt(value, 16)
    )
    // Algorithm to determine color brightness taken from https://www.w3.org/TR/AERT/#color-contrast
    // We use this value to determine the label text color (dark or bright, depending on the colorBrightness of the label)
    const colorBrightness = (r * 299 + g * 587 + b * 114) / 1000

    return (
        <span
            className={classNames('badge mr-2 badge-secondary', colorBrightness < 127 && 'text-white')}
            // eslint-disable-next-line react/forbid-dom-props
            style={{ backgroundColor: '#' + label.color }}
            data-tooltip={label.description}
        >
            {label.text}
        </span>
    )
}
