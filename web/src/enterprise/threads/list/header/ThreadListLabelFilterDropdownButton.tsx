import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import {
    ConnectionListFilterContext,
    ConnectionListFilterDropdownButton,
    ConnectionListFilterItem,
} from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import { Label } from '../../../../components/Label'

interface Props extends ConnectionListFilterContext<GQL.IThreadConnectionFilters> {}

const ITEM_FUNC = (f: GQL.ILabelFilter): ConnectionListFilterItem => ({
    ...f,
    beforeText: <Label label={f.label || { name: '', color: '#cccccc' }} colorOnly={true} className="mr-3" />,
    text: f.labelName,
    queryPart: `label:${f.labelName}`,
})

export const ThreadListLabelFilterDropdownButton: React.FunctionComponent<Props> = props => (
    <ConnectionListFilterDropdownButton<GQL.IThreadConnectionFilters, 'label'>
        {...props}
        filterKey="label"
        itemFunc={ITEM_FUNC}
        buttonText="Labels"
        noun="label"
    />
)
