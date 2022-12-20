/**
 * These types should be defined in the `web` package but because of the
 * fact that temporary settings rely on the shared global type schema there's
 * no way to inject web specific types into the `useTemporarySetting` hook.
 *
 * We need to introduce an ability to inject app specific temporary settings
 * into the hook and split the global schema into respective packages.
 * https://github.com/sourcegraph/sourcegraph/issues/45836
 *
 */
import { Optional } from 'utility-types'

/**
 * Tour supported languages
 */
export enum TourLanguage {
    C = 'C',
    Go = 'Go',
    Java = 'Java',
    Javascript = 'JavaScript',
    Php = 'PHP',
    Python = 'Python',
    Typescript = 'TypeScript',
}

/**
 * Tour task
 */
export interface TourTaskType {
    title?: string
    dataAttributes?: {}
    icon?: React.ReactNode
    steps: TourTaskStepType[]
    /**
     * Completion percentage, 0-100. Dynamically calculated field
     */
    completed?: number
}

export interface TourTaskStepType {
    id: string
    label: string
    tooltip?: string
    action:
        | {
              type: 'video'
              value: string | Record<TourLanguage, string>
          }
        | {
              type: 'link' | 'new-tab-link'
              variant?: 'button-primary'
              value: string | Record<TourLanguage, string>
          }
        | {
              type: 'restart'
              value: string
          }
    /**
     * HTML string, which will be displayed in info box when navigating to a step link.
     */
    info?: JSX.Element
    /**
     * The step will be marked as completed only if one of the "completeAfterEvents" will be triggered
     */
    completeAfterEvents?: string[]
    /**
     * Dynamically calculated field
     */
    isCompleted?: boolean
}

export interface TourState {
    completedStepIds?: string[]
    status?: 'closed' | 'completed'
    language?: TourLanguage
}

export type TourListState = Optional<Record<string, TourState>>
