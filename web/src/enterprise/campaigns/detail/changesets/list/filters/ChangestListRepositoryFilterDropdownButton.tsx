import React from 'react'
import * as GQL from '../../../../../../../../shared/src/graphql/schema'
import {
    ConnectionListFilterContext,
    ConnectionListFilterDropdownButton,
    ConnectionListFilterItem,
} from '../../../../../../components/connectionList/ConnectionListFilterDropdownButton'

interface Props extends ConnectionListFilterContext<GQL.IChangesetConnectionFilters> {}

const ITEM_FUNC = (f: GQL.IRepositoryFilter): ConnectionListFilterItem => ({
    ...f,
    text: f.repository.name,
    queryPart: `repo:${f.repository.name}`,
})

// TODO!(sqs): use CheckableDropdownItem
export const ChangesetListRepositoryFilterDropdownButton: React.FunctionComponent<Props> = props => (
    <ConnectionListFilterDropdownButton<GQL.IChangesetConnectionFilters, 'repository'>
        {...props}
        filterKey="repository"
        itemFunc={ITEM_FUNC}
        buttonText="Repository"
        noun="repository"
        pluralNoun="repositories"
    />
)
