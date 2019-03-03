import H from 'history'
import { BehaviorSubject, Observable, Subject } from 'rxjs'
import { first, map, skip, switchMap } from 'rxjs/operators'
import { ActivationCompleted, ActivationStep } from '../../../shared/src/components/activation/Activation'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { queryGraphQL } from '../backend/graphql'
import { logUserEvent } from '../user/account/backend'

/**
 * Encapsulates fetching and updating activation status.
 */
export class ActivationStatus {
    public steps: ActivationStep[]
    public completed = new BehaviorSubject<ActivationCompleted | null>(null)
    private completedFirstFetch: Promise<void>
    private refetchRequested = new Subject<void>()

    constructor(private authenticatedUser: GQL.IUser) {
        this.steps = getActivationSteps(authenticatedUser)
        this.completedFirstFetch = this.completed
            .pipe(
                skip(1),
                first(),
                map(() => void 0)
            )
            .toPromise()
        this.refetchRequested
            .pipe(switchMap(() => fetchActivationStatus(this.authenticatedUser.siteAdmin)))
            .subscribe(completed => this.completed.next(completed))
        this.refetchCompleted()
    }

    public refetchCompleted = (): void => this.refetchRequested.next()

    public updateCompleted = (update: ActivationCompleted): void => {
        this.completedFirstFetch.then(() => {
            // Send update to server for events that don't themselves trigger
            // an update.
            if (update.FoundReferences) {
                logUserEvent(GQL.UserEvent.CODEINTELREFS)
            }

            const newVal: ActivationCompleted = {}
            Object.assign(newVal, this.completed.value)
            for (const step of this.steps) {
                if (update[step.id] !== undefined) {
                    newVal[step.id] = update[step.id]
                }
            }
            this.completed.next(newVal)
        })
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

const fetchActivationStatus = (isSiteAdmin: boolean): Promise<ActivationCompleted> =>
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
                const completed: ActivationCompleted = {
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
