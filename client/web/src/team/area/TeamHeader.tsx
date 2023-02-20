import React from 'react'

import { NavLink } from 'react-router-dom'

import { Alert, Badge, PageHeader } from '@sourcegraph/wildcard'

import { TeamAreaRouteContext } from './TeamArea'

interface Props extends Pick<TeamAreaRouteContext, 'team'> {
    className?: string
}

/**
 * Header for the team area.
 */
export const TeamHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ team, className = '' }) => {
    const url = team.url

    return (
        <div className={className}>
            <div className="container">
                {team && (
                    <>
                        <PageHeader
                            path={[
                                { to: '/teams', text: 'Teams' },
                                {
                                    text: (
                                        <>
                                            {team.displayName ? (
                                                <>
                                                    {team.displayName} ({team.name})
                                                </>
                                            ) : (
                                                team.name
                                            )}
                                        </>
                                    ),
                                },
                            ]}
                            className="mb-3"
                        />

                        {team.readonly && (
                            <Alert variant="info" className="mb-3">
                                This team is managed externally and can not be modified from the UI except by
                                site-admins.
                            </Alert>
                        )}

                        <nav className="d-flex align-items-end justify-content-between" aria-label="Org">
                            <ul className="nav nav-tabs w-100">
                                <li className="nav-item">
                                    <NavLink to={url} className="nav-link" activeClassName="active" exact={true}>
                                        <span>
                                            <span className="text-content" data-tab-content="Profile">
                                                Profile
                                            </span>
                                        </span>
                                    </NavLink>
                                </li>
                                <li className="nav-item">
                                    <NavLink
                                        to={`${url}/members`}
                                        className="nav-link"
                                        activeClassName="active"
                                        exact={true}
                                    >
                                        <span>
                                            <span className="text-content" data-tab-content="Members">
                                                <span>
                                                    Members <Badge pill={true}>{team.members.totalCount}</Badge>
                                                </span>
                                            </span>
                                        </span>
                                    </NavLink>
                                </li>
                                <li className="nav-item">
                                    <NavLink
                                        to={`${url}/child-teams`}
                                        className="nav-link"
                                        activeClassName="active"
                                        exact={true}
                                    >
                                        <span>
                                            <span className="text-content" data-tab-content="Child teams">
                                                <span>
                                                    Child teams <Badge pill={true}>{team.childTeams.totalCount}</Badge>
                                                </span>
                                            </span>
                                        </span>
                                    </NavLink>
                                </li>
                            </ul>
                        </nav>
                    </>
                )}
            </div>
        </div>
    )
}
