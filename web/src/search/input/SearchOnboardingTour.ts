/**
 * This file contains utility functions for the search onboarding tour.
 */
import Shepherd from 'shepherd.js'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'

export const HAS_CANCELLED_TOUR_KEY = 'has-cancelled-onboarding-tour'
export const HAS_SEEN_TOUR_KEY = 'has-seen-onboarding-tour'

/**
 * generateStep creates the content for tooltips for the search tour. All steps that just contain
 * simple text should use this function to populate the step's `text` field.
 */
export function generateStepTooltip(
    tour: Shepherd.Tour,
    title: string,
    stepNumber: number,
    description?: string,
    additionalContent?: HTMLElement
): HTMLElement {
    const element = document.createElement('div')
    element.className = `d-flex flex-column test-tour-step-${stepNumber}`
    const titleElement = document.createElement('h4')
    titleElement.textContent = title
    titleElement.className = 'font-weight-bold'
    element.append(titleElement)
    if (description) {
        const descriptionElement = document.createElement('h4')
        descriptionElement.textContent = description
        element.append(descriptionElement)
    }
    if (additionalContent) {
        const additionalContentContainer = document.createElement('div')
        additionalContentContainer.append(additionalContent)
        element.append(additionalContent)
    }
    const bottomRow = generateBottomRow(tour, stepNumber)
    element.append(bottomRow)
    return element
}

/**
 * Generates the bottom row of the tooltip, which shows the current step number and a "close tour" button.
 *
 * @param tour the tour instance.
 * @param stepNumber the step number.
 */
export function generateBottomRow(tour: Shepherd.Tour, stepNumber: number): HTMLElement {
    const stepNumberLabel = document.createElement('span')
    stepNumberLabel.className = 'font-weight-light font-italic'
    stepNumberLabel.textContent = `${stepNumber} of 5`

    const closeTourButton = document.createElement('button')
    closeTourButton.className = 'btn btn-link p-0'
    closeTourButton.textContent = 'Close tour'
    closeTourButton.addEventListener('click', () => {
        tour.cancel()
        localStorage.setItem(HAS_CANCELLED_TOUR_KEY, 'true')
    })

    const bottomRow = document.createElement('div')
    bottomRow.className = 'd-flex justify-content-between'
    bottomRow.append(stepNumberLabel)
    bottomRow.append(closeTourButton)
    return bottomRow
}

/**
 * Generates the tooltip content for the first step in the tour.
 *
 * @param languageButtonHandler the handler for the "search a language" button.
 * @param repositoryButtonHandler the handler for the "search a repository" button.
 */
export function createStep1Tooltip(
    tour: Shepherd.Tour,
    languageButtonHandler: () => void,
    repositoryButtonHandler: () => void
): HTMLElement {
    const list = document.createElement('ul')
    list.className = 'my-4 list-group'
    const languageListItem = document.createElement('li')
    languageListItem.className = 'list-group-item p-0 border-0 mb-2'
    languageListItem.textContent = '-'
    const languageButton = document.createElement('button')
    languageButton.className = 'btn btn-link p-0 pl-1 test-tour-language-button'
    languageButton.textContent = 'Search a language'
    languageListItem.append(languageButton)
    languageButton.addEventListener('click', () => {
        languageButtonHandler()
    })
    const repositoryListItem = document.createElement('li')
    repositoryListItem.className = 'list-group-item p-0 border-0 mb-2 test-tour-repo-button'
    repositoryListItem.textContent = '-'
    const repositoryButton = document.createElement('button')
    repositoryButton.className = 'btn btn-link p-0 pl-1'
    repositoryButton.textContent = 'Search a repository'
    repositoryButton.addEventListener('click', () => {
        repositoryButtonHandler()
    })
    repositoryListItem.append(repositoryButton)
    list.append(languageListItem)
    list.append(repositoryListItem)
    return generateStepTooltip(tour, 'Code search tour', 1, 'How would you like to begin?', list)
}

/**
 * Generates the tooltip content for the "add code" step in the repository path, which asks users to input their own terms into the query.
 *
 * @param tour the tour instance
 */
export function createAddCodeStepTooltip(tour: Shepherd.Tour): HTMLElement {
    return generateStepTooltip(
        tour,
        'Add code to your search',
        3,
        'Type the name of a function, variable or other code.'
    )
}

/**
 * A map containing the language filter and the example to be displayed
 * in the "add code to your query" tooltip.
 */
export const languageFilterToSearchExamples: { [key: string]: { query: string; patternType: SearchPatternType } } = {
    'lang:c': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:cpp': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:csharp': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:css': { query: 'body {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:go': { query: 'for {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:graphql': { query: 'Query {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:haskell': { query: 'if :[my_match] else', patternType: SearchPatternType.structural },
    'lang:html': { query: '<div class="panel">:[my_match]</div>', patternType: SearchPatternType.structural },
    'lang:java': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:javascript': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:json': { query: '"object":{:[my_match]}', patternType: SearchPatternType.structural },
    'lang:lua': { query: 'function update() :[my_match] end', patternType: SearchPatternType.structural },
    'lang:markdown': { query: '', patternType: SearchPatternType.structural },
    'lang:php': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:powershell': { query: 'try {:[my_match]}', patternType: SearchPatternType.structural },
    'lang:python': { query: 'try:[my_match] except', patternType: SearchPatternType.structural },
    'lang:r': { query: 'tryCatch( :[my_match] )', patternType: SearchPatternType.structural },
    'lang:ruby': { query: 'while :[my_match] end', patternType: SearchPatternType.structural },
    'lang:sass': { query: 'transition( :[my_match] )', patternType: SearchPatternType.structural },
    'lang:swift': { query: 'switch :[a]{:[b]}', patternType: SearchPatternType.structural },
    'lang:typescript': { query: 'try{:[my_match]}', patternType: SearchPatternType.structural },
}

/**
 * Generates the tooltip content for the "add code" step in the language path, which asks users to input their own terms into the query.
 * It provides an example based on the language they selected in the previous step.
 *
 * @param tour the tour instance.
 * @param languageQuery the current query including a `lang:` filter. Used for language queries so we know what examples to suggest.
 * @param exampleCallback the callback to be run when clicking the example query.
 */
export function createAddCodeStepWithLanguageExampleTooltip(
    tour: Shepherd.Tour,
    languageQuery: string,
    exampleCallback: (query: string, patternType: SearchPatternType) => void
): HTMLElement {
    const list = document.createElement('ul')
    list.className = 'my-4 list-group'

    const listItem = document.createElement('li')
    listItem.className = 'list-group-item p-0 border-0'
    listItem.textContent = '>'

    const exampleButton = document.createElement('button')
    exampleButton.className = 'btn btn-link test-tour-language-example'

    const langsList = languageFilterToSearchExamples
    let example = { query: '', patternType: SearchPatternType.literal }
    if (languageQuery && Object.keys(langsList).includes(languageQuery)) {
        example = langsList[languageQuery]
    }
    const codeElement = document.createElement('code')
    codeElement.textContent = example.query
    exampleButton.append(codeElement)

    exampleButton.addEventListener('click', () => {
        const fullQuery = [languageQuery, example.query].join(' ')
        exampleCallback(fullQuery, example.patternType)
        tour.show('view-search-reference')
    })
    listItem.append(exampleButton)
    list.append(listItem)
    return generateStepTooltip(
        tour,
        'Add code to your search',
        3,
        'Type the name of a function, variable or other code. Or try an example:',
        list
    )
}

export const isValidLangQuery = (query: string): boolean => Object.keys(languageFilterToSearchExamples).includes(query)

/** *
 * The types below allow us to end steps in the tour from components outside of the SearchPageInput component
 * where the tour is located. In particular, we want to advance tour steps when a user types or updates the query input
 * after a debounce period, on certain conditions such as the contents of the query.
 *
 * Steps that aren't included here use Shepherd's built-in `advanceOn` field to specify events to advance on.
 */

export interface AdvanceStepCallback {
    /**
     * The ID of the step to advance from.
     */
    stepToAdvance: string
    /**
     * Conditions that must be true before advancing to the next step.
     */
    queryConditions?: (query: string) => boolean
}

/**
 * Defines a callback to advance a step.
 */
type AdvanceStandardStep = AdvanceStepCallback & { handler: (tour: Shepherd.Tour) => void }

/**
 * A special case type to define a callback for a the "add code to your query" step on the language path.
 * The handler takes a query and setQueryHandler, which allows us to generate the appropriate tooltip
 * content for the next step.
 */
type AdvanceLanguageInputStep = AdvanceStepCallback & {
    handler: (
        tour: Shepherd.Tour,
        query: string,
        setQueryHandler: (query: string, patternType?: SearchPatternType) => void
    ) => void
}

export type CallbackToAdvanceTourStep = AdvanceStandardStep | AdvanceLanguageInputStep

/**
 * A list of callbacks that will advance certain steps when the query input's value is changed.
 */
export const stepCallbacks: CallbackToAdvanceTourStep[] = [
    {
        stepToAdvance: 'filter-repository',
        handler: (tour: Shepherd.Tour, query: string): void => {
            if (tour.getById('filter-repository').isOpen() && query.endsWith(' ')) {
                tour.show('add-query-term')
                tour.getById('add-query-term').updateStepOptions({ text: createAddCodeStepTooltip(tour) })
            }
        },
        queryConditions: (query: string): boolean => query !== 'repo:',
    },
    {
        stepToAdvance: 'filter-lang',
        handler: (
            tour: Shepherd.Tour,
            query: string,
            setQueryHandler: (query: string, patternType?: SearchPatternType) => void
        ): void => {
            if (tour.getById('filter-lang').isOpen()) {
                tour.show('add-query-term')
                tour.getById('add-query-term').updateStepOptions({
                    text: createAddCodeStepWithLanguageExampleTooltip(
                        tour,
                        query.trim() ?? '',
                        (newQuery: string, patternType: SearchPatternType) => setQueryHandler(newQuery, patternType)
                    ),
                })
            }
        },
        queryConditions: (query: string): boolean => query !== 'lang:' && isValidLangQuery(query.trim()),
    },
    {
        stepToAdvance: 'add-query-term',
        handler: (tour: Shepherd.Tour): void => {
            if (tour.getById('add-query-term').isOpen()) {
                tour.show('view-search-reference')
            }
        },
        queryConditions: (query: string): boolean => query !== 'repo:' && query !== 'lang:',
    },
]
