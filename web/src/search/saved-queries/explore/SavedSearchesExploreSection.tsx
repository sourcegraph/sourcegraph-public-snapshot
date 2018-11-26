import H from 'history'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { SavedQueries } from '../SavedQueries'

export const SavedSearchesExploreSection: React.FunctionComponent<
    {
        authenticatedUser: GQL.IUser | null
        location: H.Location
        isLightTheme: boolean
    } & SettingsCascadeProps
> = props => (
    <div className="saved-searches-explore-section">
        <h2>Saved searches</h2>
        <SavedQueries {...props} hideExampleSearches={true} hideTitle={true} settingsCascade={props.settingsCascade} />
    </div>
)
