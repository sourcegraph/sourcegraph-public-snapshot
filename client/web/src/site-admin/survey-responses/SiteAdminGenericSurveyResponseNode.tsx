import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { SurveyResponseFields } from '../../graphql-operations'
import { userURL } from '../../user'

import { SurveyQuestionResponses } from './SiteAdminSurveyQuestionResponses'
import { ScoreBadge } from './SiteAdminSurveyScoreBadge'

interface SurveyResponseNodeProps {
    /**
     * The survey response to display in this list item.
     */
    node: SurveyResponseFields
}

export const GenericSurveyResponseNode: React.FunctionComponent<SurveyResponseNodeProps> = ({ node }) => (
    <li className="list-group-item py-2">
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <strong>
                    {node.user ? (
                        <Link to={userURL(node.user.username)}>{node.user.username}</Link>
                    ) : node.email ? (
                        node.email
                    ) : (
                        'anonymous user'
                    )}
                </strong>
                <ScoreBadge score={node.score} />
            </div>
            <div>
                <Timestamp date={node.createdAt} />
            </div>
        </div>
        <SurveyQuestionResponses node={node} />
    </li>
)
