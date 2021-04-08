import React from 'react'
import { Link, LinkProps} from 'react-router-dom';
import * as H from 'history';

interface MaybeLinkProps<S = H.LocationState> extends Omit<LinkProps<S>, 'to'> {
    to?: H.LocationDescriptor<S> | ((location: H.Location<S>) => H.LocationDescriptor<S>);
}

/** Wraps the children in a link if to (link href) prop is passed. */
export const MaybeLink: React.FunctionComponent<MaybeLinkProps> = ({ children, to, ...props }) =>
    to ? <Link {...props} to={to}>{children}</Link> : (children as React.ReactElement)
