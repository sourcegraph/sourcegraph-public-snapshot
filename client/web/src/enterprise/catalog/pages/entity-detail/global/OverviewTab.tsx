import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'
import { formatPersonName, PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { ComponentIcon } from '../../../components/ComponentIcon'
import { EntityOwner } from '../../../components/entity-owner/EntityOwner'

import { CatalogExplorer } from './CatalogExplorer'
import { ComponentInsights } from './ComponentInsights'
import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { OverviewStatusContexts } from './OverviewStatusContexts'
import { whoKnowsDescription } from './WhoKnowsTab'

interface Props extends TelemetryProps, SettingsCascadeProps, PlatformContextProps {
    entity: ComponentStateDetailFields
    className?: string
}

export const OverviewTab: React.FunctionComponent<Props> = ({
    entity,
    className,
    telemetryService,
    settingsCascade,
    platformContext,
}) => (
    <div className={classNames('row no-gutters', className)}>
        <div className="col-md-4 col-lg-3 col-xl-2 border-right p-3">
            {entity.name && (
                <h2 className="d-flex align-items-center mb-1">
                    <ComponentIcon component={entity} className="icon-inline mr-2" /> {entity.name}
                </h2>
            )}
            <div className="text-muted small mb-2">
                {entity.__typename === 'Component' && `${entity.kind[0]}${entity.kind.slice(1).toLowerCase()}`}
            </div>
            {entity.description && <p className="mb-3">{entity.description}</p>}
            <div>
                <Link
                    to={`/search?q=context:c/${entity.name}`}
                    className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                >
                    <SearchIcon className="icon-inline mr-1" /> Search...
                </Link>
                {entity.readme && (
                    <Link to={entity.readme.url} className="d-flex align-items-center text-body mb-3 mr-2">
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
                        <EntityOwner owner={entity.owner} className="d-block" />
                        <ul className="list-unstyled d-flex flex-wrap">
                            {entity.owner?.__typename === 'Group' &&
                                entity.owner.members.map(member => (
                                    <li key={member.email} className="mr-1 mb-1">
                                        <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                                            <UserAvatar user={member} size={18} />
                                        </LinkOrSpan>
                                    </li>
                                ))}
                        </ul>
                    </dd>
                    <dt>Lifecycle</dt>
                    <dd>{entity.lifecycle.toLowerCase()}</dd>
                </dl>
                <Link
                    to={`${entity.url}/who-knows`}
                    className="btn btn-outline-secondary mb-3"
                    data-tooltip={whoKnowsDescription(entity)}
                >
                    Who knows...?
                </Link>
            </div>
        </div>
        <div className="col-md-8 col-lg-9 col-xl-10 p-3">
            <div className="card mb-3">
                <ComponentSourceDefinitions component={entity} listGroupClassName="list-group-flush" />
                {entity.commits?.nodes[0] && <LastCommit commit={entity.commits?.nodes[0]} className="card-footer" />}
            </div>
            <OverviewStatusContexts entity={entity} itemClassName="mb-3" />
            <CatalogExplorer entity={entity.id} className="mb-3" />
            {false && (
                <ComponentInsights
                    entity={entity.id}
                    className="mb-3"
                    telemetryService={telemetryService}
                    settingsCascade={settingsCascade}
                    platformContext={platformContext}
                />
            )}
            {entity.readme && (
                <div className="card mb-3">
                    <header className="card-header">
                        <h4 className="card-title mb-0 font-weight-bold">
                            <Link to={entity.readme.url} className="d-flex align-items-center text-body">
                                <FileDocumentIcon className="icon-inline mr-2" /> {entity.readme.name}
                            </Link>
                        </h4>
                    </header>
                    <Markdown dangerousInnerHTML={entity.readme.richHTML} className="card-body p-3" />
                </div>
            )}
        </div>
    </div>
)

const LastCommit: React.FunctionComponent<{
    commit: NonNullable<ComponentStateDetailFields['commits']>['nodes'][0]
    className?: string
}> = ({ commit, className }) => (
    <div className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={commit.author.person} size={18} />
        <PersonLink person={commit.author.person} className="font-weight-bold mr-2 flex-shrink-0" />
        <Link to={commit.url} className="text-truncate flex-grow-1 text-body mr-2" title={commit.message}>
            {commit.subject}
        </Link>
        <small className="text-nowrap text-muted">
            <Link to={commit.url} className="text-monospace text-muted mr-2 d-none d-md-inline">
                {commit.abbreviatedOID}
            </Link>
            <Timestamp date={commit.author.date} noAbout={true} />
        </small>
    </div>
)
