import React from 'react'

import { mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryRecorder, TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Badge, ButtonLink, H3, Icon, Link, LinkOrSpan, SourcegraphIcon, Text } from '@sourcegraph/wildcard'

import { PageRoutes } from '../../routes.constants'

import styles from './CodyManagementPage.module.scss'

interface UseCodyInEditorSectionProps extends TelemetryV2Props {}

export const CodyEditorsAndClients: React.FunctionComponent<UseCodyInEditorSectionProps> = ({ telemetryRecorder }) => (
    <div className={styles.responsiveContainer}>
        {EDITOR_INSTRUCTIONS.map(editor => (
            <EditorInstructions key={editor.name} editor={editor} telemetryRecorder={telemetryRecorder} />
        ))}
    </div>
)

const EDITOR_ICON_HEIGHT = 34

const EditorInstructions: React.FunctionComponent<
    { editor: EditorInstructionsTile; className?: string } & TelemetryV2Props
> = ({ editor, telemetryRecorder, className }) => (
    <div className={classNames('d-flex flex-column px-3', className)}>
        {/* eslint-disable-next-line react/forbid-dom-props */}
        <div className="d-flex my-3 align-items-center" style={{ minHeight: `${EDITOR_ICON_HEIGHT}px` }}>
            {editor.icon &&
                (typeof editor.icon === 'string' ? (
                    <img
                        alt={editor.name}
                        src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${editor.icon}.svg`}
                        width={EDITOR_ICON_HEIGHT}
                        height={EDITOR_ICON_HEIGHT}
                        className="mr-3"
                    />
                ) : (
                    <editor.icon className="mr-3" />
                ))}
            <H3 className="mb-0 font-weight-normal">{editor.name}</H3>
        </div>
        {editor.instructions && <editor.instructions telemetryRecorder={telemetryRecorder} />}
    </div>
)

interface EditorInstructionsTile {
    /** Refers to gs://sourcegraph-assets/ideIcons/ideIcon${icon}.svg. */
    icon?: string | React.ComponentType<{ className?: string }>

    name: string
    instructions?: React.FunctionComponent<{
        telemetryRecorder: TelemetryRecorder
    }>
}

const EDITOR_INSTRUCTIONS: EditorInstructionsTile[] = [
    {
        icon: 'VsCode',
        name: 'VS Code',
        instructions: ({ telemetryRecorder }) => (
            <div className="d-flex flex-column flex-gap-2 align-items-start">
                <ButtonLink
                    variant="primary"
                    to="vscode:extension/sourcegraph.cody-ai"
                    target="_blank"
                    rel="noopener"
                    className="mb-2"
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickInstall', {
                            metadata: { vscode: 1 },
                        })
                    }}
                >
                    Install Cody in VS Code
                </ButtonLink>
                <Link
                    to="https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai"
                    className="text-muted d-inline-flex align-items-center flex-gap-1"
                    target="_blank"
                    rel="noopener"
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickMarketplace', {
                            metadata: { vscode: 1 },
                        })
                    }}
                >
                    View in VS Code Marketplace{' '}
                    <Icon aria-label="Open in new window" role="img" svgPath={mdiOpenInNew} />
                </Link>
                <Link
                    to="https://github.com/sourcegraph/cody"
                    className="text-muted d-inline-flex align-items-center flex-gap-1"
                    target="_blank"
                    rel="noopener"
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickSource', {
                            metadata: { vscode: 1 },
                        })
                    }}
                >
                    Install from source <Icon aria-label="Open in new window" role="img" svgPath={mdiOpenInNew} />
                </Link>
            </div>
        ),
    },
    {
        icon: 'JetBrains',
        name: 'All JetBrains IDEs',
        instructions: ({ telemetryRecorder }) => (
            <div className="d-flex flex-column flex-gap-2 align-items-start">
                <ButtonLink
                    variant="primary"
                    to="https://plugins.jetbrains.com/plugin/9682-sourcegraph-cody--code-search"
                    className="mb-2"
                    target="_blank"
                    rel="noopener"
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickMarketplace', {
                            metadata: { jetbrains: 1 },
                        })
                    }}
                >
                    Install Cody from JetBrains&nbsp;Marketplace
                </ButtonLink>
                <Link
                    to="https://github.com/sourcegraph/jetbrains"
                    className="text-muted d-inline-flex align-items-center flex-gap-1"
                    target="_blank"
                    rel="noopener"
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickSource', {
                            metadata: { jetbrains: 1 },
                        })
                    }}
                >
                    Install from source <Icon aria-label="Open in new window" role="img" svgPath={mdiOpenInNew} />
                </Link>
                <Text className="text-muted small mt-2">
                    Works in IntelliJ, PyCharm, GoLand, Android Studio, WebStorm, Rider, RubyMine, and all other
                    JetBrains IDEs.
                </Text>
            </div>
        ),
    },
    {
        icon: ({ className }) => (
            <SourcegraphIcon className={className} width={EDITOR_ICON_HEIGHT} height={EDITOR_ICON_HEIGHT} />
        ),
        name: 'Web',
        instructions: ({ telemetryRecorder }) => (
            <div className="d-flex flex-column flex-gap-2 align-items-start">
                <ButtonLink
                    variant="primary"
                    to={PageRoutes.CodyChat}
                    onClick={() => {
                        telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickWebChat', {
                            metadata: { chrome: 1 },
                        })
                    }}
                >
                    Chat with Cody on the web
                </ButtonLink>
                <Text className="text-muted small mt-2">
                    ...or open the <strong>Cody</strong> sidebar when viewing a repository, directory, or code file on
                    Sourcegraph.
                </Text>
            </div>
        ),
    },
    {
        name: 'Other editors & clients',
        instructions: ({ telemetryRecorder }) => (
            <ul className="d-flex flex-column flex-gap-2 align-items-start list-unstyled">
                {OTHER_CLIENTS.map(client => (
                    <li key={client.name} className="d-flex flex-gap-2 align-items-center">
                        <LinkOrSpan
                            to={client.url}
                            target="_blank"
                            rel="noopener"
                            className={client.url ? undefined : 'text-muted'}
                            onClick={() => {
                                telemetryRecorder.recordEvent('cody.editorExtensionsInstructions', 'clickOther', {
                                    metadata: { [client.telemetryMetadataKey]: 1 },
                                })
                            }}
                        >
                            {client.name}
                        </LinkOrSpan>
                        {client.releaseStage && (
                            <Badge variant="outlineSecondary" small={true}>
                                {client.releaseStage}
                            </Badge>
                        )}
                    </li>
                ))}
            </ul>
        ),
    },
]

const OTHER_CLIENTS: {
    name: string
    url?: string
    telemetryMetadataKey: string
    releaseStage?: 'Experimental' | 'Coming soon'
}[] = [
    {
        name: 'Neovim',
        url: 'https://github.com/sourcegraph/sg.nvim#setup',
        telemetryMetadataKey: 'neovim',
        releaseStage: 'Experimental',
    },
    {
        name: 'Cody CLI',
        url: 'https://sourcegraph.com/github.com/sourcegraph/cody@main/-/blob/cli/README.md',
        telemetryMetadataKey: 'cli',
        releaseStage: 'Experimental',
    },
    {
        name: 'Visual Studio',
        telemetryMetadataKey: 'visualstudio',
        releaseStage: 'Coming soon',
    },
    {
        name: 'Eclipse',
        telemetryMetadataKey: 'eclipse',
        releaseStage: 'Coming soon',
    },
    {
        name: 'Emacs',
        url: 'https://github.com/sourcegraph/emacs-cody',
        telemetryMetadataKey: 'emacs',
        releaseStage: 'Coming soon',
    },
]
