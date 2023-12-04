import { FC } from 'react'

import classNames from 'classnames'

import { Position } from '@sourcegraph/extension-api-classes'
import { useQuery } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Code, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import {
    HighlightResponseFormat,
    ReferencesPanelHighlightedBlobResult,
    ReferencesPanelHighlightedBlobVariables,
} from '../graphql-operations'
import { SearchPanelConfig } from '../repo/blob/codemirror/search'
import { CodeMirrorBlob } from '../repo/blob/CodeMirrorBlob'

import { FETCH_HIGHLIGHTED_BLOB } from './ReferencesPanelQueries'

import styles from './ReferencesPanel.module.scss'

export interface SideBlobProps
    extends TelemetryProps,
        SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps {
    repository: string
    commitID: string
    file: string
    activeURL?: string
    position?: Position
    blobNav?: (url: string) => void
    wrapLines?: boolean
    codeIntelAndSgExtensions?: boolean
    navigateToLineOnAnyClick?: boolean
    searchPanelConfig?: SearchPanelConfig
    className?: string
}

export const SideBlob: FC<SideBlobProps> = props => {
    const {
        activeURL,
        repository,
        commitID,
        file,
        blobNav,

        // SideBlob currently only used in ReferencePanel, where code intel and sg-extensions are permanently disabled,
        // and clicking on any line navigates to that line in the main BlobPage view.
        codeIntelAndSgExtensions = false,
        navigateToLineOnAnyClick = true,
        wrapLines = true,

        searchPanelConfig,
        extensionsController,
        settingsCascade,
        telemetryService,
        platformContext,
        className,
    } = props

    const { data, error, loading } = useQuery<
        ReferencesPanelHighlightedBlobResult,
        ReferencesPanelHighlightedBlobVariables
    >(FETCH_HIGHLIGHTED_BLOB, {
        variables: {
            repository: props.repository,
            commit: props.commitID,
            path: props.file,
            format: HighlightResponseFormat.JSON_SCIP,
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'network-only',
    })

    // If we're loading and haven't received any data yet
    if (loading && !data) {
        return (
            <>
                <LoadingSpinner inline={false} className="mx-auto my-4" />
                <Text alignment="center" className="text-muted">
                    <i>
                        Loading <Code>{props.file}</Code>...
                    </i>
                </Text>
            </>
        )
    }

    // If we received an error before we had received any data
    if (error && !data) {
        return (
            <div>
                <Text className="text-danger">
                    Loading <Code>{props.file}</Code> failed:
                </Text>
                <pre>{error.message}</pre>
            </div>
        )
    }

    // If there weren't any errors and we just didn't receive any data
    if (!data?.repository?.commit?.blob?.highlight) {
        return <>Nothing found</>
    }

    const { lsif } = data?.repository?.commit?.blob?.highlight

    // TODO: display a helpful message if syntax highlighting aborted, see https://github.com/sourcegraph/sourcegraph/issues/40841

    return (
        <CodeMirrorBlob
            activeURL={activeURL}
            nav={blobNav}
            wrapCode={wrapLines}
            codeIntelAndSgExtensions={codeIntelAndSgExtensions}
            navigateToLineOnAnyClick={navigateToLineOnAnyClick}
            blobInfo={{
                lsif: lsif ?? '',
                commitID,
                filePath: file,
                repoName: repository,
                revision: commitID,
                content: data?.repository?.commit?.blob?.content ?? '',
                mode: 'lspmode',
            }}
            searchPanelConfig={searchPanelConfig}
            className={classNames(className, styles.sideBlobCode)}
            platformContext={platformContext}
            extensionsController={extensionsController}
            settingsCascade={settingsCascade}
            telemetryService={telemetryService}
        />
    )
}
