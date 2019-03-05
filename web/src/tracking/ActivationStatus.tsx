import H from 'history'
import React from 'react'
import { concat, Observable, ReplaySubject, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, first, map, switchMap } from 'rxjs/operators'
import {
    ActivationCompletionStatus,
    ActivationProps,
    ActivationStep,
} from '../../../shared/src/components/activation/Activation'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { queryGraphQL } from '../backend/graphql'
import { logUserEvent } from '../user/account/backend'

interface Props {
    authenticatedUser: GQL.IUser | null
}

interface State {
    completed?: ActivationCompletionStatus
}

export const withActivation = <P extends ActivationProps & Props>(Component: React.ComponentType<P>) =>
    class WithActivation extends React.Component<Props & Pick<P, Exclude<keyof P, keyof ActivationProps>>, State> {
        private subscriptions = new Subscription()
        private componentUpdates = new Subject<Props & Pick<P, Exclude<keyof P, keyof ActivationProps>>>()
        public state: State = {}

        /**
         * Calling `next` triggers refetches. This ensures at most one refetch request is outstanding
         * at any given time.
         */
        private refetches = new Subject<void>()

        /**
         * This field is reinitialized every time the activation status is reset due to a
         * change in the authenticated user. This necessitates that the activation status
         * be fetched from the server and all subsequent activation updates should block
         * on this initial fetch.
         */
        private completionStatusSet = new ReplaySubject<void>(1)

        public componentDidMount = () => {
            // Refetch activation status from server when requested
            this.subscriptions.add(
                this.refetches
                    .pipe(
                        switchMap(() =>
                            this.props.authenticatedUser
                                ? fetchActivationStatus(this.props.authenticatedUser.siteAdmin)
                                : []
                        )
                    )
                    .subscribe(completed => {
                        this.setState({ completed })
                        this.completionStatusSet.next()
                    })
            )
            // Reset the activation status when the authenticated user changes
            this.subscriptions.add(
                concat([this.props], this.componentUpdates)
                    .pipe(
                        map(props => props.authenticatedUser),
                        distinctUntilChanged()
                    )
                    .subscribe(() => {
                        this.completionStatusSet = new ReplaySubject<void>(1)
                        this.setState({ completed: undefined })
                        this.refetchCompletionStatus()
                    })
            )
        }

        public componentWillUnmount = () => {
            this.subscriptions.unsubscribe()
        }

        public componentDidUpdate = () => {
            this.componentUpdates.next(this.props as Props & Pick<P, Exclude<keyof P, keyof ActivationProps>>)
        }

        private steps = (): ActivationStep[] | undefined => {
            const user: GQL.IUser | null = this.props.authenticatedUser
            if (user) {
                return getActivationSteps(user)
            }
            return undefined
        }

        private refetchCompletionStatus = () => this.refetches.next()

        private updateCompletionStatus = (update: ActivationCompletionStatus): void => {
            this.completionStatusSet.pipe(first()).subscribe(() => {
                this.recordUpdate(update)
            })
            const steps = this.steps()
            if (!steps) {
                return
            }

            const completed: ActivationCompletionStatus = {}
            Object.assign(completed, this.state.completed)
            for (const step of steps) {
                if (update[step.id] !== undefined) {
                    completed[step.id] = update[step.id]
                }
            }
            this.setState({ completed })
        }

        /**
         * Sends update to server for events that don't themselves trigger
         * an update.
         */
        private recordUpdate = (update: ActivationCompletionStatus): void => {
            if (update.FoundReferences) {
                logUserEvent(GQL.UserEvent.CODEINTELREFS)
            }
        }

        public render(): React.ReactFragment | null {
            const steps = this.steps()
            const activationProps: ActivationProps = {
                activation: steps && {
                    steps,
                    completed: this.state.completed,
                    update: this.updateCompletionStatus,
                    refetch: this.refetchCompletionStatus,
                },
            }
            // This is safe to cast to P, because this.props has everything in P *except*
            // the properties in ActivationProps
            const props = {
                ...this.props,
                ...activationProps,
            }
            return <Component {...props as P} />
        }
    }

const getActivationSteps = (authenticatedUser: GQL.IUser): ActivationStep[] => {
    const sources: (ActivationStep & { siteAdminOnly?: boolean })[] = [
        {
            id: 'ConnectedCodeHost',
            title: 'Connect your code host',
            detail: 'Configure Sourcegraph to talk to your code host and fetch a list of your repositories.',
            link: { to: '/site-admin/external-services' },
            siteAdminOnly: true,
        },
        {
            id: 'EnabledRepository',
            title: 'Enable repositories',
            detail: 'Select which repositories Sourcegraph should pull and index from your code host(s).',
            link: { to: '/site-admin/repositories' },
            siteAdminOnly: true,
        },
        {
            id: 'DidSearch',
            title: 'Search your code',
            detail: 'Perform a search query on your code.',
            link: { to: '/search' },
        },
        {
            id: 'FoundReferences',
            title: 'Find some references',
            detail:
                'To find references of a token, navigate to a code file in one of your repositories, hover over a token to activate the tooltip, and then click "Find references".',
            onClick: (event: React.MouseEvent<HTMLElement>, history: H.History) =>
                fetchReferencesLink()
                    .pipe(first())
                    .subscribe(r => {
                        if (r) {
                            history.push(r)
                        } else {
                            alert('Must add repositories before finding references')
                        }
                    }),
        },
        {
            id: 'EnabledSharing',
            title: 'Configure SSO or share with teammates',
            detail: 'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
            link: { to: 'https://docs.sourcegraph.com/admin/auth', target: '_blank' },
            siteAdminOnly: true,
        },
    ]
    return sources
        .filter(e => authenticatedUser.siteAdmin || !e.siteAdminOnly)
        .map(({ siteAdminOnly, ...step }) => step)
}

const fetchReferencesLink = (): Observable<string | null> =>
    queryGraphQL(gql`
        query {
            repositories(enabled: true, cloned: true, first: 100, indexed: true) {
                nodes {
                    url
                    gitRefs {
                        totalCount
                    }
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repositories.nodes) {
                return null
            }
            const repositoryURLs = data.repositories.nodes
                .filter(r => r.gitRefs && r.gitRefs.totalCount > 0)
                .sort((r1, r2) => r2.gitRefs!.totalCount! - r1.gitRefs!.totalCount)
                .map(r => r.url)
            if (repositoryURLs.length === 0) {
                return null
            }
            return repositoryURLs[0]
        })
    )

const fetchActivationStatus = (isSiteAdmin: boolean): Promise<ActivationCompletionStatus> =>
    queryGraphQL(
        isSiteAdmin
            ? gql`
                  query {
                      externalServices {
                          totalCount
                      }
                      repositories(enabled: true) {
                          totalCount
                      }
                      viewerSettings {
                          final
                      }
                      users {
                          totalCount
                      }
                      currentUser {
                          usageStatistics {
                              searchQueries
                              findReferencesActions
                          }
                      }
                  }
              `
            : gql`
                  query {
                      currentUser {
                          usageStatistics {
                              searchQueries
                              findReferencesActions
                          }
                      }
                  }
              `
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => {
                const authProviders = window.context.authProviders
                const completed: ActivationCompletionStatus = {
                    DidSearch: !!data.currentUser && data.currentUser.usageStatistics.searchQueries > 0,
                    FoundReferences: !!data.currentUser && data.currentUser.usageStatistics.findReferencesActions > 0,
                }
                if (isSiteAdmin) {
                    completed.ConnectedCodeHost = data.externalServices && data.externalServices.totalCount > 0
                    completed.EnabledRepository =
                        data.repositories && data.repositories.totalCount !== null && data.repositories.totalCount > 0
                    if (authProviders) {
                        completed.EnabledSharing =
                            data.users.totalCount > 1 || authProviders.filter(p => !p.isBuiltin).length > 0
                    }
                }
                return completed
            })
        )
        .toPromise()
