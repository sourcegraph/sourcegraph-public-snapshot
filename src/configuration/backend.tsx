import { Observable } from 'rxjs'
import { map, mergeMap, take } from 'rxjs/operators'
import { dataOrThrowErrors, gql, GraphQLDocument, GraphQLResult, mutateGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { configurationCascade } from '../settings/configuration'
import { refreshConfiguration } from '../user/settings/backend'

/**
 * Overwrites the settings for the subject.
 */
export function overwriteSettings(subject: GQL.ID, lastID: number | null, contents: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation OverwriteSettings($subject: ID!, $lastID: Int, $contents: String!) {
                configurationMutation(input: { subject: $subject, lastID: $lastID }) {
                    overwriteConfiguration(contents: $contents) {
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

export function editConfiguration(
    subject: GQL.ID,
    lastID: number | null,
    edit: GQL.IConfigurationEdit
): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation EditSettings($subject: ID!, $lastID: Int, $edit: ConfigurationEdit!) {
                configurationMutation(input: { subject: $subject, lastID: $lastID }) {
                    editConfiguration(edit: $edit) {
                        empty {
                            alwaysNil
                        }
                    }
                }
            }
        `,
        { subject, lastID, edit }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

/**
 * Runs a GraphQL mutation that includes configuration mutations, populating the variables object
 * with the lastID and subject for the configuration mutation.
 *
 * @param subject The subject whose configuration to update.
 * @param mutation The GraphQL mutation.
 * @param variables The GraphQL mutation's variables.
 */
export function mutateConfigurationGraphQL(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQL.ID },
    mutation: GraphQLDocument,
    variables: any = {}
): Observable<GraphQLResult<GQL.IMutation>> {
    const subjectID = subject.id
    if (!subjectID) {
        throw new Error('subject has no id')
    }
    return configurationCascade.pipe(
        take(1),
        mergeMap(config => {
            const subjectConfig = config.subjects.find(s => s.id === subjectID)
            if (!subjectConfig) {
                throw new Error(`no configuration subject: ${subjectID}`)
            }
            const lastID = subjectConfig.latestSettings ? subjectConfig.latestSettings.id : null

            return mutateGraphQL(mutation, { ...variables, subject: subjectID, lastID })
        }),
        map(result => {
            refreshConfiguration().subscribe()
            return result
        })
    )
}
