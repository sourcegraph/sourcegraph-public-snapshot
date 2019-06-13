import React from 'react'
import * as GQL from '../../../shared/src/graphql/schema'
import { contrastingForegroundColor } from '../util/contrastingForegroundColor'

interface Props extends Pick<React.HTMLAttributes<HTMLElement>, 'className' | 'onClick'> {
    label: Pick<GQL.ILabel, 'name' | 'color'>

    tag?: 'span' | 'button'
}

/**
 * A label (corresponding to the GraphQL Label type).
 */
export const Label: React.FunctionComponent<Props> = ({
    label: { name, color },
    tag: Tag = 'span',
    className = '',
    children,
    ...props
}) => (
    <Tag
        {...props}
        className={`badge ${className}`}
        // tslint:disable-next-line: jsx-ban-props
        style={{ backgroundColor: color, color: contrastingForegroundColor(color) }}
    >
        {name}
        {children}
    </Tag>
)
