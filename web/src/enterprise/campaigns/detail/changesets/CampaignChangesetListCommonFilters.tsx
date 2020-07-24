import React from 'react'
import { ConnectionListFilterContext } from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetListRepositoryFilterDropdownButton } from './list/filters/ChangestListRepositoryFilterDropdownButton'
import { ChangesetListLabelFilterDropdownButton } from './list/filters/ChangesetListLabelFilterDropdownButton'
import { ChangesetListFilterDropdownButton } from './list/filters/ChangesetListFilterDropdownButton'

export const ChangesetListHeaderCommonFilters: React.FunctionComponent<ConnectionListFilterContext<
    GQL.IChangesetConnectionFilters
>> = props => (
    <>
        <ChangesetListRepositoryFilterDropdownButton {...props} />
        <ChangesetListLabelFilterDropdownButton
            {...props}
            connection={{
                filters: {
                    label: [
                        { labelName: 'payment', label: { text: 'payment', color: 'blue' } },
                        { labelName: 'debt', label: { text: 'debt', color: 'green' } },
                        { labelName: 'webapp', label: { text: 'webapp', color: 'orange' } },
                    ],
                } as GQL.IChangesetConnectionFilters,
            }}
        />
        <ChangesetListFilterDropdownButton
            {...props}
            buttonText="Reviews"
            headerText="Filter by review state"
            items={[
                {
                    label: 'Approved',
                    queryField: 'review',
                    queryValues: [GQL.ChangesetReviewState.APPROVED.toLowerCase()],
                },
                {
                    label: 'Changes requested',
                    queryField: 'review',
                    queryValues: [GQL.ChangesetReviewState.CHANGES_REQUESTED.toLowerCase()],
                },
                {
                    label: 'Commented',
                    queryField: 'review',
                    queryValues: [GQL.ChangesetReviewState.COMMENTED.toLowerCase()],
                },
                {
                    label: 'Dismissed',
                    queryField: 'review',
                    queryValues: [GQL.ChangesetReviewState.DISMISSED.toLowerCase()],
                },
                {
                    label: 'Pending',
                    queryField: 'review',
                    queryValues: [GQL.ChangesetReviewState.PENDING.toLowerCase()],
                },
            ]}
        />
        <ChangesetListFilterDropdownButton
            {...props}
            buttonText="Checks"
            headerText="Filter by check state"
            items={[
                {
                    label: 'Passed',
                    queryField: 'checks',
                    queryValues: [GQL.ChangesetCheckState.PASSED.toLowerCase()],
                },
                {
                    label: 'Failed',
                    queryField: 'checks',
                    queryValues: [GQL.ChangesetCheckState.FAILED.toLowerCase()],
                },
                {
                    label: 'Pending',
                    queryField: 'checks',
                    queryValues: [GQL.ChangesetCheckState.PENDING.toLowerCase()],
                },
            ]}
        />
    </>
)
