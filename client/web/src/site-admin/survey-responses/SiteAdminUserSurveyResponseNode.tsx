import React, { useState } from 'react'

import classNames from 'classnames'

import { Button, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { UserWithSurveyResponseFields } from '../../graphql-operations'
import { userURL } from '../../user'

import { SurveyQuestionResponses } from './SiteAdminSurveyQuestionResponses'
import { ScoreBadge } from './SiteAdminSurveyScoreBadge'

import styles from './SiteAdminSurveyResponsesPage.module.scss'

interface UserSurveyResponseNodeProps {
    /**
     * The survey response to display in this list item.
     */
    node: UserWithSurveyResponseFields
}

export const UserSurveyResponseNode: React.FunctionComponent<UserSurveyResponseNodeProps> = ({ node }) => {
    const [displayAll, setDisplayAll] = useState(false)

    const showMoreClicked = (): void => setDisplayAll(previous => !previous)

    const responses = node.surveyResponses

    return (
        <>
            <tr>
                <td>
                    <strong>
                        <Link to={userURL(node.username)}>{node.username}</Link>
                    </strong>
                </td>
                <td>
                    {node.usageStatistics?.lastActiveTime ? (
                        <Timestamp date={node.usageStatistics.lastActiveTime} />
                    ) : (
                        '?'
                    )}
                </td>
                <td>
                    {responses && responses.length > 0 ? (
                        <>
                            <Timestamp date={responses[0].createdAt} />
                            <ScoreBadge score={responses[0].score} />
                        </>
                    ) : (
                        <>No responses</>
                    )}
                </td>
                <td>
                    {responses.length > 0 && (
                        <Button onClick={showMoreClicked} variant="secondary" size="sm">
                            {displayAll ? 'Hide' : 'See all'}
                        </Button>
                    )}
                </td>
            </tr>
            {displayAll && (
                <tr>
                    <td colSpan={4}>
                        {responses.map((response, index) => (
                            <dl key={index}>
                                <div className={classNames('pl-3 border-left', styles.wideBorder)}>
                                    <Timestamp date={response.createdAt} />
                                    <ScoreBadge score={response.score} />
                                    <br />
                                    <SurveyQuestionResponses node={response} />
                                </div>
                            </dl>
                        ))}
                    </td>
                </tr>
            )}
        </>
    )
}
