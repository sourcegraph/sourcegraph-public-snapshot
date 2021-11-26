import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { useLocation } from 'react-router'
import { Link, Route, Switch } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'
import { formatPersonName } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { ComponentIcon } from '../../../components/ComponentIcon'
import { EntityOwner } from '../../../components/entity-owner/EntityOwner'

import styles from '../OverviewTab.module.scss'
import { ComponentOverviewMain } from './ComponentOverviewMain'
import { ComponentOverviewWhoKnows, whoKnowsDescription } from './ComponentOverviewWhoKnows'

interface Props extends TelemetryProps, SettingsCascadeProps, PlatformContextProps {
    component: ComponentStateDetailFields
    className?: string
}

export const OverviewTab: React.FunctionComponent<Props> = ({
    component,
    className,
    telemetryService,
    settingsCascade,
    platformContext,
}) => {
    const location = useLocation()

    return (
        <div className={classNames(styles.container, 'p-3', className)}>
            <aside className={classNames(styles.aside)}>
                {component.name && (
                    <h2 className="mb-1">
                        <Link to={component.url} className="d-flex align-items-center text-body">
                            <ComponentIcon component={component} className="icon-inline mr-2 flex-shrink-0" />{' '}
                            {component.name}
                        </Link>
                    </h2>
                )}
                <div className="text-muted small mb-2">
                    {component.__typename === 'Component' &&
                        `${component.kind[0]}${component.kind.slice(1).toLowerCase()}`}
                </div>
                {component.description && <p className="mb-3">{component.description}</p>}
                <div>
                    <Link
                        to={`/search?q=context:c/${component.name}`}
                        className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                    >
                        <SearchIcon className="icon-inline mr-1" /> Search...
                    </Link>
                    {component.readme && (
                        <Link to={component.readme.url} className="d-flex align-items-center text-body mb-3 mr-2">
                            <FileDocumentIcon className="icon-inline mr-2" />
                            Documentation
                        </Link>
                    )}
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <FileAlertIcon className="icon-inline mr-2" />
                        Runbook
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <AlertCircleOutlineIcon className="icon-inline mr-2" />
                        Issues
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <SlackIcon className="icon-inline mr-2" />
                        #dev-frontend
                    </Link>
                    <hr className="my-3" />
                    <dl className="mb-3">
                        <dt>Owner</dt>
                        <dd>
                            <EntityOwner owner={component.owner} />
                            <ul className="list-unstyled d-flex flex-wrap">
                                {component.owner?.__typename === 'Group' &&
                                    component.owner.members.map(member => (
                                        <li key={member.email} className="mr-1 mb-1">
                                            <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                                                <UserAvatar user={member} size={18} />
                                            </LinkOrSpan>
                                        </li>
                                    ))}
                            </ul>
                        </dd>
                        <dt>Lifecycle</dt>
                        <dd>{component.lifecycle.toLowerCase()}</dd>
                        {component.labels.map(label => (
                            <React.Fragment key={label.key}>
                                <dt>{label.key}</dt>
                                <dd>{label.values.join(', ')}</dd>
                            </React.Fragment>
                        ))}
                    </dl>
                    <Link
                        to={`${component.url}/who-knows`}
                        className="btn btn-outline-secondary mb-3"
                        data-tooltip={whoKnowsDescription(component)}
                    >
                        Who knows...?
                    </Link>
                </div>
            </aside>
            <main className={classNames(styles.main)}>
                <ErrorBoundary location={location}>
                    <Switch>
                        <Route path={component.url} exact={true}>
                            <ComponentOverviewMain
                                component={component}
                                telemetryService={telemetryService}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                            />
                        </Route>
                        <Route path={`${component.url}/who-knows`} exact={true}>
                            <ComponentOverviewWhoKnows component={component} />
                        </Route>
                    </Switch>
                </ErrorBoundary>
            </main>
        </div>
    )
}
