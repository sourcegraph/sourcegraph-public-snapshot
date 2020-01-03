import React from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FileDiffTab } from './FileDiffTab'

interface Props extends ThemeProps {
    changesets: Pick<GQL.IExternalChangesetConnection | GQL.IChangesetPlanConnection, 'nodes'>
    persistLines: boolean

    history: H.History
    location: H.Location

    className?: string
}

/**
 * A list of a campaign's or campaign preview's diffs.
 */
export const CampaignDiffs: React.FunctionComponent<Props> = ({
    changesets,
    persistLines,
    history,
    location,
    className = '',
    isLightTheme,
}) => (
    <div className={className}>
        <FileDiffTab
            nodes={changesets.nodes}
            persistLines={persistLines}
            history={history}
            location={location}
            isLightTheme={isLightTheme}
        ></FileDiffTab>
    </div>
)
