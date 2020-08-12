import * as H from 'history'
import React from 'react'
import { ChangesetSpecFields } from '../../../graphql-operations'
import { HiddenChangesetSpecNode } from './HiddenChangesetSpecNode'
import { VisibleChangesetSpecNode } from './VisibleChangesetSpecNode'
import { ThemeProps } from '../../../../../shared/src/theme'

export interface ChangesetSpecNodeProps extends ThemeProps {
    node: ChangesetSpecFields
    history: H.History
    location: H.Location
}

export const ChangesetSpecNode: React.FunctionComponent<ChangesetSpecNodeProps> = ({
    node,
    history,
    location,
    isLightTheme,
}) => {
    if (node.__typename === 'HiddenChangesetSpec') {
        return <HiddenChangesetSpecNode node={node} />
    }
    return <VisibleChangesetSpecNode node={node} history={history} location={location} isLightTheme={isLightTheme} />
}
