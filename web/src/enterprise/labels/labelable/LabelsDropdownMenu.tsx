import React from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { useCampaigns } from '../list/useCampaigns'

interface Props {
    repository: Pick<GQL.IRepository, 'id'>

    /** Called when the user selects a label in the menu. */
    onSelect: (label: Pick<GQL.ILabel, 'id'>) => void
}

const LOADING = 'loading' as const

/**
 * A dropdown menu with a list of labels.
 */
export const CampaignsDropdownMenu: React.FunctionComponent<Props> = ({ repository, onSelect, ...props }) => {
    const labels = useLabels(repository)
    return (
        <DropdownMenu {...props}>
            {labels === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading labels...
                </DropdownItem>
            ) : isErrorLike(labels) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading labels
                </DropdownItem>
            ) : (
                labels.nodes.map(label => (
                    // tslint:disable-next-line: jsx-no-lambda
                    <DropdownItem key={label.id} onClick={() => onSelect(label)}>
                        <small className="text-muted">#{label.namespace.namespaceName}</small> {label.name}
                    </DropdownItem>
                ))
            )}
        </DropdownMenu>
    )
}
