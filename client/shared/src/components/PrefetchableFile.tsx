import React, { useCallback, useEffect, useRef } from 'react'

import { Observable, Subscription } from 'rxjs'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

interface FilePrefetcherParams {
    revision: string
    filePath: string
    repoName: string
}

export type FilePrefetcher = (parameters: FilePrefetcherParams) => Observable<unknown | null>

interface PrefetchableFileProps extends FilePrefetcherParams {
    prefetch?: FilePrefetcher
    isPrefetchEnabled?: boolean
    isSelected?: boolean
}

/**
 * A wrapper component that supports pre-fetching file revisions on hover.
 * Note: This is currently experimental, and should only be enabled through
 * the `enableSidebarFilePrefetch ` feature flag.
 */
export const PrefetchableFile = React.forwardRef(function PrefetchableFile(props, reference) {
    const {
        prefetch,
        revision,
        filePath,
        repoName,
        isPrefetchEnabled,
        isSelected,
        as: Component = 'div',
        ...rest
    } = props

    const observable = useRef<Subscription | null>(null)

    const startPrefetch = useCallback(() => {
        if (observable.current || !prefetch) {
            // Already fetching/fetched OR prefetch not available
            return
        }

        // Note that we don't actually do anything with this data.
        // The primary aim is to kickstart the memoized observable so that
        // when BlobPage does try to fetch the data, it is already resolved/resolving.
        observable.current = prefetch({
            revision,
            filePath,
            repoName,
        }).subscribe()
    }, [prefetch, revision, filePath, repoName])

    const stopPrefetch = useCallback(() => {
        if (observable.current && !observable.current.closed) {
            // Cancel ongoing request and reset
            observable.current.unsubscribe()
            observable.current = null
        }
    }, [])

    // Support manually triggering prefetch with the `isSelected` prop.
    useEffect(() => {
        if (isPrefetchEnabled && isSelected) {
            startPrefetch()
        }
    }, [isSelected, isPrefetchEnabled, startPrefetch])

    return (
        <Component
            onMouseOver={isPrefetchEnabled ? startPrefetch : undefined}
            onMouseLeave={isPrefetchEnabled ? stopPrefetch : undefined}
            onFocus={isPrefetchEnabled ? startPrefetch : undefined}
            onBlur={isPrefetchEnabled ? stopPrefetch : undefined}
            ref={reference}
            {...rest}
        />
    )
}) as ForwardReferenceComponent<'div', PrefetchableFileProps>
