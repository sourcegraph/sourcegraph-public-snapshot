import H from 'history'
import React from 'react'
import * as GQL from '../../../backend/graphqlschema'
import { SavedQueries } from '../SavedQueries'

export const SavedSearchesExploreSection: React.SFC<{
    authenticatedUser: GQL.IUser | null
    location: H.Location
    isLightTheme: boolean
}> = props => (
    <div className="saved-searches-explore-section">
        <h2>Saved searches</h2>
        <SavedQueries {...props} hideExampleSearches={true} hideTitle={true} />
    </div>
)
