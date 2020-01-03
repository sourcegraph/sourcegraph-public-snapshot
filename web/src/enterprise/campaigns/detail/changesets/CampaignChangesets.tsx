import React from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNode } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'

interface Props extends ThemeProps {
    changesets: Pick<GQL.IExternalChangesetConnection | GQL.IChangesetPlanConnection, 'nodes'>

    history: H.History
    location: H.Location

    className?: string
}

/**
 * A list of a campaign's or campaign preview's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    changesets,
    history,
    location,
    className = '',
    isLightTheme,
}) => (
    <div className={`list-group ${className}`}>
        {(changesets.nodes as (GQL.IExternalChangeset | GQL.IChangesetPlan)[]).map(changeset => (
            <ChangesetNode
                key={changeset.id}
                node={changeset}
                isLightTheme={isLightTheme}
                location={location}
                history={history}
            ></ChangesetNode>
        ))}
    </div>
)
