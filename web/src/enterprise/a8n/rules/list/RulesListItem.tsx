import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { RuleStateIcon } from '../components/RuleStateIcon'
import { RulesAreaContext } from '../scope/ScopeRulesArea'
import { urlToRule } from '../url'

interface Props extends Pick<RulesAreaContext, 'rulesURL'>, ExtensionsControllerProps, PlatformContextProps {
    rule: GQL.IRule

    tag: 'li'
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A single rule in a list.
 */
export const RulesListItem: React.FunctionComponent<Props> = ({ rule, tag: Tag, rulesURL, className = '' }) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <h3 className="mb-0 font-weight-normal font-size-base d-flex align-items-center">
            <Link to={urlToRule(rulesURL, rule)} className="stretched-link">
                {rule.name}
            </Link>
        </h3>
        {rule.error && (
            <p className="text-danger" style={{ whiteSpace: 'normal' }}>
                {rule.error.message}
            </p>
        )}
    </Tag>
)
