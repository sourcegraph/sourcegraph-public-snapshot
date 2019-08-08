import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { RuleForm, RuleFormData } from './RuleForm'

const createRule = (input: GQL.ICreateRuleInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation CreateRule($input: CreateRuleInput!) {
                createRule(input: $input) {
                    id
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props {
    container: Pick<GQL.RuleContainer, 'id'>

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the rule is created successfully. */
    onRuleCreate: () => void

    className?: string
}

/**
 * A form to create a new rule.
 */
export const NewRuleForm: React.FunctionComponent<Props> = ({ container, onDismiss, onRuleCreate, className = '' }) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, description, definition }: RuleFormData) => {
            setIsLoading(true)
            try {
                await createRule({ container: container.id, name, description, definition })
                setIsLoading(false)
                onDismiss()
                onRuleCreate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [container, onDismiss, onRuleCreate]
    )

    return (
        <RuleForm
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Create rule"
            isLoading={isLoading}
            className={className}
        />
    )
}
