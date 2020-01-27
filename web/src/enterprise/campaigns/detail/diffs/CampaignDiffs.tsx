import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs, Connection } from '../../../../components/FilteredConnection'
import { FileDiffTabNodeProps, FileDiffTabNode } from '../FileDiffTabNode'
import { Observable, Subject } from 'rxjs'
import { DEFAULT_CHANGESET_LIST_COUNT } from '../presentation'

interface Props extends ThemeProps {
    queryChangesetsConnection: (
        args: FilteredConnectionQueryArgs
    ) => Observable<Connection<GQL.IExternalChangeset | GQL.IChangesetPlan>>
    persistLines: boolean
    history: H.History
    location: H.Location
    changesetUpdates: Subject<void>
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
    changesetUpdates,
}) => (
    <div className={className}>
        <FilteredConnection<GQL.IExternalChangeset | GQL.IChangesetPlan, Omit<FileDiffTabNodeProps, 'node'>>
            className="mt-2"
            updates={changesetUpdates}
            nodeComponent={FileDiffTabNode}
            nodeComponentProps={{
                persistLines,
                isLightTheme,
                history,
                location,
            }}
            queryConnection={queryChangesetsConnection}
            hideSearch={true}
            defaultFirst={DEFAULT_CHANGESET_LIST_COUNT}
            noun="changeset"
            pluralNoun="changesets"
            history={history}
            location={location}
            useURLQuery={false}
        />
    </div>
)
