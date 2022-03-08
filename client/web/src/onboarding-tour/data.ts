import { OnboardingTourLanguage } from '../stores/onboardingTourState'

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
    url: string | Record<OnboardingTourLanguage, string>
    /**
     * Flag whether this step was completed of not
     */
    isCompleted: boolean
    /**
     * HTML text to show on page after redirecting to link
     */
    info?: string
    /**
     * Log "${id}Completed" event and mark item as completed after one of the events is triggered
     */
    completeAfterEvents?: string[]
}

export const ONBOARDING_STEP_ITEMS: Omit<OnboardingTourStepItem, 'isCompleted'>[] = [
    // Group: CODE_SEARCH
    {
        id: 'TourSymbolsSearch',
        title: 'Search multiple repos',
        group: CODE_SEARCH,
        url: {
            [OnboardingTourLanguage.C]:
                '/search?q=context:global+repo:torvalds/.*+lang:c+-file:.*/testing+magic&patternType=literal',
            [OnboardingTourLanguage.Go]:
                '/search?q=context:global+r:google/+lang:go+-file:test+//+TODO&patternType=literal',
            [OnboardingTourLanguage.Java]:
                '/search?q=context:global+r:github.com/square/+lang:java+-file:test+GiftCard&patternType=literal',
            [OnboardingTourLanguage.Javascript]:
                '/search?q=context:global+r:react+lang:JavaScript+-file:test+createPortal&patternType=literal',
            [OnboardingTourLanguage.Php]:
                '/search?q=context:global+repo:laravel+lang:php+-file:test+login%28&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Python]:
                '/search?q=context:global+r:aws/+lang:python+file:mock+def+test_patch&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Typescript]:
                '/search?q=context:global+r:react+lang:typescript+-file:test+createPortal%28&patternType=regexp&case=yes',
        },
        info: `<strong>Reference code in multiple repositories</strong><br/>
            The repo: query allows searching in multiple repositories matching a term. Use it to reference all of your projects or find open source examples.`,
    },
    {
        id: 'TourCommitsSearch',
        title: 'Find changes in commits',
        group: CODE_SEARCH,
        url: {
            [OnboardingTourLanguage.C]:
                '/search?q=context:global+repo:%5Egithub%5C.com/curl/doh%24+file:%5Edoh%5C.c+type:commit+option&patternType=literal',
            [OnboardingTourLanguage.Go]:
                '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:commit+bump&patternType=literal',
            [OnboardingTourLanguage.Java]:
                '/search?q=context:global+repo:sourcegraph-testing/.*+type:commit+lang:java+version&patternType=literal',
            [OnboardingTourLanguage.Javascript]:
                '/search?q=context:global+repo:%5Egithub%5C.com/hakimel/reveal%5C.js%24+type:commit+error&patternType=literal',
            [OnboardingTourLanguage.Php]:
                '/search?q=context:global+repo:square/connect-api-examples+type:commit+version&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Python]:
                '/search?q=context:global+r:elastic/elasticsearch+lang:python+type:commit+request_timeout&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Typescript]:
                '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:commit+bump&patternType=literal',
        },
        info: `<strong>Find changes in commits</strong><br/>
            Quickly find commits in history, then browse code from the commit, without checking out the branch.`,
    },
    {
        id: 'TourDiffSearch',
        title: 'Search diffs for removed code',
        group: CODE_SEARCH,
        url: {
            [OnboardingTourLanguage.C]:
                '/search?q=context:global+repo:curl/doh+type:diff+select:commit.diff.removed+mode&patternType=literal',
            [OnboardingTourLanguage.Go]:
                '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:diff+lang:go+select:commit.diff.removed+NameSpaceOrgId&patternType=literal',
            [OnboardingTourLanguage.Java]:
                '/search?q=context:global+repo:sourcegraph-testing/sg-hadoop+lang:java+type:diff+select:commit.diff.removed+getConf&patternType=literal',
            [OnboardingTourLanguage.Javascript]:
                '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:javascript+-file:test+type:diff+select:commit.diff.removed+promise&patternType=literal',
            [OnboardingTourLanguage.Php]:
                '/search?q=context:global+repo:laravel/laravel.*+lang:php++type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Python]:
                '/search?q=context:global+repo:pallets/+lang:python+type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
            [OnboardingTourLanguage.Typescript]:
                '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:typescript+type:diff+select:commit.diff.removed+authenticatedUser&patternType=regexp&case=yes',
        },
        info: `<strong>Searching diffs for removed code</strong><br/>
            Find removed code without browsing through history or trying to remember which file it was in.`,
    },
    // Group: CODE_INTEL
    // TODO:
    {
        id: 'TourFindReferences',
        title: 'Find references',
        group: CODE_INTEL,
        info: `<strong>FIND REFERENCES</strong><br/>
            Hover over a token in the highlighted line to open code intel, then click ‘Find References’ to locate all calls of this code.`,
        completeAfterEvents: ['findReferences'],
        url: {
            [OnboardingTourLanguage.C]: '/github.com/torvalds/linux/-/blob/arch/arm/kernel/atags_compat.c?L196:8',
            [OnboardingTourLanguage.Go]:
                '/github.com/sourcegraph/sourcegraph/-/blob/internal/featureflag/featureflag.go?L9:6',
            [OnboardingTourLanguage.Java]:
                '/github.com/square/okhttp/-/blob/samples/guide/src/main/java/okhttp3/recipes/PrintEvents.java?L126:27',
            [OnboardingTourLanguage.Javascript]:
                '/github.com/mozilla/pdf.js/-/blob/src/display/display_utils.js?L261:16',
            [OnboardingTourLanguage.Php]:
                '/github.com/square/connect-api-examples/-/blob/connect-examples/v1/php/payments-report.php?L164:32',
            [OnboardingTourLanguage.Python]: '/github.com/google-research/bert/-/blob/extract_features.py?L152:7',
            [OnboardingTourLanguage.Typescript]:
                '/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/search/query/hover.ts?L202:14',
        },
    },
    {
        id: 'TourGoToDefinition',
        title: 'Go to a definition',
        group: CODE_INTEL,
        info: `<strong>GO TO DEFINITION</strong><br/>
            Hover over a token in the highlighted line to open code intel, then click ‘Go to definition’ to locate a token definition.`,
        completeAfterEvents: ['goToDefinition', 'goToDefinition.preloaded'],
        url: {
            [OnboardingTourLanguage.C]: '/github.com/torvalds/linux/-/blob/arch/arm/kernel/bios32.c?L417:8',
            [OnboardingTourLanguage.Go]:
                '/github.com/sourcegraph/sourcegraph/-/blob/internal/repos/observability.go?L192:22',
            [OnboardingTourLanguage.Java]:
                '/github.com/square/okhttp/-/blob/samples/guide/src/main/java/okhttp3/recipes/CustomCipherSuites.java?L132:14',
            [OnboardingTourLanguage.Javascript]: '/github.com/mozilla/pdf.js/-/blob/src/pdf.js?L101:5',
            [OnboardingTourLanguage.Php]:
                '/github.com/square/connect-api-examples/-/blob/connect-examples/v1/php/payments-report.php?L164:32',
            [OnboardingTourLanguage.Python]:
                '/github.com/netdata/netdata@1c2465c816071ff767982116a4b19bad1d8b0c82/-/blob/collectors/python.d.plugin/python_modules/bases/charts.py?L303:48',
            [OnboardingTourLanguage.Typescript]:
                '/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/search/query/parserFuzz.ts?L295:37',
        },
    },
    // Group: TOOLS
    {
        id: 'TourEditorExtensions',
        group: TOOLS,
        title: 'IDE extensions',
        url: 'https://docs.sourcegraph.com/integration/editor',
    },
    {
        id: 'TourBrowserExtensions',
        group: TOOLS,
        title: 'Browser extensions',
        url: 'https://docs.sourcegraph.com/integration/browser_extension',
    },
]
