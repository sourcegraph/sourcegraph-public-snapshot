import React, { useCallback, useState } from 'react'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import { map } from 'rxjs/operators'
import { Form } from '../../../components/Form'
import {
    UpdateGraphResult,
    UpdateGraphVariables,
    DeleteGraphResult,
    DeleteGraphVariables,
} from '../../../graphql-operations'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { GraphFormFields, GraphFormValue } from '../form/GraphFormFields'
import { GraphSelectionProps } from '../selector/graphSelectionProps'

type FormValue = UpdateGraphVariables['input']

interface Props extends Pick<GraphSelectionProps, 'reloadGraphs'> {
    initialValue: FormValue

    /** Called when the graph is successfully updated. */
    onUpdate: (graph: Pick<GQL.IGraph, 'url'>) => void

    /** Called when the graph is successfully deleted. */
    onDelete: () => void
}

export const EditGraphForm: React.FunctionComponent<Props> = ({
    initialValue,
    onUpdate: parentOnUpdate,
    onDelete: parentOnDelete,
    reloadGraphs,
}) => {
    const [value, setValue] = useState<FormValue>(initialValue)
    const onChange = useCallback((newValue: GraphFormValue) => setValue(previous => ({ ...previous, ...newValue })), [])

    const onUpdate = useCallback<typeof parentOnUpdate>(graph => {
        reloadGraphs()
        parentOnUpdate(graph)
    }, [])

    const [opState, setOpState] = useState<boolean | Error>(false)
    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setOpState(true)
            try {
                const graph = await requestGraphQL<UpdateGraphResult, UpdateGraphVariables>(
                    gql`
                        mutation UpdateGraph($input: UpdateGraphInput!) {
                            updateGraph(input: $input) {
                                url
                            }
                        }
                    `,
                    { input: value }
                )
                    .pipe(
                        map(dataOrThrowErrors),
                        map(data => data.updateGraph)
                    )
                    .toPromise()
                onUpdate(graph)
            } catch (error) {
                setOpState(error)
            }
        },
        [onUpdate, value]
    )

    const onDelete = useCallback<typeof parentOnDelete>(() => {
        reloadGraphs()
        parentOnDelete()
    }, [])

    const onDeleteClick = useCallback<React.MouseEventHandler>(async event => {
        event.preventDefault()

        setOpState(true)

        if (!confirm('Really delete this graph?')) {
            setOpState(false)
            return
        }

        try {
            await requestGraphQL<DeleteGraphResult, DeleteGraphVariables>(
                gql`
                    mutation DeleteGraph($graph: ID!) {
                        deleteGraph(graph: $graph) {
                            alwaysNil
                        }
                    }
                `,
                { graph: initialValue.id }
            )
                .pipe(map(dataOrThrowErrors))
                .toPromise()
            onDelete()
        } catch (error) {
            setOpState(error)
        }
    }, [])

    return (
        <Form className="w-100" onSubmit={onSubmit}>
            <GraphFormFields value={value} onChange={onChange} />
            <button type="submit" className="btn btn-primary" disabled={opState === true}>
                Save
            </button>
            {isErrorLike(opState) && <div className="mt-3 alert alert-danger">Error: {opState.message}</div>}
            <hr className="my-4" />
            <button type="button" className="btn btn-danger" disabled={opState === true} onClick={onDeleteClick}>
                Delete graph
            </button>
        </Form>
    )
}
