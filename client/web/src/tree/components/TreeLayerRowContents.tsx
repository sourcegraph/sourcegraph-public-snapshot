import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'
import { Observable } from 'rxjs'

import { FetchBlobParameters } from '@sourcegraph/shared/src/backend/blob'
import { ResolvedRevision, ResolvedRevisionParameters } from '@sourcegraph/shared/src/backend/repo'
import { BlobFileFields } from '@sourcegraph/shared/src/graphql-operations'
import { LinkProps } from '@sourcegraph/wildcard'

import { FileLink } from '../../repo/FileLink'

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

type TreeLayerRowContentsLinkProps = LinkProps & {
    isNew?: boolean
    resolveRevision: (parameters: ResolvedRevisionParameters) => Observable<ResolvedRevision>
    fetchBlob: (parameters: FetchBlobParameters) => Observable<BlobFileFields | null>
}

export const TreeLayerRowContentsLink: React.FunctionComponent<
    React.PropsWithChildren<TreeLayerRowContentsLinkProps>
> = ({ className, children, isNew, ...rest }) => (
    <FileLink
        prefetch={true}
        className={classNames(styles.treeRowContents, className, isNew && styles.isNew)}
        {...rest}
    >
        {children}
    </FileLink>
)
