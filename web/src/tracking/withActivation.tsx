import H from 'history'
import React from 'react'
import { combineLatest, merge, Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, first, map, scan, startWith, switchMap, tap } from 'rxjs/operators'
import {
    ActivationCompletionStatus,
    ActivationProps,
    ActivationStep,
} from '../../../shared/src/components/activation/Activation'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { queryGraphQL } from '../backend/graphql'
import { logUserEvent } from '../user/account/backend'

/**
 * Fetches activation status from server.
 */
const fetchActivationStatus = (isSiteAdmin: boolean): Observable<ActivationCompletionStatus> =>
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
                              codeIntelligenceActions
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
                              codeIntelligenceActions
                          }
                      }
                  }
              `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const authProviders = window.context.authProviders
            const usageStats = !!data.currentUser && data.currentUser.usageStatistics
            const completed: ActivationCompletionStatus = {
                DidSearch: usageStats && usageStats.searchQueries > 0,
                FoundReferences:
                    // TODO(beyang): revert this to usageStats.findReferencesActions > 0 in 3.3 or later.
                    // Remove codeIntelligenceActions from the GraphQL query above, as well.
                    usageStats && (usageStats.findReferencesActions > 0 || usageStats.codeIntelligenceActions > 10),
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

/**
 * Returns the link a user should go to when they click on the uncompleted find-references
 * activation step. For now, this links to root page of a repository, but we could improve
 * this by linking to a code file or actual symbol.
 */
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

/**
 * Gets the activation steps that need to be completed for a given user.
 */
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

/**
 * Sends update to server for events that don't themselves trigger
 * an update.
 */
const recordUpdate = (update: Partial<ActivationCompletionStatus>): void => {
    if (update.FoundReferences) {
        logUserEvent(GQL.UserEvent.CODEINTELREFS)
    }
}

interface Props {
    authenticatedUser: GQL.IUser | null
}

interface State {
    completed?: ActivationCompletionStatus
}

/**
 * Modifies the input component to return a component that includes the activation status in the
 * `activation` field of its props.
 */
export const withActivation = <P extends ActivationProps & Props>(Component: React.ComponentType<P>) =>
    class WithActivation extends React.Component<Props & Pick<P, Exclude<keyof P, keyof ActivationProps>>, State> {
        private subscriptions = new Subscription()
        private componentUpdates = new Subject<Readonly<Props & Pick<P, Exclude<keyof P, keyof ActivationProps>>>>()
        public state: State = {}

        /**
         * Calling `next` triggers refetches. This ensures at most one refetch request is outstanding
         * at any given time.
         */
        private refetches = new Subject<void>()

        private updates = new Subject<Partial<ActivationCompletionStatus>>()

        public componentDidMount(): void {
            const authenticatedUser: Observable<GQL.IUser | null> = this.componentUpdates.pipe(
                startWith(this.props),
                map(props => props.authenticatedUser),
                distinctUntilChanged()
            )
            const serverCompletionStatus: Observable<ActivationCompletionStatus> = combineLatest(
                authenticatedUser,
                this.refetches.pipe(startWith(undefined))
            ).pipe(
                switchMap(([authenticatedUser]) =>
                    authenticatedUser ? fetchActivationStatus(authenticatedUser.siteAdmin) : []
                )
            )
            const localCompletionStatus: Observable<Partial<ActivationCompletionStatus>> = merge(
                authenticatedUser.pipe(map(() => null)), // reset on new authenticated user
                this.updates
            ).pipe(
                tap(update => update && recordUpdate(update)),
                scan<Partial<ActivationCompletionStatus> | null, Partial<ActivationCompletionStatus>>(
                    (prev, next) => (next ? { ...prev, ...next } : {}),
                    {}
                )
            )
            this.subscriptions.add(
                combineLatest(serverCompletionStatus, localCompletionStatus)
                    .pipe(
                        map(([serverCompletionStatus, localCompletionStatus]) => ({
                            ...serverCompletionStatus,
                            ...localCompletionStatus,
                        }))
                    )
                    .subscribe(completed => this.setState({ completed }))
            )
        }

        public componentWillUnmount(): void {
            this.subscriptions.unsubscribe()
        }

        public componentDidUpdate(): void {
            this.componentUpdates.next(this.props)
        }

        private steps(): ActivationStep[] | undefined {
            const user: GQL.IUser | null = this.props.authenticatedUser
            if (user) {
                return getActivationSteps(user)
            }
            return undefined
        }

        private refetchCompletionStatus = () => this.refetches.next()

        private updateCompletionStatus = (update: Partial<ActivationCompletionStatus>): void =>
            this.updates.next(update)

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
