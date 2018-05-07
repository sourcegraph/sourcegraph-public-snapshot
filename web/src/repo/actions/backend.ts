import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { createAggregateError } from '../../util/errors'

/**
 * Fetches the language server for a given language
 *
 * @return Observable that emits the language server for the given language or null if not exists
 */
export function fetchLangServer(
    language: string
): Observable<Pick<GQL.ILangServer, 'displayName' | 'homepageURL' | 'issuesURL' | 'experimental'> | null> {
    return queryGraphQL(
        gql`
            query LangServer($language: String!) {
                site {
                    langServer(language: $language) {
                        displayName
                        homepageURL
                        issuesURL
                        experimental
                    }
                }
            }
        `,
        { language }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw createAggregateError(errors)
            }
            return data.site.langServer
        })
    )
}
