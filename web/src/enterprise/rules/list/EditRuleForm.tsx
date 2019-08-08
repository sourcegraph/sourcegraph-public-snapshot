import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { RuleForm, RuleFormData } from './RuleForm'

const updateRule = (input: GQL.IUpdateRuleInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateRule($input: UpdateRuleInput!) {
                updateRule(input: $input) {
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
    rule: Pick<GQL.IRule, 'id'> & RuleFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the rule is updated successfully. */
    onRuleUpdate: () => void

    className?: string
}

/**
 * A form to update a rule.
 */
export const UpdateRuleForm: React.FunctionComponent<Props> = ({ rule, onDismiss, onRuleUpdate, className = '' }) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, description, definition }: RuleFormData) => {
            setIsLoading(true)
            try {
                await updateRule({ id: rule.id, name, description, definition })
                setIsLoading(false)
                onDismiss()
                onRuleUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [rule.id, onDismiss, onRuleUpdate]
    )

    return (
        <RuleForm
            initialValue={rule}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
