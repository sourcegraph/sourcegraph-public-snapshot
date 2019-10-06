import H from 'history'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
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
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerProps {
    rule: Pick<GQL.IRule, 'id'> & RuleFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the rule is updated successfully. */
    onRuleUpdate: () => void

    className?: string
    history: H.History
}

/**
 * A form to edit a rule.
 */
export const EditRuleForm: React.FunctionComponent<Props> = ({
    rule,
    onDismiss,
    onRuleUpdate,
    className = '',
    ...props
}) => {
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
            {...props}
            header={<h2>Edit rule</h2>}
            initialValue={rule}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
