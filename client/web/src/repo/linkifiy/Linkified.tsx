import React, { useMemo, forwardRef } from 'react'

import { type ForwardReferenceComponent, Link } from '@sourcegraph/wildcard'

import type { ExternalServiceKind } from '../../graphql-operations'

import { getLinksFromString } from './get-links'

interface LinkifiedProps {
    input: string
    externalURLs: { url: string; serviceKind: ExternalServiceKind | null }[] | undefined
}

/**
 * Takes a given input string and transforms any matching URLs into <a> tags.
 */
export const Linkified = forwardRef((props, ref) => {
    const { input, externalURLs, as: Component = React.Fragment, ...otherProps } = props

    const elements = useMemo(() => {
        const result: React.ReactNode[] = []

        const links = getLinksFromString({ input, externalURLs })
        let lastIndex = 0

        for (const link of links) {
            const { start, end, href, value } = link
            if (start > lastIndex) {
                result.push(
                    <Component key={`${lastIndex}-${start}`} {...otherProps}>
                        {input.slice(lastIndex, start)}
                    </Component>
                )
            }
            result.push(
                <Link key={`${start}-${end}`} to={href} target="_blank" rel="noreferrer noopener">
                    {value}
                </Link>
            )
            lastIndex = end
        }

        if (lastIndex < input.length) {
            result.push(
                <Component key={`${lastIndex}-${input.length}`} {...otherProps}>
                    {input.slice(lastIndex)}
                </Component>
            )
        }

        return result
    }, [Component, externalURLs, input, otherProps])

    return <>{elements}</>
}) as ForwardReferenceComponent<React.ExoticComponent, LinkifiedProps>

Linkified.displayName = 'Linkified'
