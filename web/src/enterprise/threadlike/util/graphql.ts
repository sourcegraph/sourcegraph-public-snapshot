import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'

const threadOrIssueOrChangesetFields: (keyof GQL.ThreadOrIssueOrChangeset)[] = [
    '__typename',
    'id',
    'number',
    'title',
    'url',
    'externalURL',
    'status',
]
const threadOrIssueOrChangesetTypeNames: GQL.ThreadOrIssueOrChangeset['__typename'][] = ['Thread', 'Changeset']

export const threadOrIssueOrChangesetFieldsFragment = gql`
    ${threadOrIssueOrChangesetTypeNames.map(
        typeName =>
            `fragment ${typeName}Fields on ${typeName} { ${threadOrIssueOrChangesetFields.join(
                '\n'
            )} repository { name } }`
    )}
`

export const threadOrIssueOrChangesetFieldsQuery = gql`
__typename
${threadOrIssueOrChangesetTypeNames.map(typeName => `... on ${typeName} { ...${typeName}Fields }`).join('\n')}
`
