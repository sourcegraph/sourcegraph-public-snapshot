import SettingsIcon from 'mdi-react/SettingsIcon'
import TuneVerticalIcon from 'mdi-react/TuneVerticalIcon'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { orgURL } from '../../org'
import { OrgAvatar } from '../../org/OrgAvatar'
import { UserAvatar } from '../UserAvatar'
import { UserAreaPageProps } from './UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    className: string
}

/**
 * Header for the user area.
 */
export const UserHeader: React.SFC<Props> = (props: Props) => (
    <div className={`user-header area-header ${props.className}`}>
        <div className={`${props.className}-inner`}>
            {props.user && (
                <>
                    <h2 className="user-header__title">
                        {props.user.avatarURL && <UserAvatar className="user-header__avatar" user={props.user} />}
                        {props.user.displayName ? (
                            <>
                                {props.user.displayName}{' '}
                                <span className="user-header__title-subtitle">{props.user.username}</span>
                            </>
                        ) : (
                            props.user.username
                        )}
                    </h2>
                    <div className="area-header__nav">
                        <div className="area-header__nav-links">
                            <NavLink
                                to={`${props.match.url}`}
                                exact={true}
                                className="btn area-header__nav-link"
                                activeClassName="area-header__nav-link--active"
                            >
                                Overview
                            </NavLink>
                            {props.user.viewerCanAdminister && (
                                <NavLink
                                    to={`${props.match.url}/settings`}
                                    exact={true}
                                    className="btn area-header__nav-link"
                                    activeClassName="area-header__nav-link--active"
                                >
                                    <SettingsIcon className="icon-inline" /> Settings
                                </NavLink>
                            )}
                            {props.user.viewerCanAdminister && (
                                <NavLink
                                    to={`${props.match.url}/account`}
                                    className="btn area-header__nav-link"
                                    activeClassName="area-header__nav-link--active"
                                >
                                    <TuneVerticalIcon className="icon-inline" /> Account
                                </NavLink>
                            )}
                        </div>
                        {props.user.organizations.nodes.length > 0 && (
                            <div className="area-header__nav-actions">
                                <small className="area-header__nav-actions-label">Organizations</small>
                                {props.user.organizations.nodes.map(org => (
                                    <Link
                                        className="area-header__nav-action"
                                        key={org.id}
                                        to={orgURL(org.name)}
                                        data-tooltip={org.displayName || org.name}
                                    >
                                        <OrgAvatar org={org.name} />
                                    </Link>
                                ))}
                            </div>
                        )}
                    </div>
                </>
            )}
        </div>
    </div>
)
