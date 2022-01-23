import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { ComponentDetailFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { CatalogRelations } from '../CatalogRelations'
import { ComponentSourceLocations } from '../ComponentSourceLocations'

interface Props
    extends TelemetryProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ThemeProps,
        ExtensionsControllerProps {
    component: ComponentDetailFields
}

const hideOnTree = true

export const ComponentOverviewMain: React.FunctionComponent<Props> = ({
    component,
    telemetryService,
    settingsCascade,
    platformContext,
    ...props
}) => (
    <>
        {!hideOnTree && (
            <div className="card mb-3">
                {!hideOnTree && (
                    <ComponentSourceLocations component={component} listGroupClassName="list-group-flush" />
                )}
                {component.commits?.nodes[0] && (
                    <LastCommit
                        commit={component.commits?.nodes[0]}
                        className={hideOnTree ? 'card-body' : 'card-footer'}
                    />
                )}
            </div>
        )}
        <div className="card mb-3">
            {component.commits?.nodes[0] && (
                <LastCommit commit={component.commits?.nodes[0]} className="card-body border-bottom" />
            )}
            <ComponentSourceLocations {...props} component={component} className="card-body" />
        </div>
        <CatalogRelations component={component.id} useURLForConnectionParams={false} className="mb-3" />
        {component.readme && <ComponentReadme readme={component.readme} />}
    </>
)

export const LastCommit: React.FunctionComponent<{
    commit: NonNullable<ComponentDetailFields['commits']>['nodes'][0]
    after?: React.ReactFragment
    className?: string
}> = ({ commit, after, className }) => (
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
        {after}
    </div>
)

export const ComponentReadme: React.FunctionComponent<{
    readme: NonNullable<ComponentDetailFields['readme']>
}> = ({ readme }) => (
    <div className="card mb-3">
        <header className="card-header bg-transparent">
            <h4 className="card-title mb-0 font-weight-bold">
                <Link to={readme.url} className="d-flex align-items-center text-body">
                    <FileDocumentIcon className="icon-inline mr-2" /> {readme.name}
                </Link>
            </h4>
        </header>
        <Markdown dangerousInnerHTML={readme.richHTML} className="card-body p-3" />
    </div>
)
