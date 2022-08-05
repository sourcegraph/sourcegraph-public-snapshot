import React, { HTMLAttributes, useCallback, useRef } from 'react'

import classNames from 'classnames'
import { Subscription } from 'rxjs'

import { Link, LinkProps } from '@sourcegraph/wildcard'

import { fetchBlob } from '../../repo/blob/backend'
import { parseBrowserRepoURL } from '../../util/url'

import styles from '../Tree.module.scss'

type TreeLayerRowContentsProps = HTMLAttributes<HTMLDivElement> & {
    isNew?: boolean
}

export const TreeLayerRowContents: React.FunctionComponent<React.PropsWithChildren<TreeLayerRowContentsProps>> = ({
    className,
    children,
    isNew,
    ...rest
}) => (
    <div className={classNames(styles.treeRowContents, className, isNew && styles.isNew)} {...rest}>
        {children}
    </div>
)

interface TreeLayerRowContentsLinkProps extends LinkProps {
    isNew?: boolean
    prefetch?: boolean
    commitID: string
}

export const TreeLayerRowContentsLink: React.FunctionComponent<
    React.PropsWithChildren<TreeLayerRowContentsLinkProps>
> = ({ to, commitID, className, children, isNew, prefetch, ...rest }) => {
    const observable = useRef<Subscription | null>(null)

    const startPrefetch = useCallback(() => {
        if (observable.current) {
            // Already fetching or already fetched
            return
        }

        const repo = parseBrowserRepoURL(to)
        if (repo?.filePath) {
            // Note that we don't actually do anything with this data.
            // The primary aim is to kickstart the memoized observable so that
            // when BlobPage does try to fetch the data, it is already resolved/resolving.
            observable.current = fetchBlob({
                commitID,
                filePath: repo.filePath,
                repoName: repo.repoName,
                formatOnly: true,
                // eslint-disable-next-line rxjs/no-ignored-subscription
            }).subscribe()
        }
    }, [commitID, to])

    const stopPrefetch = useCallback(() => {
        if (observable.current && !observable.current.closed) {
            // Cancel ongoing request and reset
            observable.current.unsubscribe()
            observable.current = null
        }
    }, [])

    return (
        <Link
            to={to}
            className={classNames(styles.treeRowContents, className, isNew && styles.isNew)}
            onMouseOver={prefetch ? startPrefetch : undefined}
            onMouseLeave={prefetch ? stopPrefetch : undefined}
            {...rest}
        >
            {children}
        </Link>
    )
}
