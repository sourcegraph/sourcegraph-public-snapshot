import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { CatalogExplorer } from '../CatalogExplorer'
import { ComponentInsights } from '../ComponentInsights'
import { ComponentSourceDefinitions } from '../ComponentSourceLocations'
import { OverviewStatusContexts } from '../OverviewStatusContexts'

interface Props extends TelemetryProps, SettingsCascadeProps, PlatformContextProps {
    component: ComponentStateDetailFields
}

export const ComponentOverviewMain: React.FunctionComponent<Props> = ({
    component,
    telemetryService,
    settingsCascade,
    platformContext,
}) => (
    <>
        <div className="card mb-3">
            <ComponentSourceDefinitions component={component} listGroupClassName="list-group-flush" />
            {component.commits?.nodes[0] && <LastCommit commit={component.commits?.nodes[0]} className="card-footer" />}
        </div>
        <OverviewStatusContexts component={component} itemClassName="mb-3" />
        <CatalogExplorer component={component.id} useURLForConnectionParams={false} className="mb-3" />
        {false && (
            <ComponentInsights
                component={component.id}
                className="mb-3"
                telemetryService={telemetryService}
                settingsCascade={settingsCascade}
                platformContext={platformContext}
            />
        )}
        {component.readme && (
            <div className="card mb-3">
                <header className="card-header">
                    <h4 className="card-title mb-0 font-weight-bold">
                        <Link to={component.readme.url} className="d-flex align-items-center text-body">
                            <FileDocumentIcon className="icon-inline mr-2" /> {component.readme.name}
                        </Link>
                    </h4>
                </header>
                <Markdown dangerousInnerHTML={component.readme.richHTML} className="card-body p-3" />
            </div>
        )}
    </>
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
