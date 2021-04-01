import React from 'react'

/** Wraps the children in a link if an href is passed. */
export const MaybeLink: React.FunctionComponent<React.AnchorHTMLAttributes<unknown>> = ({ children, ...props }) =>
    props.href
        ? <a {...props}>{children}</a>
        : (children as React.ReactElement)
