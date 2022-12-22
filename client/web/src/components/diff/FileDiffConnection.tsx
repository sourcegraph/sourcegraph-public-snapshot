import React from 'react'

import { FileDiffFields, Scalars } from '../../graphql-operations'
import { FilteredConnection } from '../FilteredConnection'

import { FileDiffNodeProps } from './FileDiffNode'

class FilteredFileDiffConnection extends FilteredConnection<FileDiffFields, NodeComponentProps> {}

type NodeComponentProps = Omit<FileDiffNodeProps, 'node'>

type FileDiffConnectionProps = FilteredFileDiffConnection['props']

export type PartInfo<ExtraData extends object = {}> = {
    repoName: string
    repoID: Scalars['ID']
    revision: string
    commitID: string
} & ExtraData

/**
 * Displays a list of file diffs.
 */
export const FileDiffConnection: React.FunctionComponent<React.PropsWithChildren<FileDiffConnectionProps>> = props => (
    <FilteredFileDiffConnection {...props} withCenteredSummary={true} />
)
