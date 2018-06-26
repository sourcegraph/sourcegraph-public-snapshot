import InfoIcon from '@sourcegraph/icons/lib/Info'
import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import ViewIcon from '@sourcegraph/icons/lib/View'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { DynamicallyImportedMonacoSettingsEditor } from './DynamicallyImportedMonacoSettingsEditor'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'

export const RegistryExtensionNoManifestAlert: React.SFC<{
    extension: Pick<GQL.IRegistryExtension, 'viewerCanAdminister' | 'url'>
}> = ({ extension }) => (
    <div className="alert alert-info">
        This extension's publisher hasn't yet provided an extension manifest.
        {extension.viewerCanAdminister && (
            <>
                <br />
                <Link className="mt-3 btn btn-primary" to={`${extension.url}/-/edit`}>
                    <PencilIcon className="icon-inline" /> Edit manifest
                </Link>
            </>
        )}
    </div>
)

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

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
        const v = localStorage.getItem(RegistryExtensionManifestPage.STORAGE_KEY)
        if (v === ViewMode.Rich || v === ViewMode.Plain) {
            return v
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
            <div className="registry-extension-manifest-page">
                <PageTitle title={`Manifest of ${this.props.extension.extensionID}`} />
                <div className="d-flex align-items-center justify-content-between">
                    <div className="d-flex align-items-center">
                        <h3 className="mb-0 mr-1">Manifest</h3>
                        <InfoIcon
                            className="icon-inline text-muted"
                            data-tooltip="The published JSON description of how to run or access the extension"
                        />
                    </div>
                    <div>
                        {this.props.extension.manifest && (
                            <button type="button" className="btn btn-secondary" onClick={this.onViewModeButtonClick}>
                                <ViewIcon className="icon-inline" /> Use{' '}
                                {this.state.viewMode === ViewMode.Plain ? ViewMode.Rich : ViewMode.Plain} viewer
                            </button>
                        )}{' '}
                        {this.props.extension.viewerCanAdminister && (
                            <Link className="btn btn-primary" to={`${this.props.extension.url}/-/edit`}>
                                <PencilIcon className="icon-inline" /> Edit manifest
                            </Link>
                        )}
                    </div>
                </div>
                <div className="mt-2">
                    {this.props.extension.manifest === null ? (
                        <RegistryExtensionNoManifestAlert extension={this.props.extension} />
                    ) : this.state.viewMode === ViewMode.Rich ? (
                        <DynamicallyImportedMonacoSettingsEditor
                            id="registry-extension-edit-page__data"
                            value={this.props.extension.manifest.raw}
                            height={500}
                            jsonSchema="https://sourcegraph.com/v1/extension.schema.json#"
                            readOnly={true}
                            isLightTheme={this.props.isLightTheme}
                            history={this.props.history}
                        />
                    ) : (
                        <pre className="form-control">
                            <code>{this.props.extension.manifest.raw}</code>
                        </pre>
                    )}
                </div>
            </div>
        )
    }

    private onViewModeButtonClick = () => {
        this.setState(
            prevState => ({ viewMode: prevState.viewMode === ViewMode.Rich ? ViewMode.Plain : ViewMode.Rich }),
            () => RegistryExtensionManifestPage.setViewMode(this.state.viewMode)
        )
    }
}
