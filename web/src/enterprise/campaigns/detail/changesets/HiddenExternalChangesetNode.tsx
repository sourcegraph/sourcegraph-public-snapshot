import React from 'react'
import { HiddenExternalChangesetFields } from '../../../../graphql-operations'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { HiddenExternalChangesetInfoCell } from './HiddenExternalChangesetInfoCell'

export interface HiddenExternalChangesetNodeProps {
    node: Pick<
        HiddenExternalChangesetFields,
        'id' | 'nextSyncAt' | 'updatedAt' | 'externalState' | 'publicationState' | 'reconcilerState'
    >
}

export const HiddenExternalChangesetNode: React.FunctionComponent<HiddenExternalChangesetNodeProps> = ({ node }) => (
    <>
        <span />
        <ChangesetStatusCell changeset={node} />
        <HiddenExternalChangesetInfoCell node={node} />
        <span />
        <span />
        <span />
    </>
)
