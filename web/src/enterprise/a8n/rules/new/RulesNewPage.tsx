import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { Form } from '../../../../components/Form'
import { ModalPage } from '../../../../components/ModalPage'
import { PageTitle } from '../../../../components/PageTitle'
import { RuleNameFormGroup } from '../form/RuleNameFormGroup'
import { RulesAreaContext } from '../scope/ScopeRulesArea'

interface Props extends Pick<RulesAreaContext, 'scope' | 'setBreadcrumbItem'> {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new rule.
 */
export const RulesNewPage: React.FunctionComponent<Props> = ({ scope, setBreadcrumbItem }) => {
    useEffect(() => {
        setBreadcrumbItem({ text: 'New' })
        return () => setBreadcrumbItem(undefined)
    }, [setBreadcrumbItem])

    const [name, setName] = useState('')
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setName(e.currentTarget.value),
        []
    )

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | Pick<GQL.IRule, 'url'> | ErrorLike>(
        null
    )
    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setCreationOrError(LOADING)
            try {
                const input: GQL.ICreateRuleInput = {
                    project: scope.id,
                    name,
                }
                setCreationOrError(
                    await mutateGraphQL(
                        gql`
                            mutation CreateRule($input: CreateRuleInput!) {
                                rules {
                                    createRule(input: $input) {
                                        url
                                    }
                                }
                            }
                        `,
                        { input }
                    )
                        .pipe(
                            map(dataOrThrowErrors),
                            map(data => data.rules.createRule)
                        )
                        .toPromise()
                )
            } catch (err) {
                setCreationOrError(asError(err))
            }
        },
        [name, scope.id]
    )

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && !isErrorLike(creationOrError) && (
                <Redirect to={creationOrError.url} />
            )}
            <PageTitle title="New rule" />
            <ModalPage>
                <h2>New rule</h2>
                <Form onSubmit={onSubmit}>
                    <RuleNameFormGroup value={name} onChange={onNameChange} disabled={creationOrError === LOADING} />
                    <button type="submit" disabled={creationOrError === LOADING} className="btn btn-primary">
                        {creationOrError === LOADING ? (
                            <LoadingSpinner className="icon-inline" />
                        ) : (
                            <AddIcon className="icon-inline" />
                        )}{' '}
                        Create rule
                    </button>
                </Form>
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
