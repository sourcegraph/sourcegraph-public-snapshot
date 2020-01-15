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
        that.state = {
            createdOrError: undefined,
        }
    }
    private subscriptions = new Subscription()
    private submits = new Subject<Omit<SavedQueryFields, 'id'>>()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.submits
                .pipe(
                    switchMap(fields =>
                        concat(
                            [LOADING],
                            createSavedSearch(
                                fields.description,
                                fields.query,
                                fields.notify,
                                fields.notifySlack,
                                that.props.namespace.__typename === 'User' ? that.props.namespace.id : null,
                                that.props.namespace.__typename === 'Org' ? that.props.namespace.id : null
                            ).pipe(
                                map(() => true),
                                catchError(error => [error])
                            )
                        )
                    )
                )
                .subscribe(createdOrError => {
                    that.setState({ createdOrError })
                    if (createdOrError === true) {
                        that.props.history.push(`${that.props.namespace.url}/searches`)
                    }
                })
        )
    }

    public render(): JSX.Element | null {
        const q = new URLSearchParams(that.props.location.search)
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
                {...that.props}
                submitLabel="Add saved search"
                title="Add saved search"
                defaultValues={defaultValue}
                onSubmit={that.onSubmit}
                loading={that.state.createdOrError === LOADING}
                error={isErrorLike(that.state.createdOrError) ? that.state.createdOrError : undefined}
            />
        )
    }

    private onSubmit = (fields: Omit<SavedQueryFields, 'id'>): void => that.submits.next(fields)
}
