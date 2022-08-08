import React, { useCallback, useRef } from 'react'

import { Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { fetchBlob } from '../backend/blob'
import { resolveRevision } from '../backend/repo'
import { PlatformContextProps } from '../platform/context'

interface PrefetchableFileProps extends PlatformContextProps<'requestGraphQL'> {
    revision?: string
    filePath: string
    repoName: string
    prefetch?: boolean
}

export const PrefetchableFile = React.forwardRef(function PrefetchableFile(
    { revision, filePath, repoName, prefetch = true, platformContext, as: Component = 'div', ...rest },
    reference
) {
    const observable = useRef<Subscription | null>(null)

    const startPrefetch = useCallback(() => {
        if (observable.current) {
            // Already fetching or already fetched
            return
        }

        // Note that we don't actually do anything with this data.
        // The primary aim is to kickstart the memoized observable so that
        // when BlobPage does try to fetch the data, it is already resolved/resolving.
        observable.current = resolveRevision({
            revision,
            repoName,
            requestGraphQL: platformContext.requestGraphQL,
        })
            .pipe(
                switchMap(({ commitID }) =>
                    fetchBlob({
                        commitID,
                        filePath,
                        repoName,
                        formatOnly: true,
                        requestGraphQL: platformContext.requestGraphQL,
                    })
                )
            )
            .subscribe()
    }, [filePath, platformContext.requestGraphQL, repoName, revision])

    const stopPrefetch = useCallback(() => {
        if (observable.current && !observable.current.closed) {
            // Cancel ongoing request and reset
            observable.current.unsubscribe()
            observable.current = null
        }
    }, [])

    return (
        <Component
            onMouseOver={prefetch ? startPrefetch : undefined}
            onMouseLeave={prefetch ? stopPrefetch : undefined}
            ref={reference}
            {...rest}
        />
    )
}) as ForwardReferenceComponent<'div', PrefetchableFileProps>
