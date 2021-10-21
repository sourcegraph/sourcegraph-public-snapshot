import React from 'react'

import styles from './SearchContextMenu.module.scss'

const NAMESPACED_SEARCH_CONTEXT_SPEC_REGEX = /@(.*?)\/(.*)/

interface ParsedSearchContextSpec {
    namespaceName?: string
    searchContextName?: string
}

function parseSearchContextSpec(searchContextSpec: string): ParsedSearchContextSpec {
    const namespacedMatch = searchContextSpec.match(NAMESPACED_SEARCH_CONTEXT_SPEC_REGEX)
    if (namespacedMatch) {
        return { namespaceName: namespacedMatch[1], searchContextName: namespacedMatch[2] }
    }
    if (searchContextSpec.startsWith('@')) {
        return { namespaceName: searchContextSpec.slice(1) }
    }
    return { searchContextName: searchContextSpec }
}

function highlightText(text: string, highlightPart: string, highlightClass: string): JSX.Element {
    const index = text.toLowerCase().indexOf(highlightPart.toLowerCase())
    if (index === -1) {
        return <>{text}</>
    }
    const before = text.slice(0, index)
    const highlighted = text.slice(index, index + highlightPart.length)
    const after = text.slice(index + highlightPart.length)
    return (
        <>
            {before}
            <strong className={highlightClass}>{highlighted}</strong>
            {after}
        </>
    )
}

function highlightSearchContextSpecPart(specPart?: string, highlightPart?: string): JSX.Element | string | undefined {
    return specPart && highlightPart ? <>{highlightText(specPart, highlightPart, styles.itemHighlighted)}</> : specPart
}

export const HighlightedSearchContextSpec: React.FunctionComponent<{ spec: string; searchFilter: string }> = ({
    spec,
    searchFilter,
}) => {
    if (searchFilter.length === 0) {
        return <>{spec}</>
    }
    // Parse search filter as a search context spec and
    // highlight both parts (namespace and name) separately
    const parsedSpec = parseSearchContextSpec(spec)
    const parsedSearchFilter = parseSearchContextSpec(searchFilter)

    const highlightedNamespaceName = highlightSearchContextSpecPart(
        parsedSpec.namespaceName,
        parsedSearchFilter.namespaceName
    )

    const highlightedSearchContextName = highlightSearchContextSpecPart(
        parsedSpec.searchContextName,
        parsedSearchFilter.searchContextName
    )

    return (
        <>
            {highlightedNamespaceName && <>@{highlightedNamespaceName}</>}
            {highlightedNamespaceName && highlightedSearchContextName ? (
                <>/{highlightedSearchContextName}</>
            ) : (
                <>{highlightedSearchContextName}</>
            )}
        </>
    )
}
