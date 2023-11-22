import React from 'react'

import type { ChangesetFields } from '../../../graphql-operations'
import type { queryExternalChangesetWithFileDiffs } from '../detail/backend'

import { ExternalChangesetCloseNode } from './ExternalChangesetCloseNode'
import { HiddenExternalChangesetCloseNode } from './HiddenExternalChangesetCloseNode'

import styles from './ChangesetCloseNode.module.scss'

export interface ChangesetCloseNodeProps {
    node: ChangesetFields
    viewerCanAdminister: boolean
    queryExternalChangesetWithFileDiffs?: typeof queryExternalChangesetWithFileDiffs
    willClose: boolean
}

export const ChangesetCloseNode: React.FunctionComponent<React.PropsWithChildren<ChangesetCloseNodeProps>> = ({
    node,
    ...props
}) => (
    <li className={styles.changesetCloseNode}>
        <span className={styles.changesetCloseNodeSeparator} />
        {node.__typename === 'ExternalChangeset' ? (
            <ExternalChangesetCloseNode node={node} {...props} />
        ) : (
            <HiddenExternalChangesetCloseNode node={node} {...props} />
        )}
    </li>
)
