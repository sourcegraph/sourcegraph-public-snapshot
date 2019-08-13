import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import {
    ThreadListFilterContext,
    ThreadListFilterDropdownButton,
    ThreadListFilterItem,
} from './ThreadListFilterDropdownButton'

interface Props extends ThreadListFilterContext {}

const ITEM_FUNC = (f: GQL.IRepositoryFilter): ThreadListFilterItem => ({
    ...f,
    text: f.repository.name,
    queryPart: `repo:${f.repository.name}`,
})

export const ThreadListRepositoryFilterDropdownButton: React.FunctionComponent<Props> = props => (
    <ThreadListFilterDropdownButton
        {...props}
        filterKey="repository"
        itemFunc={ITEM_FUNC}
        buttonText="Repository"
        noun="repository"
        pluralNoun="repositories"
    />
)
