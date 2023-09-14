import React from 'react'

import type { ChangesetFields } from '../../../../graphql-operations'
import type { queryExternalChangesetWithFileDiffs } from '../backend'

import { ExternalChangesetNode } from './ExternalChangesetNode'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'

import styles from './ChangesetNode.module.scss'

export interface ChangesetNodeProps {
    node: ChangesetFields
    viewerCanAdminister: boolean
    selectable?: {
        onSelect: (id: string) => void
        isSelected: (id: string) => boolean
    }
    /**
     * Element to precede the changeset so that it is separated from its neighbors when
     * viewed in a list, defaults to a full-width light gray horizontal rule
     */
    separator?: React.ReactNode
    /** For testing purposes. */
    queryExternalChangesetWithFileDiffs?: typeof queryExternalChangesetWithFileDiffs
    /** For testing purposes. */
    expandByDefault?: boolean
}

export const ChangesetNode: React.FunctionComponent<React.PropsWithChildren<ChangesetNodeProps>> = ({
    node,
    separator = <span className={styles.changesetNodeSeparator} />,
    ...props
}) => (
    <li className={styles.changesetNode}>
        {separator}
        {node.__typename === 'ExternalChangeset' ? (
            <ExternalChangesetNode node={node} {...props} />
        ) : (
            <HiddenExternalChangesetNode node={node} {...props} />
        )}
    </li>
)
