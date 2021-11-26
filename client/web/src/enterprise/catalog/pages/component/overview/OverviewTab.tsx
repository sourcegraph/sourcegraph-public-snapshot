import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { useLocation, useRouteMatch } from 'react-router'
import { Link, Route, Switch } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ErrorBoundary } from '../../../../../components/ErrorBoundary'
import {
    ComponentLabelsFields,
    ComponentStateDetailFields,
    ComponentTagsFields,
} from '../../../../../graphql-operations'
import { formatPersonName } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { ComponentAncestorsPath } from '../../../components/catalog-area-header/CatalogAreaHeader'
import { ComponentIcon } from '../../../components/ComponentIcon'
import { catalogPagePathForComponent } from '../ComponentDetailContent'
import { ComponentTag } from '../ComponentHeaderActions'
import styles from '../OverviewTab.module.scss'

import { ComponentOverviewMain } from './ComponentOverviewMain'
import { ComponentOverviewWhoKnows, whoKnowsDescription } from './ComponentOverviewWhoKnows'

interface Props
    extends TelemetryProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        ExtensionsControllerProps {
    component: ComponentStateDetailFields
    useHash?: boolean
    isTree?: boolean
    className?: string
}

export const OverviewTab: React.FunctionComponent<Props> = ({ component, useHash, isTree, className, ...props }) => {
    const match = useRouteMatch()
    const location = useLocation()

    return (
        <div className={classNames(styles.container, isTree ? 'pt-3' : 'p-3', className)}>
            <aside className={classNames(styles.aside)} style={{ display: 'none !important' }}>
                {!isTree && (
                    <>
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
                    </>
                )}
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
                            <ComponentOwnerSidebarItem component={component} />
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
                        to={`${component.catalogURL}/who-knows`}
                        className="btn btn-outline-secondary mb-3"
                        data-tooltip={whoKnowsDescription(component)}
                    >
                        Who knows...?
                    </Link>
                </div>
            </aside>
            <main className={classNames(styles.main)}>
                <ErrorBoundary location={location}>
                    <Switch
                        /* TODO(sqs): hack to make the router work with hashes */
                        location={useHash ? { ...location, pathname: location.pathname + location.hash } : undefined}
                    >
                        <Route path={match.url} exact={true}>
                            <ComponentOverviewMain {...props} component={component} />
                        </Route>
                        <Route path={`${match.url}${useHash ? '#' : '/'}who-knows`} exact={true}>
                            <ComponentOverviewWhoKnows
                                component={component}
                                noun={`the ${component.name} ${component.kind.toLowerCase()}`}
                            />
                        </Route>
                    </Switch>
                </ErrorBoundary>
            </main>
        </div>
    )
}

export const ComponentOwnerSidebarItem: React.FunctionComponent<{
    component: Pick<ComponentStateDetailFields, 'owner' | 'name' | '__typename' | 'kind' | 'url'>
    isTree?: boolean
}> = ({ component, isTree }) => (
    <>
        <ComponentAncestorsPath
            path={catalogPagePathForComponent(component).slice(1, -1)}
            divider=">"
            className="mb-1"
            componentClassName="d-block"
            lastComponentClassName="font-weight-bold"
        />
        {/* <ComponentOwner owner={component.owner} /> */}
        <ul className="list-unstyled d-flex flex-wrap mb-0">
            {component.owner?.__typename === 'Group' &&
                component.owner.members.map(member => (
                    <li key={member.email} className="mr-1 mb-1">
                        <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                            <UserAvatar user={member} size={19} />
                        </LinkOrSpan>
                    </li>
                ))}
        </ul>
    </>
)

export const ComponentLabelsSidebarItem: React.FunctionComponent<{
    component: ComponentLabelsFields
}> = ({ component }) =>
    component.labels.length > 0 ? (
        <dl>
            {component.labels.map(label => (
                <React.Fragment key={label.key}>
                    <dt>{label.key}</dt>
                    <dd>{label.values.join(', ')}</dd>
                </React.Fragment>
            ))}
        </dl>
    ) : null

export const ComponentTagsSidebarItem: React.FunctionComponent<{
    component: ComponentTagsFields
}> = ({ component: { tags } }) => (
    <>
        {tags.map(tag => (
            <ComponentTag
                key={tag.name}
                name={tag.name}
                components={tag.components.nodes}
                buttonClassName="p-1 border small text-muted"
            />
        ))}
    </>
)
