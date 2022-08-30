import React from 'react'

import { SurveyResponseFields } from '../../graphql-operations'
import { NPS_QUESTIONS } from '../../marketing/constants'

interface SurveyQuestionResponsesProps {
    /**
     * The survey response to display in this list item.
     */
    node: Partial<SurveyResponseFields>
}

export const SurveyQuestionResponses: React.FunctionComponent<SurveyQuestionResponsesProps> = ({ node }) => {
    const hasResponses =
        node.reason || node.better || node.additionalInformation || node.otherUseCase || node.useCases?.length

    if (!hasResponses) {
        return null
    }

    // Preserved from previous NPS survey
    const legacyResponses = (
        <>
            {node.reason && node.reason !== '' && (
                <>
                    <dt>What is the most important reason for the score you gave Sourcegraph?</dt>
                    <dd>{node.reason}</dd>
                </>
            )}
            {node.better && node.better !== '' && (
                <>
                    <dt>What could Sourcegraph do to provide a better product?</dt>
                    <dd>{node.better}</dd>
                </>
            )}
        </>
    )

    return (
        <dl className="mt-3">
            {Array.from(NPS_QUESTIONS).map(([key, value]) => (
                <>
                    <dt>{value}</dt>
                    <dd>{node[key]}</dd>
                </>
            ))}
            {legacyResponses}
        </dl>
    )
}
