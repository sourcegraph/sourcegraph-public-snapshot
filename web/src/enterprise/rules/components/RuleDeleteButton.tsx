import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const deleteRule = (args: GQL.IDeleteRuleOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation DeleteRule($rule: ID!) {
                deleteRule(rule: $rule) {
                    alwaysNil
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    rule: Pick<GQL.IRule, 'id'>
    onDelete: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that permanently deletes a rule.
 */
export const RuleDeleteButton: React.FunctionComponent<Props> = ({
    rule,
    onDelete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (!confirm('Really delete the rule?')) {
                return
            }
            setIsLoading(true)
            try {
                await deleteRule({ rule: rule.id })
                setIsLoading(false)
                onDelete()
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting rule: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, rule.id, onDelete]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <DeleteIcon className="icon-inline" />} Delete
        </button>
    )
}
