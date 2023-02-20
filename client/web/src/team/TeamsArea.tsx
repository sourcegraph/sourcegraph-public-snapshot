import * as React from 'react'

import { Routes, Route } from 'react-router-dom-v5-compat'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { NotFoundPage } from '../components/HeroPage'

import type { TeamAreaProps } from './area/TeamArea'
import type { TeamListPageProps } from './list/TeamListPage'
import type { NewTeamPageProps } from './new/NewTeamPage'

const TeamArea = lazyComponent<TeamAreaProps, 'TeamArea'>(() => import('./area/TeamArea'), 'TeamArea')
const TeamListPage = lazyComponent<TeamListPageProps, 'TeamListPage'>(
    () => import('./list/TeamListPage'),
    'TeamListPage'
)
const NewTeamPage = lazyComponent<NewTeamPageProps, 'NewTeamPage'>(() => import('./new/NewTeamPage'), 'NewTeamPage')

export interface Props {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display team-related pages.
 */
const AuthenticatedTeamsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    // No teams on sourcegraph.com
    if (props.isSourcegraphDotCom) {
        return <NotFoundPage pageType="team" />
    }
    return (
        <>
            <Routes>
                <Route path="new" element={<NewTeamPage />} />
                <Route path="" element={<TeamListPage {...props} />} />
                <Route path=":teamName/*" element={<TeamArea {...props} />} />
                <Route element={<NotFoundPage pageType="team" />} />
            </Routes>
        </>
    )
}

export const TeamsArea = withAuthenticatedUser(AuthenticatedTeamsArea)
