import React from 'react'

import * as H from 'history'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'

import { queryChangesetSpecFileDiffs } from './backend'
import { HiddenChangesetApplyPreviewNode } from './HiddenChangesetApplyPreviewNode'
import { VisibleChangesetApplyPreviewNode } from './VisibleChangesetApplyPreviewNode'

import styles from './ChangesetApplyPreviewNode.module.scss'

export interface ChangesetApplyPreviewNodeProps extends ThemeProps {
    node: ChangesetApplyPreviewFields
    history: H.History
    location: H.Location
    authenticatedUser: PreviewPageAuthenticatedUser
    selectable?: {
        onSelect: (id: string) => void
        isSelected: (id: string) => boolean
    }

    /** Used for testing. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

export const ChangesetApplyPreviewNode: React.FunctionComponent<
    React.PropsWithChildren<ChangesetApplyPreviewNodeProps>
> = ({ node, queryChangesetSpecFileDiffs, expandChangesetDescriptions, ...props }) => {
    if (node.__typename === 'HiddenChangesetApplyPreview') {
        return (
            <>
                <span className={styles.changesetApplyPreviewNodeSeparator} />
                <HiddenChangesetApplyPreviewNode node={node} />
            </>
        )
    }
    return (
        <>
            <span className={styles.changesetApplyPreviewNodeSeparator} />
            <VisibleChangesetApplyPreviewNode
                node={node}
                {...props}
                queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                expandChangesetDescriptions={expandChangesetDescriptions}
            />
        </>
    )
}
