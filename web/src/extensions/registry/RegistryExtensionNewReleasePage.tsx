import extensionSchemaJSON from '@sourcegraph/extensions-client-common/lib/schema/extension.schema.json'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, delay, map, tap } from 'rxjs/operators'
import { gql, mutateGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { DynamicallyImportedMonacoSettingsEditor } from '../../settings/DynamicallyImportedMonacoSettingsEditor'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { ExtensionAreaPageProps } from '../extension/ExtensionArea'

function updateExtension(
    args: Pick<GQL.IUpdateExtensionOnExtensionRegistryMutationArguments, 'extension' | 'manifest'>
): Observable<GQL.IExtensionRegistryUpdateExtensionResult> {
    return mutateGraphQL(
        gql`
            mutation UpdateRegistryExtension($extension: ID!, $manifest: String!) {
                extensionRegistry {
                    updateExtension(extension: $extension, manifest: $manifest) {
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

interface Props extends ExtensionAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

interface State {
    manifest?: string

    /** The update result, undefined if not triggered, 'loading', or an error. */
    updateOrError?: 'loading' | GQL.IExtensionRegistryUpdateExtensionResult | ErrorLike
}

/** A page for publishing a new release of an extension to the extension registry. */
export class RegistryExtensionNewReleasePage extends React.PureComponent<Props, State> {
    public state: State = {}

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionNewRelease')

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    concatMap(() =>
                        concat(
                            [{ updateOrError: 'loading' }],
                            updateExtension({
                                extension: this.props.extension.registryExtension!.id,
                                manifest: this.state.manifest,
                            }).pipe(
                                // If it's too fast, it feels like nothing happened.
                                delay(500),
                                tap(() => this.props.onDidUpdateExtension()),
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

        if (!this.props.extension.registryExtension || !this.props.extension.registryExtension.viewerCanAdminister) {
            return (
                <HeroPage
                    icon={ErrorIcon}
                    title="Unauthorized"
                    subtitle="You are not authorized to adminster this extension."
                />
            )
        }

        let manifestValue: string
        if (this.state.manifest !== undefined) {
            manifestValue = this.state.manifest
        } else if (this.props.extension.rawManifest !== null) {
            manifestValue = this.props.extension.rawManifest
        } else {
            manifestValue = `// This extension manifest describes the extension, how to run it, and when to activate it.
{
  // How does Sourcegraph communicate with the extension?
  "platform": {
    "type": "tcp", // also supported: websocket, exec
    "address": "host:port" // set other field(s) depending on the platform type
  },

  // "*" means the extension is always activated (for every file and repository). Use "onLanguage:foo"
  // to only activate it when the user views a file associated with the "foo" language.
  "activationEvents": [
    "*"
  ]
}`
        }

        return (
            <div className="registry-extension-new-release-page">
                <PageTitle title="Publish new release" />
                <h2>Publish new release</h2>
                <p>All users of this extension will be automatically updated to the new release.</p>
                <Form onSubmit={this.onSubmit} className="mb-3">
                    <div className="form-group">
                        <label htmlFor="registry-extension-edit-page__manifest" className="sr-only">
                            Manifest
                        </label>
                        <DynamicallyImportedMonacoSettingsEditor
                            id="registry-extension-edit-page__manifest"
                            value={manifestValue}
                            onChange={this.onManifestChange}
                            jsonSchema={extensionSchemaJSON}
                            readOnly={this.state.updateOrError === 'loading'}
                            isLightTheme={this.props.isLightTheme}
                            history={this.props.history}
                        />
                    </div>
                    <button type="submit" disabled={this.state.updateOrError === 'loading'} className="btn btn-primary">
                        {this.state.updateOrError === 'loading' ? <LoaderIcon className="icon-inline" /> : 'Publish'}
                    </button>
                </Form>
                {this.state.updateOrError &&
                    (isErrorLike(this.state.updateOrError) ? (
                        <div className="alert alert-danger">{upperFirst(this.state.updateOrError.message)}</div>
                    ) : (
                        this.state.updateOrError !== 'loading' && (
                            <div className="alert alert-success">Published release successfully.</div>
                        )
                    ))}
            </div>
        )
    }

    private onManifestChange = (newValue: string) => this.setState({ manifest: newValue, updateOrError: undefined })

    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)
}
