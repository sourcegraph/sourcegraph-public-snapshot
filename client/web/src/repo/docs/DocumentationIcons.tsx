import * as React from 'react'

import { DocumentationIcon } from './DocumentationIcon'
import { Tag } from './graphql'

/**
 * Picks the tags to use for rendering icons for documentation. This includes any
 * private/deprecated/test/benchmark/example/license/owner tag, plus the first tag not matching
 * those (i.e. the first symbol tag.)
 *
 * @param tags of the documentation node
 */
function pickIconTags(tags: Tag[]): Tag[] {
    const genericTags = new Set(['private', 'deprecated', 'test', 'benchmark', 'example', 'license', 'owner'])
    const generics = tags.filter(tag => genericTags.has(tag))
    const symbols = tags.filter(tag => !genericTags.has(tag))
    if (symbols.length > 0) {
        generics.push(symbols[0])
        return generics
    }
    return generics
}

interface Props {
    tags: Tag[]
}

/**
 * Renders icons for the given documentation tags on a single node.
 */
export const DocumentationIcons: React.FunctionComponent<Props> = ({ tags }) => (
    <>
        {pickIconTags(tags).map(tag => (
            <DocumentationIcon key={tag} tag={tag} />
        ))}
    </>
)
