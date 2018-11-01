import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, map, tap } from 'rxjs/operators'
import { withAuthenticatedUser } from '../../../../packages/webapp/src/auth/withAuthenticatedUser'
import { gql, mutateGraphQL } from '../../../../packages/webapp/src/backend/graphql'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'
import { Form } from '../../../../packages/webapp/src/components/Form'
import { HeroPage } from '../../../../packages/webapp/src/components/HeroPage'
import { PageTitle } from '../../../../packages/webapp/src/components/PageTitle'
import { toExtensionID } from '../../../../packages/webapp/src/extensions/extension/extension'
import { ExtensionAreaRouteContext } from '../../../../packages/webapp/src/extensions/extension/ExtensionArea'
import { eventLogger } from '../../../../packages/webapp/src/tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../packages/webapp/src/util/errors'
import { RegistryExtensionDeleteButton } from './RegistryExtensionDeleteButton'
import { RegistryExtensionNameFormGroup, RegistryPublisherFormGroup } from './RegistryExtensionForm'

function updateExtension(
    args: Pick<
        GQL.IUpdateExtensionOnExtensionRegistryMutationArguments,
        Exclude<keyof GQL.IUpdateExtensionOnExtensionRegistryMutationArguments, 'manifest'>
    >
): Observable<GQL.IExtensionRegistryUpdateExtensionResult> {
    return mutateGraphQL(
        gql`
            mutation UpdateRegistryExtension($extension: ID!, $name: String) {
                extensionRegistry {
                    updateExtension(extension: $extension, name: $name) {
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

interface Props extends ExtensionAreaRouteContext, RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser
}

interface State {
    name?: string

    /** The update result, undefined if not triggered, 'loading', or an error. */
    updateOrError?: 'loading' | GQL.IExtensionRegistryUpdateExtensionResult | ErrorLike
}

/** A page for managing an extension in the extension registry. */
export const RegistryExtensionManagePage = withAuthenticatedUser(
    class RegistryExtensionManagePage extends React.PureComponent<Props, State> {
        public state: State = {}

        private submits = new Subject<React.FormEvent<HTMLFormElement>>()
        private componentUpdates = new Subject<Props>()
        private subscriptions = new Subscription()

        public componentDidMount(): void {
            eventLogger.logViewEvent('RegistryExtensionManage')

            this.subscriptions.add(
                this.submits
                    .pipe(
                        tap(e => e.preventDefault()),
                        concatMap(() =>
                            concat(
                                [{ updateOrError: 'loading' }],
                                updateExtension({
                                    extension: this.props.extension.registryExtension!.id,
                                    name: this.state.name,
                                }).pipe(
                                    tap(result => {
                                        // Redirect to the extension's new URL (if it changed).
                                        if (this.props.extension.registryExtension!.url !== result.extension.url) {
                                            this.props.history.push(result.extension.url + '/-/manage')
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
            if (
                !this.props.extension.registryExtension ||
                !this.props.extension.registryExtension.viewerCanAdminister
            ) {
                return (
                    <HeroPage
                        icon={AlertCircleIcon}
                        title="Unauthorized"
                        subtitle="You are not authorized to adminster this extension."
                    />
                )
            }

            const publisher = this.props.extension.registryExtension.publisher
            if (!publisher) {
                return <HeroPage icon={AlertCircleIcon} title="Publisher not found" />
            }

            const extensionName =
                this.state.name === undefined ? this.props.extension.registryExtension.name : this.state.name

            const extensionID = toExtensionID(publisher, extensionName)

            return (
                <div className="registry-extension-manage-page">
                    <PageTitle title="Manage extension" />
                    <h2>Manage extension</h2>
                    <Form onSubmit={this.onSubmit}>
                        <RegistryPublisherFormGroup
                            className="registry-extension-manage-page__input"
                            value={publisher.id}
                            publishersOrError={[publisher]}
                            disabled={true}
                        />
                        <RegistryExtensionNameFormGroup
                            className="registry-extension-manage-page__input"
                            value={extensionName}
                            onChange={this.onNameChange}
                            disabled={this.state.updateOrError === 'loading'}
                        />
                        {extensionID &&
                            this.state.name &&
                            this.state.name !== this.props.extension.registryExtension.name && (
                                <div className="alert alert-primary">
                                    Extension will be renamed. New extension ID:{' '}
                                    <code id="registry-extension__extensionID">
                                        <strong>{extensionID}</strong>
                                    </code>
                                </div>
                            )}
                        <button
                            type="submit"
                            disabled={this.state.updateOrError === 'loading'}
                            className="btn btn-primary"
                        >
                            {this.state.updateOrError === 'loading' ? (
                                <LoadingSpinner className="icon-inline" />
                            ) : (
                                'Update extension'
                            )}
                        </button>
                    </Form>
                    {isErrorLike(this.state.updateOrError) && (
                        <div className="alert alert-danger">{upperFirst(this.state.updateOrError.message)}</div>
                    )}
                    <div className="card mt-5 registry-extension-manage-page__other-actions">
                        <div className="card-header">Other actions</div>
                        <div className="card-body">
                            <Link
                                to={`${this.props.extension.registryExtension.url}/-/releases/new`}
                                className="btn btn-success mr-2"
                            >
                                Publish new release
                            </Link>
                            <RegistryExtensionDeleteButton
                                extension={this.props.extension.registryExtension}
                                onDidUpdate={this.onDidDelete}
                            />
                        </div>
                    </div>
                </div>
            )
        }

        private onNameChange: React.ChangeEventHandler<HTMLInputElement> = e =>
            this.setState({ name: e.currentTarget.value })

        private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)

        private onDidDelete = () => {
            this.props.history.push('/extensions')
            this.props.onDidUpdateExtension()
        }
    }
)
