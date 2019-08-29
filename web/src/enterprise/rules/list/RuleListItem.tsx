import PencilIcon from 'mdi-react/PencilIcon'
import H from 'history'
import React, { useState, useCallback, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { RuleDeleteButton } from '../components/RuleDeleteButton'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'

interface Props extends ExtensionsControllerProps {
    rule: GQL.IRule

    /** Called when the rule is updated. */
    onRuleUpdate: () => void

    isEditing: boolean
    stopEditingUrl: string

    className?: string
    history: H.History
}

const GenericRuleListItem: React.FunctionComponent<{
    headerLeft?: React.ReactFragment
    headerRight?: React.ReactFragment
    description?: React.ReactFragment
    className?: string
}> = ({ headerLeft, headerRight, description, className = '' }) => (
    <div className={`card ${className}`}>
        <header className="card-header d-flex align-items-center">
            {headerLeft}
            <div className="flex-1" />
            {headerRight}
        </header>
        {description}
    </div>
)

/**
 * An item in the list of rules.
 */
export const RuleListItem: React.FunctionComponent<Props> = ({
    rule,
    onRuleUpdate,
    isEditing,
    stopEditingUrl,
    className = '',
    history,
    ...props
}) => {
    useEffect(() => {
        if (isEditing) {
            return history.block('Discard rule edits?')
        }
        return undefined
    }, [history, isEditing])

    return !isEditing ? (
        <GenericRuleListItem
            headerLeft={<h3 className="mb-0">{rule.name}</h3>}
            headerRight={
                <>
                    <Link to={rule.url} className="btn btn-sm btn-secondary text-decoration-none mr-2">
                        <PencilIcon className="icon-inline" /> Edit
                    </Link>
                    <RuleDeleteButton
                        {...props}
                        rule={rule}
                        onDelete={onRuleUpdate}
                        buttonClassName="btn-sm btn-secondary"
                    />
                </>
            }
            description={
                rule.description ? (
                    <div className="card-body">
                        <Markdown dangerousInnerHTML={renderMarkdown(rule.description)} />
                    </div>
                ) : (
                    undefined
                )
            }
            className={className}
        />
    ) : (
        <GenericRuleListItem
            headerLeft={<h3 className="mb-0">{rule.name}</h3>}
            headerRight={
                <>
                    <Link to={stopEditingUrl} className="btn btn-sm btn-secondary text-decoration-none mr-2">
                        Cancel
                    </Link>
                </>
            }
            description={
                rule.description ? (
                    <div className="card-body">
                        <Markdown dangerousInnerHTML={renderMarkdown(rule.description)} />
                    </div>
                ) : (
                    undefined
                )
            }
            className={className}
        />
    )
}
