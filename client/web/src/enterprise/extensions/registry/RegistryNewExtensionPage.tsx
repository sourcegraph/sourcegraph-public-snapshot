import * as React from 'react'

import { mdiPuzzleOutline } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import HelpCircleOutline from 'mdi-react/HelpCircleOutlineIcon'
import { RouteComponentProps } from 'react-router'
import { concat, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, map, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Code, Container, Icon, Label, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { BreadcrumbSetters } from '../../../components/Breadcrumbs'
import { PageTitle } from '../../../components/PageTitle'
import { RegistryPublisher, toExtensionID } from '../../../extensions/extension/extension'
import { eventLogger } from '../../../tracking/eventLogger'
import { RegistryExtensionNameFormGroup, RegistryPublisherFormGroup } from '../extension/RegistryExtensionForm'

import { createExtension, queryViewerRegistryPublishers } from './backend'
import { RegistryAreaPageProps } from './RegistryArea'

import styles from './RegistryNewExtensionPage.module.scss'

interface Props extends RegistryAreaPageProps, RouteComponentProps<{}>, BreadcrumbSetters {
    authenticatedUser: AuthenticatedUser
    history: H.History
}

interface State {
    /** The viewer's authorized publishers, undefined while loading, or an error. */
    publishersOrError: 'loading' | RegistryPublisher[] | ErrorLike

    name: string
    publisher?: Scalars['ID']

    /** The creation result, undefined while loading, or an error. */
    creationOrError?: 'loading' | GQL.IExtensionRegistryCreateExtensionResult | ErrorLike
}

/** A page with a form to create a new extension in the extension registry. */
export const RegistryNewExtensionPage = withAuthenticatedUser(
    class RegistryNewExtensionPage extends React.PureComponent<Props, State> {
        public state: State = {
            publishersOrError: 'loading',
            name: '',
        }

        private submits = new Subject<React.FormEvent<HTMLFormElement>>()
        private componentUpdates = new Subject<Props>()
        private subscriptions = new Subscription()

        public componentDidMount(): void {
            eventLogger.logViewEvent('ExtensionRegistryCreateExtension')

            this.subscriptions.add(
                concat(
                    [{ publishersOrError: 'loading' }],
                    queryViewerRegistryPublishers().pipe(
                        map(result => ({ publishersOrError: result, publisher: result[0]?.id })),
                        catchError(error => [{ publishersOrError: asError(error) }])
                    )
                ).subscribe(
                    stateUpdate => this.setState(stateUpdate as State),
                    error => console.error(error)
                )
            )

            this.subscriptions.add(
                this.submits
                    .pipe(
                        tap(event => event.preventDefault()),
                        concatMap(() =>
                            concat(
                                [{ creationOrError: 'loading' }],
                                createExtension(this.state.publisher!, this.state.name).pipe(
                                    tap(result => {
                                        // Go to the page for the newly created extension.
                                        this.props.history.push(result.extension.url)
                                    }),
                                    map(result => ({ creationOrError: result })),
                                    catchError(error => [{ creationOrError: asError(error) }])
                                )
                            )
                        )
                    )
                    .subscribe(
                        stateUpdate => this.setState(stateUpdate as State),
                        error => console.error(error)
                    )
            )

            this.subscriptions.add(
                this.props.setBreadcrumb({ key: 'create-new-extension', element: <>Create extension</> })
            )

            this.componentUpdates.next(this.props)
        }

        public componentDidUpdate(): void {
            this.componentUpdates.next(this.props)
        }

        public componentWillUnmount(): void {
            this.subscriptions.unsubscribe()
        }

        public render(): JSX.Element | null {
            let extensionID: string | undefined
            if (
                this.state.publishersOrError !== 'loading' &&
                !isErrorLike(this.state.publishersOrError) &&
                this.state.publisher
            ) {
                const publisher = this.state.publishersOrError.find(publisher => publisher.id === this.state.publisher)
                if (publisher) {
                    extensionID = toExtensionID(publisher, this.state.name)
                }
            }

            return (
                <div className="container col-8">
                    <PageTitle title="Create new extension" />
                    <PageHeader
                        path={[
                            { icon: mdiPuzzleOutline, to: '/extensions', ariaLabel: 'Extensions' },
                            { text: 'Create extension' },
                        ]}
                        description={
                            <>
                                <Link target="_blank" rel="noopener" to="/help/extensions">
                                    Learn more
                                </Link>{' '}
                                about Sourcegraph extensions{' '}
                                <Link
                                    aria-label="Learn more about extensions"
                                    target="_blank"
                                    rel="noopener"
                                    to="/help/extensions"
                                >
                                    <Icon aria-hidden={true} as={HelpCircleOutline} />
                                </Link>
                            </>
                        }
                    />
                    <Form onSubmit={this.onSubmit} className="my-4 pb-5 test-registry-new-extension">
                        <Container className="mb-4">
                            <RegistryPublisherFormGroup
                                value={this.state.publisher}
                                publishersOrError={this.state.publishersOrError}
                                onChange={this.onPublisherChange}
                                disabled={this.state.creationOrError === 'loading'}
                            />
                            <RegistryExtensionNameFormGroup
                                value={this.state.name}
                                disabled={this.state.creationOrError === 'loading'}
                                onChange={this.onNameChange}
                            />
                            {extensionID && (
                                <div className="form-group d-flex flex-wrap align-items-baseline mb-0">
                                    <Label
                                        htmlFor="extension-registry-create-extension-page__extensionID"
                                        className="mr-1 mb-0 mt-1"
                                    >
                                        Extension ID:
                                    </Label>
                                    <Code
                                        id="extension-registry-create-extension-page__extensionID"
                                        className={classNames('mt-1', styles.extensionId)}
                                        weight="bold"
                                    >
                                        {extensionID}
                                    </Code>
                                </div>
                            )}
                            {isErrorLike(this.state.creationOrError) && (
                                <ErrorAlert className="mt-3" error={this.state.creationOrError} />
                            )}
                        </Container>
                        <Button
                            type="submit"
                            disabled={
                                isErrorLike(this.state.publishersOrError) ||
                                this.state.publishersOrError === 'loading' ||
                                this.state.creationOrError === 'loading' ||
                                !this.state.name
                            }
                            variant="primary"
                            className="mr-2"
                        >
                            {this.state.creationOrError === 'loading' && (
                                <>
                                    <LoadingSpinner />{' '}
                                </>
                            )}
                            Create extension
                        </Button>
                        <Button type="button" variant="secondary" as={Link} to="..">
                            Cancel
                        </Button>
                    </Form>
                </div>
            )
        }

        private onPublisherChange: React.ChangeEventHandler<HTMLSelectElement> = event =>
            this.setState({ publisher: event.currentTarget.value })

        private onNameChange: React.ChangeEventHandler<HTMLInputElement> = event =>
            this.setState({ name: event.currentTarget.value })

        private onSubmit: React.FormEventHandler<HTMLFormElement> = event => this.submits.next(event)
    }
)
