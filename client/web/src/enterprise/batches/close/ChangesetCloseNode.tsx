import React from 'react'

import * as H from 'history'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'

import { ChangesetFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs } from '../detail/backend'

import { ExternalChangesetCloseNode } from './ExternalChangesetCloseNode'
import { HiddenExternalChangesetCloseNode } from './HiddenExternalChangesetCloseNode'

import styles from './ChangesetCloseNode.module.scss'

export interface ChangesetCloseNodeProps extends ThemeProps {
    node: ChangesetFields
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    queryExternalChangesetWithFileDiffs?: typeof queryExternalChangesetWithFileDiffs
    willClose: boolean
}

export const ChangesetCloseNode: React.FunctionComponent<React.PropsWithChildren<ChangesetCloseNodeProps>> = ({
    node,
    ...props
}) => {
    if (node.__typename === 'ExternalChangeset') {
        return (
            <>
                <span className={styles.changesetCloseNodeSeparator} />
                <ExternalChangesetCloseNode node={node} {...props} />
            </>
        )
    }
    return (
        <>
            <span className={styles.changesetCloseNodeSeparator} />
            <HiddenExternalChangesetCloseNode node={node} {...props} />
        </>
    )
}
