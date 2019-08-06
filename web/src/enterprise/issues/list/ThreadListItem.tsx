import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadStateIcon } from '../../threadlike/threadState/ThreadStateIcon'

interface Props {
    changeset: Pick<GQL.IChangeset, '__typename' | 'number' | 'title' | 'state' | 'url'>
}

/**
 * An item in the list of changesets.
 */
export const ChangesetListItem: React.FunctionComponent<Props> = ({ changeset }) => (
    <Link to={changeset.url} className="d-flex align-items-center text-decoration-none">
        <ThreadStateIcon thread={changeset} className="mr-2" />
        <span className="text-muted mr-2">#{changeset.number}</span> {changeset.title}
    </Link>
)
