import type { FC, AnchorHTMLAttributes, HTMLAttributes } from 'react'

import { Link } from '../../../Link'

interface MaybeLinkProps extends AnchorHTMLAttributes<HTMLElement> {
    to?: string | void | null
}

/** Wraps the children in a link if to (link href) prop is passed. */
export const MaybeLink: FC<MaybeLinkProps> = props => {
    const { to, target, rel, ...attributes } = props

    return to ? (
        <Link {...attributes} to={to} target={target} rel={rel} />
    ) : (
        <g {...(attributes as HTMLAttributes<SVGGElement>)} />
    )
}
