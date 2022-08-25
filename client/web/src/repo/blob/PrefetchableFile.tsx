import React, { useCallback, useEffect, useRef } from 'react'

import { Subscription } from 'rxjs'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { HighlightResponseFormat } from '../../graphql-operations'

import { fetchBlob } from './backend'

interface PrefetchableFileProps {
    revision: string
    filePath: string
    repoName: string
    prefetch?: boolean
    isSelected: boolean
}

/**
 * A wrapper component that supports pre-fetching file revisions on hover.
 * Note: This is currently experimental, and should only be enabled through
 * the `enableSidebarFilePrefetch ` feature flag.
 */
export const PrefetchableFile = React.forwardRef(function PrefetchableFile(props, reference) {
    const { revision, filePath, repoName, prefetch, as: Component = 'div', isSelected, ...rest } = props

    const observable = useRef<Subscription | null>(null)

    const startPrefetch = useCallback(() => {
        if (observable.current) {
            // Already fetching or already fetched
            return
        }

        // Note that we don't actually do anything with this data.
        // The primary aim is to kickstart the memoized observable so that
        // when BlobPage does try to fetch the data, it is already resolved/resolving.
        observable.current = fetchBlob({
            commitID: revision,
            filePath,
            repoName,
            format: HighlightResponseFormat.HTML_PLAINTEXT,
        }).subscribe()
    }, [filePath, repoName, revision])

    const stopPrefetch = useCallback(() => {
        if (observable.current && !observable.current.closed) {
            // Cancel ongoing request and reset
            observable.current.unsubscribe()
            observable.current = null
        }
    }, [])

    // Start file prefetch if it's selected via keyboard navigation.
    useEffect(() => {
        if (isSelected) {
            startPrefetch()
        }
    }, [isSelected, startPrefetch])

    return (
        <Component
            onMouseOver={prefetch ? startPrefetch : undefined}
            onMouseLeave={prefetch ? stopPrefetch : undefined}
            ref={reference}
            {...rest}
        />
    )
}) as ForwardReferenceComponent<'div', PrefetchableFileProps>
