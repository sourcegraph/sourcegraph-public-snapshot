const CODE_SEARCH = 'Code search use cases'
const CODE_INTEL = 'The power of code intel'
const TOOLS = 'Tools to improve workflow'

export interface OnboardingTourStepItem {
    id: string
    /**
     * Title of the group which step belongs
     */
    group: string
    /**
     * Title/name of the step
     */
    title: string
    /**
     * URL to redirect
     */
    to: string
    /**
     * Flag whether this step was completed of not
     */
    isCompleted: boolean
    /**
     * HTML text to show on page after redirecting to link
     */
    info?: {
        selector: string
        content: string
    }
    /**
     * Log "${id}Completed" event and mark item as completed after one of the events is triggered
     */
    completeAfterEvents?: string[]
}

export const ONBOARDING_STEP_ITEMS: Omit<OnboardingTourStepItem, 'isCompleted'>[] = [
    // Group: CODE_SEARCH
    {
        id: 'TourDiffSearch',
        title: 'Find removed code in diffs',
        group: CODE_SEARCH,
        to:
            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:diff+lang:go+select:commit.diff.removed+magic&patternType=literal',
    },
    {
        id: 'TourCommitsSearch',
        title: 'Find changes in commits',
        group: CODE_SEARCH,
        to:
            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:commit+bump&patternType=literal',
    },
    {
        id: 'TourSymbolsSearch',
        title: 'Find symbols via a string',
        group: CODE_SEARCH,
        to: '/search?q=context:global+r:sourcegraph/sourcegraph%24+lang:go+type:symbol+auth&patternType=literal',
    },
    // Group: CODE_INTEL
    {
        id: 'TourFindReferences',
        title: 'Find references',
        group: CODE_INTEL,
        info: {
            selector: '.onboarding-tour-info-marker',
            content: `<strong>FIND REFERENCES</strong><br/>
            Hover over a token in the highlighted line to open code intel, then click ‘Find References’ to locate all calls of this code.`,
        },
        completeAfterEvents: ['findReferences'],
        to: '/github.com/sourcegraph/sourcegraph/-/blob/internal/featureflag/featureflag.go?L9:6',
    },
    {
        id: 'TourGoToDefinition',
        title: 'Go to a definition',
        group: CODE_INTEL,
        info: {
            selector: '.onboarding-tour-info-marker',
            content: `<strong>GO TO DEFINITION</strong><br/>
            Hover over a token in the highlighted line to open code intel, then click ‘Go to definition’ to locate a token definition.`,
        },
        completeAfterEvents: ['goToDefinition', 'goToDefinition.preloaded'],
        to: '/github.com/sourcegraph/sourcegraph/-/blob/internal/repos/observability.go?L192:22',
    },
    // Group: TOOLS
    {
        id: 'TourEditorExtensions',
        group: TOOLS,
        title: 'IDE extentions',
        to: 'https://docs.sourcegraph.com/integration/editor',
    },
    {
        id: 'TourBrowserExtensions',
        group: TOOLS,
        title: 'Browser extensions',
        to: 'https://docs.sourcegraph.com/integration/browser_extension',
    },
]
