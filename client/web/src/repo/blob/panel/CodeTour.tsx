import classNames from 'classnames'
import Check from 'mdi-react/CheckCircleIcon'
import Delete from 'mdi-react/DeleteCircleIcon'
import Edit from 'mdi-react/EditIcon'
import React, { useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '@sourcegraph/wildcard'

import { useCodeTour } from './useCodeTour'

export const CodeTour: React.FunctionComponent = () => {
    const { routesList } = useCodeTour()
    return (
        <table className="table">
            <tbody>
                <tr>
                    <th>Path</th>
                    <th>Lines number</th>
                    <th>Description</th>
                    <th>Actions</th>
                </tr>
                {routesList.map(value => (
                    <ContentEditable
                        key={value.index}
                        index={value.index}
                        lineNumber={value.lineNumber}
                        description={value.description}
                    />
                ))}
            </tbody>
        </table>
    )
}

const ContentEditable = ({
    index,
    lineNumber,
    description,
}: {
    index: string
    lineNumber: string
    description: string
}): JSX.Element => {
    const { removeRoute, currentRoute, updateDescription } = useCodeTour()
    const [editDescription, setEditDescription] = useState(false)
    const [descriptionContent, setDescriptionContent] = useState('')

    const saveDescription = (): void => {
        setEditDescription(false)
        updateDescription(index, descriptionContent)
    }

    return (
        <tr className={classNames({ 'bg-2': currentRoute === index })}>
            <td className="align-middle">
                <Link to={`${index}#tab=codetour`}>{index}</Link>
            </td>
            <td>{lineNumber}</td>
            <td>
                {editDescription ? (
                    <textarea onChange={e => setDescriptionContent(e.target.value)} defaultValue={description} />
                ) : (
                    description
                )}
            </td>
            <td>
                <Button data-tooltip="Delete line" onClick={() => removeRoute(index)}>
                    <Delete />
                </Button>
                {editDescription ? (
                    <Button data-tooltip="Save description" onClick={() => saveDescription()}>
                        <Check />
                    </Button>
                ) : (
                    <Button>
                        <Edit data-tooltip="Edit description" onClick={() => setEditDescription(true)} />
                    </Button>
                )}
            </td>
        </tr>
    )
}
