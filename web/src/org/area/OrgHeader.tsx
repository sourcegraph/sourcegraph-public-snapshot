import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import UserIcon from 'mdi-react/UserIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { NamespaceAreaHeaderLinks } from '../../namespaces/NamespaceAreaHeaderLinks'
import { OrgAvatar } from '../OrgAvatar'
import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    className?: string
}

/**
 * Header for the organization area.
 */
export const OrgHeader: React.FunctionComponent<Props> = ({ org, match, className = '' }) => (
    <div className={`org-header ${className}`}>
        <div className="container">
            {org && (
                <>
                    <h2 className="org-header__title">
                        <OrgAvatar org={org.name} />{' '}
                        <span className="org-header__title-text">{org.displayName || org.name}</span>
                    </h2>
                    <div className="d-flex align-items-end justify-content-between">
                        <ul className="nav nav-tabs border-bottom-0">
                            <li className="nav-item">
                                <NavLink to={`${match.url}`} exact={true} className="nav-link" activeClassName="active">
                                    Overview
                                </NavLink>
                            </li>
                            <li className="nav-item">
                                <NavLink
                                    to={`${match.url}/members`}
                                    exact={true}
                                    className="nav-link"
                                    activeClassName="active"
                                >
                                    <UserIcon className="icon-inline" /> Members
                                </NavLink>
                            </li>
                            <li className="nav-item">
                                <NavLink
                                    to={`${match.url}/searches`}
                                    exact={false}
                                    className="nav-link"
                                    activeClassName="active"
                                >
                                    <FeatureSearchOutlineIcon className="icon-inline" /> Saved searches
                                </NavLink>
                            </li>
                            {org.viewerCanAdminister && (
                                <li className="nav-item">
                                    <NavLink to={`${match.url}/settings`} className="nav-link" activeClassName="active">
                                        <SettingsIcon className="icon-inline" /> Settings
                                    </NavLink>
                                </li>
                            )}
                            <NamespaceAreaHeaderLinks url={match.url} />
                        </ul>
                        <div className="flex-1" />
                        {org.viewerPendingInvitation && org.viewerPendingInvitation.respondURL && (
                            <div className="pb-1">
                                <small className="mr-2">Join organization:</small>
                                <Link to={org.viewerPendingInvitation.respondURL} className="btn btn-success btn-sm">
                                    View invitation
                                </Link>
                            </div>
                        )}
                    </div>
                </>
            )}
        </div>
    </div>
)
