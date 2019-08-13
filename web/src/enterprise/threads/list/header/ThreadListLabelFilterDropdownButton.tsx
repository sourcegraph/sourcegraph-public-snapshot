import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Label } from '../../../../components/Label'
import {
    ThreadListFilterContext,
    ThreadListFilterDropdownButton,
    ThreadListFilterItem,
} from './ThreadListFilterDropdownButton'

interface Props extends ThreadListFilterContext {}

const ITEM_FUNC = (f: GQL.ILabelFilter): ThreadListFilterItem => ({
    ...f,
    beforeText: <Label label={f.label || { name: '', color: '#cccccc' }} colorOnly={true} className="mr-3" />,
    text: f.labelName,
    queryPart: `label:${f.labelName}`,
})

export const ThreadListLabelFilterDropdownButton: React.FunctionComponent<Props> = props => (
    <ThreadListFilterDropdownButton
        {...props}
        filterKey="label"
        itemFunc={ITEM_FUNC}
        buttonText="Labels"
        noun="label"
    />
)
