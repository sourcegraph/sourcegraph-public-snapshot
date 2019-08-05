import { queryAndFragmentForUnion } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'

const DEFAULT_FIELDS: (keyof GQL.ThreadOrIssueOrChangeset)[] = ['__typename', 'id', 'number', 'title', 'url', 'state']

const DEFAULT_NESTED_FIELDS = ['repository { name }']

const TYPE_NAMES: GQL.ThreadOrIssueOrChangeset['__typename'][] = ['Thread', 'Changeset']

export const queryAndFragmentForThreadOrIssueOrChangeset = (
    fields: (keyof GQL.ThreadOrIssueOrChangeset)[] = DEFAULT_FIELDS,
    nestedFields: string[] = DEFAULT_NESTED_FIELDS
) => queryAndFragmentForUnion(TYPE_NAMES, fields, nestedFields)
