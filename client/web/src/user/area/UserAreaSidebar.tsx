import SettingsIcon from 'mdi-react/SettingsIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { orgURL } from '../../org'
import { OrgAvatar } from '../../org/OrgAvatar'
import { UserAvatar } from '../UserAvatar'
import { UserAreaRouteContext } from './UserArea'

interface Props extends UserAreaRouteContext {
    size: 'small' | 'large'
    className?: string
}

/**
 * Sidebar for the user area.
 */
export const UserAreaSidebar: React.FunctionComponent<Props> = ({ url, size, className = '', ...props }) => (
    <div className={`user-area-sidebar d-flex flex-column ${className}`}>
        {size === 'large' ? (
            <>
                {props.user.avatarURL && (
                    <UserAvatar className="user-area-sidebar__avatar--large align-self-center mb-2" user={props.user} />
                )}
                {props.user.displayName ? (
                    <header>
                        <h1 className="test-user-area-sidebar__display-name h2 font-weight-bold">
                            {props.user.displayName}{' '}
                        </h1>
                        <h2 className="font-weight-normal h4 text-muted">{props.user.username}</h2>
                    </header>
                ) : (
                    <h1 className="h2 font-weight-bold">props.user.username</h1>
                )}
                {props.user.viewerCanAdminister && (
                    <Link
                        to={props.user.settingsURL!}
                        className="btn btn-link align-self-start pl-0 d-flex align-items-center"
                    >
                        <SettingsIcon className="mr-2" /> Settings{' '}
                    </Link>
                )}

                {props.user.organizations.nodes.length > 0 && (
                    <div className="mt-3 pt-3 border-top">
                        <h3>Organizations</h3>
                        {props.user.organizations.nodes.map(org => (
                            <Link
                                className="d-flex align-items-center"
                                key={org.id}
                                to={orgURL(org.name)}
                                data-tooltip={org.displayName || org.name}
                            >
                                <OrgAvatar org={org.name} className="d-inline-flex mr-2" /> {org.name}
                            </Link>
                        ))}
                    </div>
                )}
            </>
        ) : (
            <div className="d-flex align-items-center" style={{ position: 'relative' }}>
                {props.user.avatarURL && (
                    <UserAvatar className="user-area-sidebar__avatar--small align-self-center mr-2" user={props.user} />
                )}{' '}
                <h1 className="h5 font-weight-normal mb-0">
                    <Link to={props.user.url} className="stretched-link">
                        {props.user.username}
                    </Link>
                </h1>
            </div>
        )}
    </div>
)
