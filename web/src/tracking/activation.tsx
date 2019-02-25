import H from 'history'
import { pick } from 'lodash'
import { Observable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { ActivationStatus, ActivationStep } from '../../../shared/src/components/activation/Activation'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { queryGraphQL } from '../backend/graphql'
import { logUserEvent } from '../user/account/backend'

/**
 * Returns the URL to navigate to when the user clicks on the "Find references" activation step.
 */
const fetchReferencesLink: () => Observable<string | null> = () =>
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

const fetchActivationStatus = (isSiteAdmin: boolean) => () =>
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
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const authProviders = window.context.authProviders
            const completed: { [key: string]: boolean } = {
                didSearch: !!data.currentUser && data.currentUser.usageStatistics.searchQueries > 0,
                'action:findReferences':
                    !!data.currentUser && data.currentUser.usageStatistics.findReferencesActions > 0,
            }
            if (isSiteAdmin) {
                completed.connectedCodeHost = data.externalServices && data.externalServices.totalCount > 0
                completed.enabledRepository =
                    data.repositories && data.repositories.totalCount !== null && data.repositories.totalCount > 0
                if (authProviders) {
                    completed.enabledSignOn =
                        data.users.totalCount > 1 || authProviders.filter(p => !p.isBuiltin).length > 0
                }
            }
            return completed
        })
    )

/**
 * Returns a new ActivationStatus instance to use in the web app.
 *
 * @param isSiteAdmin determines if site-admin-only activation steps are included.
 */
export const createActivationStatus = (isSiteAdmin: boolean) => {
    const s = new ActivationStatus(
        [
            {
                id: 'connectedCodeHost',
                title: 'Connect your code host',
                detail: 'Configure Sourcegraph to talk to your code host and fetch a list of your repositories.',
                action: (h: H.History) => h.push('/site-admin/external-services'),
                siteAdminOnly: true,
            },
            {
                id: 'enabledRepository',
                title: 'Enable repositories',
                detail: 'Select which repositories Sourcegraph should pull and index from your code host(s).',
                action: (h: H.History) => h.push('/site-admin/repositories'),
                siteAdminOnly: true,
            },
            {
                id: 'didSearch',
                title: 'Search your code',
                detail: 'Perform a search query on your code.',
                action: (h: H.History) => h.push('/search'),
            },
            {
                id: 'action:findReferences',
                title: 'Find some references',
                detail:
                    'To find references of a token, navigate to a code file in one of your repositories, hover over a token to activate the tooltip, and then click "Find references".',
                action: (h: H.History) =>
                    fetchReferencesLink()
                        .pipe(first())
                        .subscribe(r => {
                            if (r) {
                                h.push(r)
                            } else {
                                alert('Must add repositories before finding references')
                            }
                        }),
            },
            {
                id: 'enabledSignOn',
                title: 'Configure SSO or share with teammates',
                detail: 'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
                action: () => window.open('https://docs.sourcegraph.com/admin/auth', '_blank'),
                siteAdminOnly: true,
            },
        ]
            .filter(e => true || !e.siteAdminOnly)
            .map(e => pick<any, keyof ActivationStep>(e, 'id', 'title', 'detail', 'action')),
        fetchActivationStatus(true)
    )

    // Subscribe to activation events that require server updates.
    // Only certain events require server updates here, because others
    // trigger server updates elsewhere.
    s.updateCompleted.subscribe(u => {
        if (u['action:findReferences']) {
            logUserEvent(GQL.UserEvent.CODEINTELREFS)
        }
    })

    return s
}
