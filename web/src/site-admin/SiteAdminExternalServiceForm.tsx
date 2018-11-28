import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { upperFirst } from 'lodash'
import * as React from 'react'
import siteSchemaJSON from '../../../schema/site.schema.json'
import * as GQL from '../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../shared/src/util/errors'
import { Form } from '../components/Form'
import { DynamicallyImportedMonacoSettingsEditor } from '../settings/DynamicallyImportedMonacoSettingsEditor'

interface Props {
    history: H.History
    input: GQL.IAddExternalServiceInput
    isLightTheme: boolean
    error?: ErrorLike
    mode: 'edit' | 'create'
    loading: boolean
    onSubmit: (event?: React.FormEvent<HTMLFormElement>) => void
    onChange: (change: GQL.IAddExternalServiceInput) => void
}

export class SiteAdminExternalServiceForm extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <Form className="external-service-form" onSubmit={this.props.onSubmit}>
                {this.props.error && <p className="alert alert-danger">{upperFirst(this.props.error.message)}</p>}
                <div className="form-group">
                    <label htmlFor="external-service-form-display-name">Display name</label>
                    <input
                        id="external-service-form-display-name"
                        type="text"
                        className="form-control"
                        placeholder="ACME GitHub Enterprise"
                        required={true}
                        autoCorrect="off"
                        autoComplete="off"
                        autoFocus={true}
                        value={this.props.input.displayName}
                        onChange={this.onDisplayNameChange}
                        disabled={this.props.loading}
                    />
                </div>

                <div className="form-group">
                    <label htmlFor="external-service-page-form-kind">Kind</label>
                    <input
                        id="external-service-page-form-kind"
                        type="text"
                        className="form-control"
                        placeholder="GITHUB"
                        required={true}
                        autoCorrect="off"
                        value={this.props.input.kind}
                        onChange={this.onKindChange}
                        disabled={this.props.loading || this.props.mode === 'edit'}
                    />
                </div>

                <div>
                    <DynamicallyImportedMonacoSettingsEditor
                        value={this.props.input.config}
                        jsonSchemaId={`site.schema.json#definitions/${getKindDefinitionId(this.props.input.kind)}`}
                        extraSchemas={[siteSchemaJSON]}
                        canEdit={false}
                        loading={this.props.loading}
                        height={300}
                        isLightTheme={this.props.isLightTheme}
                        onSave={this.onSave}
                        onChange={this.onConfigChange}
                        history={this.props.history}
                    />
                    <p className="form-text text-muted">
                        <small>Use Ctrl+Space for completion, and hover over JSON properties for documentation.</small>
                    </p>
                </div>
                <button type="submit" className="btn btn-primary" disabled={this.props.loading}>
                    {this.props.mode === 'edit' ? 'Update' : 'Add external service'}
                </button>
                {this.props.loading && <LoadingSpinner className="icon-inline" />}
            </Form>
        )
    }

    private onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.props.onChange({ ...this.props.input, displayName: event.currentTarget.value })
    }

    private onKindChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.props.onChange({ ...this.props.input, kind: event.currentTarget.value as GQL.ExternalServiceKind })
    }

    private onConfigChange = (config: string) => {
        this.props.onChange({ ...this.props.input, config })
    }

    private onSave = () => {
        this.props.onSubmit()
    }
}

function getKindDefinitionId(kind: GQL.ExternalServiceKind): string {
    switch (kind) {
        case GQL.ExternalServiceKind.AWSCODECOMMIT:
            return 'AWSCodeCommitConnection'
        case GQL.ExternalServiceKind.BITBUCKETSERVER:
            return 'BitbucketServerConnection'
        case GQL.ExternalServiceKind.GITHUB:
            return 'GitHubConnection'
        case GQL.ExternalServiceKind.GITLAB:
            return 'GitLabConnection'
        case GQL.ExternalServiceKind.GITOLITE:
            return 'GitoliteConnection'
        case GQL.ExternalServiceKind.PHABRICATOR:
            return 'PhabricatorConnection'
    }
    throw new Error(`unknown external service kind: ${kind}`)
}
