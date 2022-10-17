import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

export const ProductLicenseTags: React.FunctionComponent<
    React.PropsWithChildren<{
        tags: string[]
    }>
> = ({ tags }) => (
    <>
        {tags.map(tag => (
            <Badge variant="secondary" className="mr-1" key={tag} as="div">
                {tag}
            </Badge>
        ))}
    </>
)
