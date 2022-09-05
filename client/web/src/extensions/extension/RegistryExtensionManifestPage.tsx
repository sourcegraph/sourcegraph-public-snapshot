import * as React from 'react'

import { mdiInformationOutline, mdiEye } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'

import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import extensionSchemaJSON from '@sourcegraph/shared/src/schema/extension.schema.json'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Link, Alert, Icon, Code, H3, Tooltip } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { DynamicallyImportedMonacoSettingsEditor } from '../../settings/DynamicallyImportedMonacoSettingsEditor'
import { eventLogger } from '../../tracking/eventLogger'

import { ExtensionAreaRouteContext } from './ExtensionArea'

import styles from './RegistryExtensionManifestPage.module.scss'

export const ExtensionNoManifestAlert: React.FunctionComponent<{
    extension: ConfiguredRegistryExtension
}> = ({ extension }) => (
    <Alert variant="info">
        This extension is not yet published.
        {extension.registryExtension?.viewerCanAdminister && (
            <>
                <br />
                <Button
                    className="mt-3"
                    to={`${extension.registryExtension.url}/-/releases/new`}
                    variant="primary"
                    as={Link}
                >
                    Publish first release of extension
                </Button>
            </>
        )}
    </Alert>
)

interface Props extends ExtensionAreaRouteContext, RouteComponentProps<{}>, ThemeProps {}

interface State {
    viewMode: ViewMode
}

enum ViewMode {
    Rich = 'rich',
    Plain = 'plain',
}

/** A page that displays an extension's manifest. */
export class RegistryExtensionManifestPage extends React.PureComponent<Props, State> {
    private static STORAGE_KEY = 'RegistryExtensionManifestPage.viewMode'
    private static getViewMode(): ViewMode {
        const storedViewMode = localStorage.getItem(RegistryExtensionManifestPage.STORAGE_KEY)
        if (storedViewMode === ViewMode.Rich || storedViewMode === ViewMode.Plain) {
            return storedViewMode
        }
        return ViewMode.Rich
    }
    private static setViewMode(value: ViewMode): void {
        localStorage.setItem(RegistryExtensionManifestPage.STORAGE_KEY, value)
    }

    public state: State = { viewMode: RegistryExtensionManifestPage.getViewMode() }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionManifest')
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <PageTitle title={`Manifest of ${this.props.extension.id}`} />
                <div className="d-flex align-items-center justify-content-between">
                    <div className="d-flex align-items-center">
                        <H3 className="mb-0 mr-1">Manifest</H3>
                        <Tooltip content="The published JSON description of how to run or access the extension">
                            <Icon
                                className="text-muted"
                                aria-label="The published JSON description of how to run or access the extension"
                                svgPath={mdiInformationOutline}
                            />
                        </Tooltip>
                    </div>
                    <div>
                        {this.props.extension.manifest && (
                            <Button onClick={this.onViewModeButtonClick} variant="secondary">
                                <Icon aria-hidden={true} svgPath={mdiEye} /> Use{' '}
                                {this.state.viewMode === ViewMode.Plain ? ViewMode.Rich : ViewMode.Plain} viewer
                            </Button>
                        )}{' '}
                        {this.props.extension.registryExtension?.viewerCanAdminister && (
                            <Button
                                to={`${this.props.extension.registryExtension.url}/-/releases/new`}
                                variant="primary"
                                as={Link}
                            >
                                Publish new release
                            </Button>
                        )}
                    </div>
                </div>
                <div className="mt-2">
                    {this.props.extension.rawManifest === null ? (
                        <ExtensionNoManifestAlert extension={this.props.extension} />
                    ) : this.state.viewMode === ViewMode.Rich ? (
                        <DynamicallyImportedMonacoSettingsEditor
                            id="registry-extension-edit-page__data"
                            value={this.props.extension.rawManifest}
                            height={500}
                            jsonSchema={extensionSchemaJSON}
                            readOnly={true}
                            isLightTheme={this.props.isLightTheme}
                            history={this.props.history}
                            telemetryService={this.props.telemetryService}
                        />
                    ) : (
                        <pre className={classNames('form-control', styles.plainViewer)}>
                            <Code>{this.props.extension.rawManifest}</Code>
                        </pre>
                    )}
                </div>
            </div>
        )
    }

    private onViewModeButtonClick = (): void => {
        this.setState(
            previousState => ({ viewMode: previousState.viewMode === ViewMode.Rich ? ViewMode.Plain : ViewMode.Rich }),
            () => RegistryExtensionManifestPage.setViewMode(this.state.viewMode)
        )
    }
}
