import H from 'history'
import { pick } from 'lodash'
import { Observable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { ActivationStatus, ActivationStep } from '../../../shared/src/components/activation/Activation'
import { dataAndErrors, dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import { queryGraphQL } from '../backend/graphql'

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
        map(dataAndErrors),
        map(dataAndErrors => {
            if (!dataAndErrors.data) {
                return null
            }
            const data = dataAndErrors.data
            if (!data.repositories.nodes) {
                return null
            }
            const rURLs = data.repositories.nodes
                .filter(r => r.gitRefs && r.gitRefs.totalCount > 0)
                .sort(r => (r.gitRefs ? -r.gitRefs.totalCount : 0))
                .map(r => r.url)
            if (rURLs.length === 0) {
                return null
            }
            return rURLs[0]
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
                              findRefsActions
                          }
                      }
                  }
              `
            : gql`
                  query {
                      currentUser {
                          usageStatistics {
                              searchQueries
                              findRefsActions
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
                'action:findReferences': !!data.currentUser && data.currentUser.usageStatistics.findRefsActions > 0,
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

// TODO(PROTOTYPE)
export const globalActivationStatus = new ActivationStatus(
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
            detail: 'Issue a search query over your code.',
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
                    .subscribe(r => r && h.push(r)),
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
globalActivationStatus.update(null)

export const newActivationStatus = (isSiteAdmin: boolean) => globalActivationStatus

// /**
//  * Returns a new ActivationStatus instance to use in the web app.
//  *
//  * @param isSiteAdmin determines if site-admin-only activation steps are included.
//  */
// export const newActivationStatus = (isSiteAdmin: boolean) => {
//     const s = new ActivationStatus(
//         [
//             {
//                 id: 'connectedCodeHost',
//                 title: 'Connect your code host',
//                 detail: 'Configure Sourcegraph to talk to your code host and fetch a list of your repositories.',
//                 action: (h: H.History) => h.push('/site-admin/external-services'),
//                 siteAdminOnly: true,
//             },
//             {
//                 id: 'enabledRepository',
//                 title: 'Enable repositories',
//                 detail: 'Select which repositories Sourcegraph should pull and index from your code host(s).',
//                 action: (h: H.History) => h.push('/site-admin/repositories'),
//                 siteAdminOnly: true,
//             },
//             {
//                 id: 'didSearch',
//                 title: 'Search your code',
//                 detail: 'Issue a search query over your code.',
//                 action: (h: H.History) => h.push('/search'),
//             },
//             {
//                 id: 'action:findReferences',
//                 title: 'Find references',
//                 detail:
//                     'To find references of a token, navigate to a code file in one of your repositories, hover over a token to activate the tooltip, and then click "Find references".',
//                 action: (h: H.History) =>
//                     fetchReferencesLink()
//                         .pipe(first())
//                         .subscribe(r => r && h.push(r)),
//             },
//             {
//                 id: 'enabledSignOn',
//                 title: 'Configure sign-on or share',
//                 detail: 'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
//                 action: () => window.open('https://docs.sourcegraph.com/admin/auth', '_blank'),
//                 siteAdminOnly: true,
//             },
//         ]
//             .filter(e => isSiteAdmin || !e.siteAdminOnly)
//             .map(e => pick<any, keyof ActivationStep>(e, 'id', 'title', 'detail', 'action')),
//         fetchActivationStatus(isSiteAdmin)
//     )
//     s.update(null)
//     return s
// }
