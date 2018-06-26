import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, map, tap } from 'rxjs/operators'
import { gql, mutateGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { Form } from '../components/Form'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { queryViewerRegistryPublishers } from './backend'
import { DynamicallyImportedMonacoSettingsEditor } from './DynamicallyImportedMonacoSettingsEditor'
import { RegistryPublisher, toExtensionID } from './extension'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionDeleteButton } from './RegistryExtensionDeleteButton'
import { RegistryExtensionNameFormGroup, RegistryPublisherFormGroup } from './RegistryExtensionForm'

function updateExtension(
    args: GQL.IUpdateExtensionOnExtensionRegistryMutationArguments
): Observable<GQL.IExtensionRegistryUpdateExtensionResult> {
    return mutateGraphQL(
        gql`
            mutation UpdateRegistryExtension($extension: ID!, $name: String, $manifest: String) {
                extensionRegistry {
                    updateExtension(extension: $extension, name: $name, manifest: $manifest) {
                        extension {
                            url
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.extensionRegistry ||
                !data.extensionRegistry.updateExtension ||
                (errors && errors.length > 0)
            ) {
                throw createAggregateError(errors)
            }
            return data.extensionRegistry.updateExtension
        })
    )
}

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

interface State {
    /** The viewer's authorized publishers, 'loading', or an error. */
    publishersOrError: 'loading' | RegistryPublisher[] | ErrorLike

    name?: string
    manifest?: string

    /** The update result, undefined if not triggered, 'loading', or an error. */
    updateOrError?: 'loading' | GQL.IExtensionRegistryUpdateExtensionResult | ErrorLike
}

/** A page with a form to edit information about an extension in the extension registry. */
export class RegistryExtensionEditPage extends React.PureComponent<Props, State> {
    public state: State = {
        publishersOrError: 'loading',
    }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionEdit')

        this.subscriptions.add(
            concat(
                [{ publishersOrError: 'loading' }],
                queryViewerRegistryPublishers().pipe(
                    map(result => ({ publishersOrError: result, publisher: result[0] && result[0].id })),
                    catchError(error => [{ publishersOrError: asError(error) }])
                )
            ).subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    concatMap(() =>
                        concat(
                            [{ updateOrError: 'loading' }],
                            updateExtension({
                                extension: this.props.extension.id,
                                name: this.state.name,
                                manifest: this.state.manifest,
                            }).pipe(
                                tap(result => {
                                    // Redirect to the extension's new URL (if it changed).
                                    if (this.props.extension.url !== result.extension.url) {
                                        this.props.history.push(result.extension.url + '/-/edit')
                                    }
                                    this.props.onDidUpdateExtension()
                                }),
                                map(result => ({ updateOrError: result })),
                                catchError(error => [{ updateOrError: asError(error) }])
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in
        if (!this.props.authenticatedUser) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        if (!this.props.extension.viewerCanAdminister) {
            return (
                <HeroPage
                    icon={ErrorIcon}
                    title="Unauthorized"
                    subtitle="You are not authorized to edit this extension."
                />
            )
        }

        const publisher = this.props.extension.publisher
        if (!publisher) {
            return <HeroPage icon={ErrorIcon} title="Publisher not found" />
        }

        const extensionName = this.state.name === undefined ? this.props.extension.name : this.state.name

        let extensionID: string | undefined
        if (this.state.publishersOrError !== 'loading' && !isErrorLike(this.state.publishersOrError)) {
            const p = this.state.publishersOrError.find(p => p.id === publisher.id)
            if (p) {
                extensionID = toExtensionID(p, extensionName)
            }
        }

        let manifestValue: string
        if (this.state.manifest !== undefined) {
            manifestValue = this.state.manifest
        } else if (this.props.extension.manifest !== null) {
            manifestValue = this.props.extension.manifest.raw
        } else {
            manifestValue = `// This extension manifest describes the extension, how to run it, and when to activate it.
{
  // How does Sourcegraph communicate with the extension?
  "platform": {
    "type": "tcp", // also supported: websocket, exec
    "address": "host:port" // set other field(s) depending on the platform type
  },

  // This means the extension is always activated (for every file and repository). Selective
  // activation is not yet supported.
  "activationEvents": [
    "*"
  ]
}`
        }

        return (
            <div className="registry-extension-edit-page">
                <PageTitle title="Edit extension" />
                <h2>Edit extension</h2>
                <Form onSubmit={this.onSubmit}>
                    <RegistryPublisherFormGroup
                        className="registry-extension-edit-page__input"
                        value={publisher.id}
                        publishersOrError={this.state.publishersOrError}
                        disabled={true}
                    />
                    <RegistryExtensionNameFormGroup
                        className="registry-extension-edit-page__input"
                        value={extensionName}
                        onChange={this.onNameChange}
                        disabled={this.state.updateOrError === 'loading'}
                    />
                    {extensionID &&
                        this.state.name &&
                        this.state.name !== this.props.extension.name && (
                            <div className="alert alert-primary">
                                Extension will be renamed. New extension ID:{' '}
                                <code id="registry-extension__extensionID">
                                    <strong>{extensionID}</strong>
                                </code>
                            </div>
                        )}
                    <div className="form-group">
                        <label htmlFor="registry-extension-edit-page__manifest" className="d-block">
                            Manifest
                        </label>
                        <DynamicallyImportedMonacoSettingsEditor
                            id="registry-extension-edit-page__manifest"
                            value={manifestValue}
                            onChange={this.onManifestChange}
                            jsonSchema="https://sourcegraph.com/v1/extension.schema.json#"
                            readOnly={this.state.updateOrError === 'loading'}
                            isLightTheme={this.props.isLightTheme}
                            history={this.props.history}
                        />
                    </div>
                    <button type="submit" disabled={this.state.updateOrError === 'loading'} className="btn btn-primary">
                        {this.state.updateOrError === 'loading' ? (
                            <LoaderIcon className="icon-inline" />
                        ) : (
                            'Update extension'
                        )}
                    </button>
                </Form>
                {isErrorLike(this.state.updateOrError) && (
                    <div className="alert alert-danger">{upperFirst(this.state.updateOrError.message)}</div>
                )}
                <hr />
                <RegistryExtensionDeleteButton
                    className="mb-3"
                    extension={this.props.extension}
                    onDidUpdate={this.onDidDelete}
                />
            </div>
        )
    }

    private onNameChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ name: e.currentTarget.value })

    private onManifestChange = (newValue: string) => this.setState({ manifest: newValue })

    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)

    private onDidDelete = () => {
        this.props.history.push('/registry')
        this.props.onDidUpdateExtension()
    }
}
