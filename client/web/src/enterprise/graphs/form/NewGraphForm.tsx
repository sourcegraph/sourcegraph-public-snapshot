import React, { useCallback, useState } from 'react'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import { map } from 'rxjs/operators'
import { Form } from '../../../components/Form'
import { CreateGraphResult, CreateGraphVariables } from '../../../graphql-operations'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { GraphFormFields, GraphFormValue } from './GraphFormFields'
import { GraphSelectionProps } from '../selector/graphSelectionProps'

type FormValue = CreateGraphVariables['input']

interface Props extends Pick<GraphSelectionProps, 'reloadGraphs'> {
    initialValue: FormValue

    /** Called when the graph is successfully created. */
    onCreate: (graph: Pick<GQL.IGraph, 'url'>) => void
}

export const NewGraphForm: React.FunctionComponent<Props> = ({
    initialValue,
    onCreate: parentOnCreate,
    reloadGraphs,
}) => {
    const [value, setValue] = useState<FormValue>(initialValue)
    const onChange = useCallback((newValue: GraphFormValue) => setValue(previous => ({ ...previous, ...newValue })), [])

    const onCreate = useCallback<typeof parentOnCreate>(graph => {
        reloadGraphs()
        parentOnCreate(graph)
    }, [])

    const [opState, setOpState] = useState<boolean | Error>(false)
    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            setOpState(true)
            try {
                const graph = await requestGraphQL<CreateGraphResult, CreateGraphVariables>(
                    gql`
                        mutation CreateGraph($input: CreateGraphInput!) {
                            createGraph(input: $input) {
                                url
                            }
                        }
                    `,
                    { input: value }
                )
                    .pipe(
                        map(dataOrThrowErrors),
                        map(data => data.createGraph)
                    )
                    .toPromise()
                onCreate(graph)
            } catch (error) {
                setOpState(error)
            }
        },
        [onCreate, value]
    )

    return (
        <Form className="w-100" onSubmit={onSubmit}>
            <GraphFormFields value={value} onChange={onChange} />
            <button type="submit" className="btn btn-primary" disabled={opState === true}>
                Create
            </button>
            {isErrorLike(opState) && <div className="mt-3 alert alert-danger">Error: {opState.message}</div>}
        </Form>
    )
}
