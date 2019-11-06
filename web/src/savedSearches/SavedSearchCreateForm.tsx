import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Omit } from 'utility-types'
import * as GQL from '../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { NamespaceProps } from '../namespaces'
import { createSavedSearch } from '../search/backend'
import { SavedQueryFields, SavedSearchForm } from './SavedSearchForm'

interface Props extends RouteComponentProps, NamespaceProps {
    authenticatedUser: GQL.IUser | null
}

const LOADING: 'loading' = 'loading'

interface State {
    createdOrError: undefined | typeof LOADING | true | ErrorLike
}

export class SavedSearchCreateForm extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            createdOrError: undefined,
        }
    }
    private subscriptions = new Subscription()
    private submits = new Subject<Omit<SavedQueryFields, 'id'>>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(fields =>
                        concat(
                            [LOADING],
                            createSavedSearch(
                                fields.description,
                                fields.query,
                                fields.notify,
                                fields.notifySlack,
                                this.props.namespace.__typename === 'User' ? this.props.namespace.id : null,
                                this.props.namespace.__typename === 'Org' ? this.props.namespace.id : null
                            ).pipe(
                                map(() => true),
                                catchError(error => [error])
                            )
                        )
                    )
                )
                .subscribe(createdOrError => {
                    this.setState({ createdOrError })
                    if (createdOrError === true) {
                        this.props.history.push(`${this.props.namespace.url}/searches`)
                    }
                })
        )
    }

    public render(): JSX.Element | null {
        const q = new URLSearchParams(this.props.location.search)
        let defaultValue: Partial<SavedQueryFields> = {}
        const query = q.get('query')
        const patternType = q.get('patternType')

        if (query && patternType) {
            defaultValue = { query: query + ` patternType:${patternType}` }
        } else if (query) {
            defaultValue = { query }
        }

        return (
            <SavedSearchForm
                {...this.props}
                submitLabel="Add saved search"
                title="Add saved search"
                defaultValues={defaultValue}
                onSubmit={this.onSubmit}
                loading={this.state.createdOrError === LOADING}
                error={isErrorLike(this.state.createdOrError) ? this.state.createdOrError : undefined}
            />
        )
    }

    private onSubmit = (fields: Omit<SavedQueryFields, 'id'>): void => this.submits.next(fields)
}
