import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { isErrorLike, ErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import H from 'history'
import { ListHeaderQueryLinksNav } from '../../../../components/listHeaderQueryLinks/ListHeaderQueryLinks'

interface Props extends Pick<QueryParameterProps, 'query'> {
    changesets: undefined | GQL.IChangesetConnection | ErrorLike
    location: H.Location
}

/**
 * A list of changeset states with counts for the changeset list header.
 */
export const ChangesetListHeaderStates: React.FunctionComponent<Props> = ({ changesets, ...props }) =>
    changesets !== undefined && !isErrorLike(changesets) ? (
        <ListHeaderQueryLinksNav
            {...props}
            links={[
                {
                    label: 'unpublished',
                    queryField: 'is',
                    queryValues: ['unpublished'],
                    count: 123, // TODO(sqs)
                },
                {
                    label: 'open',
                    queryField: 'is',
                    queryValues: ['open'],
                    count: changesets.filters.openCount,
                },
                {
                    label: 'merged',
                    queryField: 'is',
                    queryValues: ['merged'],
                    count: 123, // TODO(sqs)
                },
                {
                    label: 'closed',
                    queryField: 'is',
                    queryValues: ['closed'],
                    count: changesets.filters.closedCount,
                },
            ]}
            className="flex-1 nav"
        />
    ) : null
