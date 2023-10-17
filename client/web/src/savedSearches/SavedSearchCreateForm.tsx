import React, { type FC } from 'react'

import { useLocation, useNavigate, type NavigateFunction, type Location } from 'react-router-dom'
import { concat, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import type { Omit } from 'utility-types'

import { type ErrorLike, isErrorLike, asError } from '@sourcegraph/common'
import { screenReaderAnnounce } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import type { NamespaceProps } from '../namespaces'
import { createSavedSearch } from '../search/backend'
import { eventLogger } from '../tracking/eventLogger'

import { type SavedQueryFields, SavedSearchForm } from './SavedSearchForm'

interface Props extends NamespaceProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    location: Location
    navigate: NavigateFunction
}

const LOADING = 'loading' as const

interface State {
    createdOrError: undefined | typeof LOADING | true | ErrorLike
}

export const SavedSearchCreateForm: FC<Omit<Props, 'location' | 'navigate'>> = props => {
    const location = useLocation()
    const navigate = useNavigate()

    return <InnerSavedSearchCreateForm {...props} location={location} navigate={navigate} />
}

class InnerSavedSearchCreateForm extends React.Component<Props, State> {
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
                                map(() => true as const),
                                catchError((error): [ErrorLike] => [asError(error)])
                            )
                        ).pipe(map(createdOrError => [createdOrError, fields.description] as const))
                    )
                )
                .subscribe(([createdOrError, queryDescription]) => {
                    this.setState({ createdOrError })
                    if (createdOrError === true) {
                        eventLogger.log('SavedSearchCreated')
                        screenReaderAnnounce(`Saved ${queryDescription} search`)
                        this.props.navigate(`${this.props.namespace.url}/searches`, {
                            state: { description: queryDescription },
                        })
                    }
                })
        )
        eventLogger.logViewEvent('NewSavedSearchPage')
    }

    public render(): JSX.Element | null {
        const searchParameters = new URLSearchParams(this.props.location.search)
        let defaultValue: Partial<SavedQueryFields> = {}
        const query = searchParameters.get('query')
        const patternType = searchParameters.get('patternType')

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
