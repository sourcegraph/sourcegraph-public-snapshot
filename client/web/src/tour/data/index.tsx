import React from 'react'

import { TourLanguage, TourTaskType } from '../components/Tour/types'

import {
    IconAllDone,
    IconCreateTeam,
    IconFindCodeReference,
    IconInstallIDEExtension,
    IconPowerfulCodeNavigation,
    IconResolveIncidentsFaster,
} from './icons'

/**
 * Tour tasks for non-authenticated users
 */
export const visitorsTasks: TourTaskType[] = [
    {
        title: 'Code search use cases',
        steps: [
            {
                id: 'TourSymbolsSearch',
                label: 'Search multiple repos',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]:
                            '/search?q=context:global+repo:torvalds/.*+lang:c+-file:.*/testing+magic&patternType=literal',
                        [TourLanguage.Go]:
                            '/search?q=context:global+r:google/+lang:go+-file:test+//+TODO&patternType=literal',
                        [TourLanguage.Java]:
                            '/search?q=context:global+r:github.com/square/+lang:java+-file:test+GiftCard&patternType=literal',
                        [TourLanguage.Javascript]:
                            '/search?q=context:global+r:react+lang:JavaScript+-file:test+createPortal&patternType=literal',
                        [TourLanguage.Php]:
                            '/search?q=context:global+repo:laravel+lang:php+-file:test+login%28&patternType=regexp&case=yes',
                        [TourLanguage.Python]:
                            '/search?q=context:global+r:aws/+lang:python+file:mock+def+test_patch&patternType=regexp&case=yes',
                        [TourLanguage.Typescript]:
                            '/search?q=context:global+r:react+lang:typescript+-file:test+createPortal%28&patternType=regexp&case=yes',
                    },
                },
                info: `<strong>Reference code in multiple repositories</strong><br/>
            The repo: query allows searching in multiple repositories matching a term. Use it to reference all of your projects or find open source examples.`,
            },
            {
                id: 'TourCommitsSearch',
                label: 'Find changes in commits',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]:
                            '/search?q=context:global+repo:%5Egithub%5C.com/chref/doh%24+file:%5Edoh%5C.c+type:commit+option&patternType=literal',
                        [TourLanguage.Go]:
                            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:commit+bump&patternType=literal',
                        [TourLanguage.Java]:
                            '/search?q=context:global+repo:sourcegraph-testing/.*+type:commit+lang:java+version&patternType=literal',
                        [TourLanguage.Javascript]:
                            '/search?q=context:global+repo:%5Egithub%5C.com/hakimel/reveal%5C.js%24+type:commit+error&patternType=literal',
                        [TourLanguage.Php]:
                            '/search?q=context:global+repo:square/connect-api-examples+type:commit+version&patternType=regexp&case=yes',
                        [TourLanguage.Python]:
                            '/search?q=context:global+r:elastic/elasticsearch+lang:python+type:commit+request_timeout&patternType=regexp&case=yes',
                        [TourLanguage.Typescript]:
                            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:commit+bump&patternType=literal',
                    },
                },
                info: `<strong>Find changes in commits</strong><br/>
            Quickly find commits in history, then browse code from the commit, without checking out the branch.`,
            },
            {
                id: 'TourDiffSearch',
                label: 'Search diffs for removed code',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]:
                            '/search?q=context:global+repo:chref/doh+type:diff+select:commit.diff.removed+mode&patternType=literal',
                        [TourLanguage.Go]:
                            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:diff+lang:go+select:commit.diff.removed+NameSpaceOrgId&patternType=literal',
                        [TourLanguage.Java]:
                            '/search?q=context:global+repo:sourcegraph-testing/sg-hadoop+lang:java+type:diff+select:commit.diff.removed+getConf&patternType=literal',
                        [TourLanguage.Javascript]:
                            '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:javascript+-file:test+type:diff+select:commit.diff.removed+promise&patternType=literal',
                        [TourLanguage.Php]:
                            '/search?q=context:global+repo:laravel/laravel.*+lang:php++type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
                        [TourLanguage.Python]:
                            '/search?q=context:global+repo:pallets/+lang:python+type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
                        [TourLanguage.Typescript]:
                            '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:typescript+type:diff+select:commit.diff.removed+authenticatedUser&patternType=regexp&case=yes',
                    },
                },
                info:
                    '<strong>Searching diffs for removed code</strong><br/>Find removed code without browsing through history or trying to remember which file it was in.',
            },
        ],
    },
    {
        title: 'The power of code intel',
        steps: [
            {
                id: 'TourFindReferences',
                label: 'Find references',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]: '/github.com/torvalds/linux/-/blob/arch/arm/kernel/atags_compat.c?L196:8',
                        [TourLanguage.Go]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/internal/featureflag/featureflag.go?L9:6',
                        [TourLanguage.Java]:
                            '/github.com/square/okhttp/-/blob/samples/guide/src/main/java/okhttp3/recipes/PrintEvents.java?L126:27',
                        [TourLanguage.Javascript]:
                            '/github.com/mozilla/pdf.js/-/blob/src/display/display_utils.js?L261:16',
                        [TourLanguage.Php]:
                            '/github.com/square/connect-api-examples/-/blob/connect-examples/v1/php/payments-report.php?L164:32',
                        [TourLanguage.Python]: '/github.com/google-research/bert/-/blob/extract_features.py?L152:7',
                        [TourLanguage.Typescript]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/search/query/hover.ts?L202:14',
                    },
                },
                info:
                    '<strong>FIND REFERENCES</strong><br/>Hover over a token in the highlighted line to open code intel, then click ‘Find References’ to locate all calls of this code.',
                completeAfterEvents: ['findReferences'],
            },
            {
                id: 'TourGoToDefinition',
                label: 'Go to a definition',
                info: `<strong>GO TO DEFINITION</strong><br/>
            Hover over a token in the highlighted line to open code intel, then click ‘Go to definition’ to locate a token definition.`,
                completeAfterEvents: ['goToDefinition', 'goToDefinition.preloaded'],
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]: '/github.com/torvalds/linux/-/blob/arch/arm/kernel/bios32.c?L417:8',
                        [TourLanguage.Go]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/internal/repos/observability.go?L192:22',
                        [TourLanguage.Java]:
                            '/github.com/square/okhttp/-/blob/samples/guide/src/main/java/okhttp3/recipes/CustomCipherSuites.java?L132:14',
                        [TourLanguage.Javascript]: '/github.com/mozilla/pdf.js/-/blob/src/pdf.js?L101:5',
                        [TourLanguage.Php]:
                            '/github.com/square/connect-api-examples/-/blob/connect-examples/v1/php/payments-report.php?L164:32',
                        [TourLanguage.Python]:
                            '/github.com/netdata/netdata@1c2465c816071ff767982116a4b19bad1d8b0c82/-/blob/collectors/python.d.plugin/python_modules/bases/charts.py?L303:48',
                        [TourLanguage.Typescript]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/search/query/parserFuzz.ts?L295:37',
                    },
                },
            },
        ],
    },
    {
        title: 'Tools to improve workflow',
        steps: [
            {
                id: 'TourEditorExtensions',
                label: 'IDE extensions',
                action: { type: 'link', value: 'https://docs.sourcegraph.com/integration/editor' },
            },
            {
                id: 'TourBrowserExtensions',
                label: 'Browser extensions',
                action: { type: 'link', value: 'https://docs.sourcegraph.com/integration/browser_extension' },
            },
        ],
    },
    {
        title: 'Install or sign up',
        steps: [
            {
                id: 'InstallOrSignUp',
                label: 'Get powerful code search and other features on your private code.',
                action: {
                    type: 'link',
                    value:
                        'https://about.sourcegraph.com/get-started?utm_medium=inproduct&utm_source=getting-started-tour&utm_campaign=inproduct-cta&_ga=2.130711115.51352124.1647511547-1994718421.1647511547',
                },
                // This is done to mimic user creating an account, and signed in there is a different tour
                completeAfterEvents: ['non-existing-event'],
            },
        ],
    },
]

/**
 * Tour tasks for authenticated users. Extended/all use-cases.
 */
export const authenticatedTasks: TourTaskType[] = [
    {
        title: 'Find code to reference',
        icon: <IconFindCodeReference />,
        steps: [
            {
                id: 'FindCodeRef',
                label: 'Search for code in a user or org’s repos while excluding test files.',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]:
                            '/search?q=context:global+repo:torvalds/.*+lang:c+-file:.*/testing+magic&patternType=literal',
                        [TourLanguage.Go]:
                            '/search?q=context:global+r:google/+lang:go+-file:test+//+TODO&patternType=literal',
                        [TourLanguage.Java]:
                            '/search?q=context:global+r:github.com/square/+lang:java+-file:test+GiftCard&patternType=literal',
                        [TourLanguage.Javascript]:
                            '/search?q=context:global+r:react+lang:JavaScript+-file:test+createPortal&patternType=literal',
                        [TourLanguage.Php]:
                            '/search?q=context:global+repo:laravel+lang:php+-file:test+login%28&patternType=regexp&case=yes',
                        [TourLanguage.Python]:
                            '/search?q=context:global+r:aws/+lang:python+file:mock+def+test_patch&patternType=regexp&case=yes',
                        [TourLanguage.Typescript]:
                            '/search?q=context:global+r:react+lang:typescript+-file:test+createPortal%28&patternType=regexp&case=yes',
                    },
                },
                info: `<strong>Reference code in multiple repositories</strong><br/>
            The repo: query allows searching in multiple repositories matching a term. Use it to reference all of your projects or find open source examples.`,
            },
        ],
    },
    {
        title: 'Resolve incidents faster',
        icon: <IconResolveIncidentsFaster />,
        steps: [
            {
                id: 'WatchVideo',
                label: 'Watch the 60 second video',
                action: { type: 'video', value: 'https://www.youtube-nocookie.com/embed/XLfE2YuRwvw' },
            },
            {
                id: 'DiffSearch',
                label: 'Run a diff search',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]:
                            '/search?q=context:global+repo:chref/doh+type:diff+select:commit.diff.removed+mode&patternType=literal',
                        [TourLanguage.Go]:
                            '/search?q=context:global+repo:%5Egitlab%5C.com/sourcegraph/sourcegraph%24+type:diff+lang:go+select:commit.diff.removed+NameSpaceOrgId&patternType=literal',
                        [TourLanguage.Java]:
                            '/search?q=context:global+repo:sourcegraph-testing/sg-hadoop+lang:java+type:diff+select:commit.diff.removed+getConf&patternType=literal',
                        [TourLanguage.Javascript]:
                            '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:javascript+-file:test+type:diff+select:commit.diff.removed+promise&patternType=literal',
                        [TourLanguage.Php]:
                            '/search?q=context:global+repo:laravel/laravel.*+lang:php++type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
                        [TourLanguage.Python]:
                            '/search?q=context:global+repo:pallets/+lang:python+type:diff+select:commit.diff.removed+password&patternType=regexp&case=yes',
                        [TourLanguage.Typescript]:
                            '/search?q=context:global+repo:sourcegraph/sourcegraph%24+lang:typescript+type:diff+select:commit.diff.removed+authenticatedUser&patternType=regexp&case=yes',
                    },
                },
                info:
                    '<strong>Searching diffs for removed code</strong><br/>Find removed code without browsing through history or trying to remember which file it was in.',
            },
        ],
    },
    {
        title: 'Powerful code navigation',
        icon: <IconPowerfulCodeNavigation />,
        steps: [
            {
                id: 'PowerCodeNav',
                label: 'Get IDE like go to definition and find references across many repositories',
                action: {
                    type: 'link',
                    value: {
                        [TourLanguage.C]: '/github.com/torvalds/linux/-/blob/arch/arm/kernel/atags_compat.c?L196:8',
                        [TourLanguage.Go]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/internal/featureflag/featureflag.go?L9:6',
                        [TourLanguage.Java]:
                            '/github.com/square/okhttp/-/blob/samples/guide/src/main/java/okhttp3/recipes/PrintEvents.java?L126:27',
                        [TourLanguage.Javascript]:
                            '/github.com/mozilla/pdf.js/-/blob/src/display/display_utils.js?L261:16',
                        [TourLanguage.Php]:
                            '/github.com/square/connect-api-examples/-/blob/connect-examples/v1/php/payments-report.php?L164:32',
                        [TourLanguage.Python]: '/github.com/google-research/bert/-/blob/extract_features.py?L152:7',
                        [TourLanguage.Typescript]:
                            '/github.com/sourcegraph/sourcegraph/-/blob/client/shared/src/search/query/hover.ts?L202:14',
                    },
                },
                info:
                    '<strong>FIND REFERENCES</strong><br/>Hover over a token in the highlighted line to open code intel, then click ‘Find References’ to locate all calls of this code.',
                completeAfterEvents: ['findReferences'],
            },
        ],
    },
    {
        title: 'Install an IDE extension',
        icon: <IconInstallIDEExtension />,
        steps: [
            {
                id: 'InstallIDEExtension',
                label: 'Integrate Sourcegraph with VSCode, Jetbrains, Sublime, and Atom.',
                action: { type: 'link', value: 'https://docs.sourcegraph.com/integration/editor' },
            },
        ],
    },
    {
        title: 'Create a team',
        icon: <IconCreateTeam />,
        steps: [
            {
                id: 'CreateTeam',
                label: 'Sourcegraph helps teams from 2 to any size collaborate.',
                action: {
                    type: 'link',
                    value:
                        'https://share.hsforms.com/14OQ3RoPpQTOXvZlUpgx6-A1n7ku?utm_medium=direct-traffic&utm_source=in-product&utm_term=in-product-banner&utm_content=cloud-product-beta-teams',
                },
            },
        ],
    },
]

/**
 * Tour extra tasks for authenticated users.
 */
export const authenticatedExtraTask: TourTaskType = {
    title: 'All done!',
    icon: <IconAllDone />,
    steps: [
        {
            id: 'RestartTour',
            label: 'You can restart the tour to go through the steps again.',
            action: { type: 'restart', value: 'Restart tour' },
        },
    ],
}
