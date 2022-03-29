import React from 'react'

import { RouteComponentProps } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import {
    ExternalLinkFields,
    RepositoryFields,
} from '../../graphql-operations'
import { RepositoryCommitPage } from '../commit/RepositoryCommitPage'

interface Props
    extends RouteComponentProps<{ revspec: string }>,
        TelemetryProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    repo: RepositoryFields
    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void
}

export const CommitsTab: React.FunctionComponent<Props> = ({
    ...props

}) => (
        <RepositoryCommitPage {...props} />
    )
