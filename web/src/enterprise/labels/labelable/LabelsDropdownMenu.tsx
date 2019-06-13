import React from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { Label } from '../../../components/Label'
import { useLabels } from '../useLabels'

interface Props {
    repository: Pick<GQL.IRepository, 'id'>

    /** Called when the user selects a label in the menu. */
    onSelect: (label: Pick<GQL.ILabel, 'id'>) => void
}

const LOADING = 'loading' as const

/**
 * A dropdown menu with a list of labels.
 */
export const LabelsDropdownMenu: React.FunctionComponent<Props> = ({ repository, onSelect, ...props }) => {
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
            ) : labels.nodes.length === 0 ? (
                <DropdownItem header={true}>No labels defined in repository</DropdownItem>
            ) : (
                labels.nodes.map(label => (
                    // tslint:disable-next-line: jsx-no-lambda
                    <DropdownItem key={label.id} onClick={() => onSelect(label)} className="d-flex align-items-stretch">
                        <Label label={label} colorOnly={true} className="mr-3" /> {label.name}
                    </DropdownItem>
                ))
            )}
        </DropdownMenu>
    )
}
