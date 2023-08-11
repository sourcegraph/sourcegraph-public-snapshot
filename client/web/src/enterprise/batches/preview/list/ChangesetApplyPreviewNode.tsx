import React from 'react'

import type { ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import type { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'

import type { queryChangesetSpecFileDiffs } from './backend'
import { HiddenChangesetApplyPreviewNode } from './HiddenChangesetApplyPreviewNode'
import { VisibleChangesetApplyPreviewNode } from './VisibleChangesetApplyPreviewNode'

import styles from './ChangesetApplyPreviewNode.module.scss'

export interface ChangesetApplyPreviewNodeProps {
    node: ChangesetApplyPreviewFields
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
> = ({ node, queryChangesetSpecFileDiffs, expandChangesetDescriptions, ...props }) => (
    <li className={styles.changesetApplyPreviewNode}>
        <span className={styles.changesetApplyPreviewNodeSeparator} />
        {node.__typename === 'HiddenChangesetApplyPreview' ? (
            <HiddenChangesetApplyPreviewNode node={node} />
        ) : (
            <VisibleChangesetApplyPreviewNode
                node={node}
                {...props}
                queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                expandChangesetDescriptions={expandChangesetDescriptions}
            />
        )}
    </li>
)
