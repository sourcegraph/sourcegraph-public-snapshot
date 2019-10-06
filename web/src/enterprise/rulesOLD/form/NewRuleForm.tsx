import H from 'history'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
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
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerProps {
    container: Pick<GQL.RuleContainer, 'id'>

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the rule is created successfully. */
    onRuleCreate: () => void

    className?: string
    history: H.History
}

/**
 * A form to create a new rule.
 */
export const NewRuleForm: React.FunctionComponent<Props> = ({
    container,
    onDismiss,
    onRuleCreate,
    className = '',
    ...props
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, description, definition }: RuleFormData) => {
            setIsLoading(true)
            try {
                await createRule({ container: container.id, rule: { name, description, definition } })
                setIsLoading(false)
                onDismiss()
                onRuleCreate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [container.id, onDismiss, onRuleCreate]
    )

    return (
        <RuleForm
            {...props}
            header={<h2>New rule</h2>}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Create rule"
            isLoading={isLoading}
            className={className}
        />
    )
}
