import React from 'react'
import { IChangesetLabel } from '../../../../../../shared/src/graphql/schema'
import classNames from 'classnames'
import { getBrightness } from 'color-brightness'

interface Props {
    label: IChangesetLabel
}

export const ChangesetLabel: React.FunctionComponent<Props> = ({ label }) => {
    let colorBrightness = 0
    if (label.color) {
        const [r, g, b] = [label.color.substr(0, 2), label.color.substr(2, 2), label.color.substr(4, 2)].map(value =>
            parseInt(value, 16)
        )
        colorBrightness = getBrightness(r, g, b)
    }
    return (
        <span
            className={classNames('badge mr-2 badge-secondary', colorBrightness < 127 && 'text-white')}
            // eslint-disable-next-line react/forbid-dom-props
            style={label.color && { backgroundColor: '#' + label.color }}
            data-tooltip={label.description}
        >
            {label.text}
        </span>
    )
}
