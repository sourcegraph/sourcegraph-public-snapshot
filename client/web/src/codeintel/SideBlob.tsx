import { FC } from 'react'

import classNames from 'classnames'
import { fetchBlob } from 'src/repo/blob/backend'

import { Position } from '@sourcegraph/extension-api-classes'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Code, LoadingSpinner, Text, useObservableWithStatus } from '@sourcegraph/wildcard'

import { HighlightResponseFormat } from '../graphql-operations'
import { SearchPanelConfig } from '../repo/blob/codemirror/search'
import { Range } from '../repo/blob/codemirror/static-highlights'
import { CodeMirrorBlob } from '../repo/blob/CodeMirrorBlob'

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
    navigateToLineOnAnyClick?: boolean
    searchPanelConfig?: SearchPanelConfig
    className?: string
    staticHighlightRanges?: Range[]
}

export const SideBlob: FC<SideBlobProps> = props => {
    const {
        activeURL,
        repository,
        commitID,
        file,
        blobNav,
        wrapLines = true,
        navigateToLineOnAnyClick = true,
        searchPanelConfig,
        extensionsController,
        settingsCascade,
        telemetryService,
        platformContext,
        className,
        staticHighlightRanges,
    } = props

    const [blob, loading, error] = useObservableWithStatus(
        fetchBlob({
            repoName: props.repository,
            revision: props.commitID,
            filePath: props.file,
            format: HighlightResponseFormat.JSON_SCIP,
        })
    )

    // If we're loading and haven't received any data yet
    if (loading && !blob) {
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
    if (error && !blob) {
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
    if (!blob || !blob.highlight) {
        return <>Nothing found</>
    }

    // TODO: display a helpful message if syntax highlighting aborted, see https://github.com/sourcegraph/sourcegraph/issues/40841

    return (
        <CodeMirrorBlob
            activeURL={activeURL}
            nav={blobNav}
            wrapCode={wrapLines}
            navigateToLineOnAnyClick={navigateToLineOnAnyClick}
            blobInfo={{
                lsif: blob.highlight.lsif ?? '',
                commitID,
                filePath: file,
                repoName: repository,
                revision: commitID,
                content: blob.content,
                mode: 'lspmode',
                languages: blob.languages,
            }}
            searchPanelConfig={searchPanelConfig}
            className={classNames(className, styles.sideBlobCode)}
            platformContext={platformContext}
            extensionsController={extensionsController}
            settingsCascade={settingsCascade}
            telemetryService={telemetryService}
            staticHighlightRanges={staticHighlightRanges}
        />
    )
}
