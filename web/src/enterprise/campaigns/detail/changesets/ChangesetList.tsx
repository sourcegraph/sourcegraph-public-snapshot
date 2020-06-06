import React from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../../../shared/src/theme'
import {
    ConnectionListHeaderItems,
    ConnectionListHeader,
} from '../../../../components/connectionList/ConnectionListHeader'
import { isErrorLike, ErrorLike } from '../../../../../../shared/src/util/errors'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ChangesetNode } from './ChangesetNode'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    changesets: undefined | GQL.IChangesetConnection | ErrorLike
    campaign: Pick<GQL.ICampaign, 'viewerCanAdminister'>

    headerItems?: ConnectionListHeaderItems

    className?: string
    history: H.History
    location: H.Location
}

export const ChangesetList: React.FunctionComponent<Props> = ({
    changesets,
    campaign,
    headerItems,
    query,
    onQueryChange,
    className = '',
    ...props
}) => (
    <div className={className}>
        {isErrorLike(changesets) ? (
            <div className="alert alert-danger">{changesets.message}</div>
        ) : (
            <div className="card">
                <ConnectionListHeader {...props} items={headerItems} />
                {changesets === undefined ? (
                    <LoadingSpinner className="m-3" />
                ) : changesets.nodes.length === 0 ? (
                    <p className="p-3 mb-0 text-muted">No changesets found.</p>
                ) : (
                    <ul className="list-group list-group-flush">
                        {changesets.nodes.map(changeset => (
                            <ChangesetNode
                                {...props}
                                key={changeset.id}
                                node={changeset}
                                viewerCanAdminister={campaign.viewerCanAdminister}
                            />
                        ))}
                    </ul>
                )}
            </div>
        )}
    </div>
)
