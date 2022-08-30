import { SurveySubmissionInput } from '../graphql-operations'

type SurveyFields = keyof SurveySubmissionInput

export const NPS_QUESTIONS = new Map<SurveyFields, string>([
    ['score', 'How likely is it that you would recommend Sourcegraph to a friend?'],
    ['useCases', 'You are using sourcegraph to...'],
    ['otherUseCase', 'What else are you using Sourcegraph to do?'],
    ['additionalInformation', ' Anything else you would like to share with us?'],
    ['email', 'What is your email?'],
])

/**
 * These questions are no longer directly used in the NPS survey.
 * They are preserved as we still show them on the site admin "Survey responses" page
 */
export const LEGACY_NPS_QUESTIONS: Record<string, string> = {
    reason: 'What is the most important reason for the score you gave Sourcegraph?',
    better: 'What could Sourcegraph do to provide a better product?',
} as const

Array.from(NPS_QUESTIONS).map(([key, value]) => ({ key, value }))
