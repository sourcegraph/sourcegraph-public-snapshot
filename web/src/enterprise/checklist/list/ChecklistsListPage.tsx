import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { ChecklistsAreaTitle } from '../components/ChecklistsAreaTitle'
import { ChecklistsAreaContext } from '../global/ChecklistsArea'
import { ChecklistsList } from './ChecklistsList'

interface Props extends ChecklistsAreaContext, RouteComponentProps<{}> {}

/**
 * The checklists list page.
 */
export const ChecklistsListPage: React.FunctionComponent<Props> = ({ match, ...props }) => (
    <div className="w-100 mt-3">
        <PageTitle title="Checklists" />
        <div className="container-fluid">
            <ChecklistsAreaTitle />
        </div>
        <WithQueryParameter {...props}>
            {({ query, onQueryChange }) => (
                <ChecklistsList {...props} query={query} onQueryChange={onQueryChange} itemClassName="container-fluid" />
            )}
        </WithQueryParameter>
    </div>
)
