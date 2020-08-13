import * as H from 'history'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { PageHeader } from '../components/PageHeader'
import { eventLogger } from '../tracking/eventLogger'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsList } from './ExtensionsList'

interface Props extends ExtensionsAreaRouteContext, SettingsCascadeProps {
    location: H.Location
    history: H.History
}

/** A page that displays overview information about the available extensions. */
export const ExtensionsOverviewPage: React.FunctionComponent<Props> = props => {
    useEffect(() => {
        eventLogger.logViewEvent('ExtensionsOverview')
    }, [])
    const headerActionsContext = { ...props, isPrimaryHeader: true }
    const headerActions = (
        <>
            {props.extensionsAreaHeaderActionButtons.map(
                ({ condition = () => true, to, icon: Icon, label }) =>
                    condition(headerActionsContext) && (
                        <Link className="btn ml-2 btn-secondary" to={to(headerActionsContext)} key={label}>
                            {Icon && <Icon className="icon-inline" />} {label}
                        </Link>
                    )
            )}
        </>
    )
    return (
        <>
            <PageHeader
                title="Extensions"
                icon={<PuzzleOutlineIcon />}
                breadcrumbs={[{ key: 'extensions', element: 'Extensions' }]}
                actions={headerActions}
            />
            <div className="container">
                <div className="py-3">
                    {!props.authenticatedUser && (
                        <div className="alert alert-info">
                            <Link to="/sign-in" className="btn btn-primary mr-2">
                                Sign in to add and configure extensions
                            </Link>
                            <small>An account is required.</small>
                        </div>
                    )}
                    <ExtensionsList {...props} subject={props.subject} settingsCascade={props.settingsCascade} />
                </div>
            </div>
        </>
    )
}
