import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { FileDiffTabNodeProps, FileDiffTabNode } from '../FileDiffTabNode'
import { Observable } from 'rxjs'
import classNames from 'classnames'
import { DEFAULT_LIST_COUNT } from '../presentation'

interface Props extends ThemeProps {
    queryChangesetsConnection: (
        args: FilteredConnectionQueryArgs
    ) => Observable<GQL.IExternalChangesetConnection | GQL.IChangesetPlanConnection>
    persistLines: boolean
    history: H.History
    location: H.Location
    className?: string
}

/**
 * A list of a campaign's or campaign preview's diffs.
 */
export const CampaignDiffs: React.FunctionComponent<Props> = ({
    queryChangesetsConnection,
    persistLines = true,
    isLightTheme,
    history,
    location,
    className,
}) => (
    <FilteredConnection<GQL.IExternalChangeset | GQL.IChangesetPlan, Omit<FileDiffTabNodeProps, 'node'>>
        className={classNames('mt-2', className)}
        // updates={changesetUpdates}
        nodeComponent={FileDiffTabNode}
        nodeComponentProps={{
            persistLines,
            isLightTheme,
            history,
            location,
        }}
        queryConnection={queryChangesetsConnection}
        hideSearch={true}
        defaultFirst={DEFAULT_LIST_COUNT}
        noun="changeset"
        pluralNoun="changesets"
        history={history}
        location={location}
        useURLQuery={false}
    />
)
