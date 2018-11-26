import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../backend/graphql'

/**
 * Overwrites the settings for the subject.
 */
export function overwriteSettings(subject: GQL.ID, lastID: number | null, contents: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation OverwriteSettings($subject: ID!, $lastID: Int, $contents: String!) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    overwriteSettings(contents: $contents) {
                        empty {
                            alwaysNil
                        }
                    }
                }
            }
        `,
        { subject, lastID, contents }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}
