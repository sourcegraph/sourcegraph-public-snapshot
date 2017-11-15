import * as React from 'react'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { ResultContainer } from './ResultContainer'

interface Props {
    /**
     * The repository of this search result.
     */
    repoPath: string

    /**
     * The icon to show left to the title.
     */
    icon: React.ComponentType<{ className: string }>
}

export const RepoSearchResult: React.StatelessComponent<Props> = (props: Props) => (
    <ResultContainer collapsible={false} icon={props.icon} title={<RepoBreadcrumb repoPath={props.repoPath} />} />
)
