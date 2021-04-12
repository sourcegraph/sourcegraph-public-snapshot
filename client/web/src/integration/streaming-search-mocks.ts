/* eslint-disable no-template-curly-in-string */
import { WebGraphQlOperations } from '../graphql-operations'
import { SearchEvent } from '../search/stream'

export const diffSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'commit',
                label:
                    '[sourcegraph/sourcegraph](/gitlab.sgdev.org/sourcegraph/sourcegraph) › [Rijnard van Tonder](/gitlab.sgdev.org/sourcegraph/sourcegraph/-/commit/b6dd338737c090fdab31d324542bfdaa7ce9f766): [search: if not specified, set fork:no by default (#8739)](/gitlab.sgdev.org/sourcegraph/sourcegraph/-/commit/b6dd338737c090fdab31d324542bfdaa7ce9f766)',
                url: '/gitlab.sgdev.org/sourcegraph/sourcegraph/-/commit/b6dd338737c090fdab31d324542bfdaa7ce9f766',
                detail:
                    '[`b6dd338` one year ago](/gitlab.sgdev.org/sourcegraph/sourcegraph/-/commit/b6dd338737c090fdab31d324542bfdaa7ce9f766)',
                content:
                    "```diff\nweb/src/regression/search.test.ts web/src/regression/search.test.ts\n@@ -434,0 +435,3 @@ describe('Search regression test suite', () => {\n+        test('Fork repos excluded by default', async () => {\n+            const urlQuery = buildSearchURLQuery('type:repo sgtest/mux', GQL.SearchPatternType.regexp, false)\n+            await driver.page.goto(config.sourcegraphBaseUrl + '/search?' + urlQuery)\n@@ -434,0 +439,4 @@ describe('Search regression test suite', () => {\n+        })\n+        test('Forked repos included by by fork option', async () => {\n+            const urlQuery = buildSearchURLQuery('type:repo sgtest/mux fork:yes', GQL.SearchPatternType.regexp, false)\n+            await driver.page.goto(config.sourcegraphBaseUrl + '/search?' + urlQuery)\n```",
                ranges: [
                    [-1, 30, 4],
                    [-4, 30, 4],
                    [3, 9, 4],
                    [4, 63, 4],
                    [8, 9, 4],
                    [9, 63, 4],
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

export const diffHighlightResult: Partial<WebGraphQlOperations> = {
    highlightCode: ({ isLightTheme }) => ({
        highlightCode: isLightTheme
            ? '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#657b83;">web/src/regression/search.test.ts web/src/regression/search.test.ts\n</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#268bd2;">@@ -434,0 +435,3 @@ </span><span style="color:#cb4b16;">describe(&#39;Search regression test suite&#39;, () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+        test(&#39;Fork repos excluded by default&#39;, async () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+            const urlQuery = buildSearchURLQuery(&#39;type:repo sgtest/mux&#39;, GQL.SearchPatternType.regexp, false)\n</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+            await driver.page.goto(config.sourcegraphBaseUrl + &#39;/search?&#39; + urlQuery)\n</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#268bd2;">@@ -434,0 +439,4 @@ </span><span style="color:#cb4b16;">describe(&#39;Search regression test suite&#39;, () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+        })\n</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+        test(&#39;Forked repos included by by fork option&#39;, async () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="9"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+            const urlQuery = buildSearchURLQuery(&#39;type:repo sgtest/mux fork:yes&#39;, GQL.SearchPatternType.regexp, false)\n</span></div></td></tr><tr><td class="line" data-line="10"></td><td class="code"><div><span style="background-color:#deeade;color:#2b3750;">+            await driver.page.goto(config.sourcegraphBaseUrl + &#39;/search?&#39; + urlQuery)</span></div></td></tr></tbody></table>'
            : '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#969896;">web/src/regression/search.test.ts web/src/regression/search.test.ts\n</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#268bd2;">@@ -434,0 +435,3 @@ </span><span style="color:#8fa1b3;">describe(&#39;Search regression test suite&#39;, () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+        test(&#39;Fork repos excluded by default&#39;, async () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+            const urlQuery = buildSearchURLQuery(&#39;type:repo sgtest/mux&#39;, GQL.SearchPatternType.regexp, false)\n</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+            await driver.page.goto(config.sourcegraphBaseUrl + &#39;/search?&#39; + urlQuery)\n</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#268bd2;">@@ -434,0 +439,4 @@ </span><span style="color:#8fa1b3;">describe(&#39;Search regression test suite&#39;, () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+        })\n</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+        test(&#39;Forked repos included by by fork option&#39;, async () =&gt; {\n</span></div></td></tr><tr><td class="line" data-line="9"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+            const urlQuery = buildSearchURLQuery(&#39;type:repo sgtest/mux fork:yes&#39;, GQL.SearchPatternType.regexp, false)\n</span></div></td></tr><tr><td class="line" data-line="10"></td><td class="code"><div><span style="background-color:#0e2414;color:#f2f4f8;">+            await driver.page.goto(config.sourcegraphBaseUrl + &#39;/search?&#39; + urlQuery)</span></div></td></tr></tbody></table>',
    }),
}

export const commitSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'commit',
                label:
                    '[sourcegraph/sourcegraph](/github.com/sourcegraph/sourcegraph) › [Camden Cheek](/github.com/sourcegraph/sourcegraph/-/commit/f7d28599cad80e200913d9c4612618a73199bac1): [search: Incorporate search blitz (#19567)](/github.com/sourcegraph/sourcegraph/-/commit/f7d28599cad80e200913d9c4612618a73199bac1)',
                url: '/github.com/sourcegraph/sourcegraph/-/commit/f7d28599cad80e200913d9c4612618a73199bac1',
                detail:
                    '[`f7d2859` 2 days ago](/github.com/sourcegraph/sourcegraph/-/commit/f7d28599cad80e200913d9c4612618a73199bac1)',
                content:
                    '```COMMIT_EDITMSG\nsearch: Incorporate search blitz (#19567)\n\nIncorporates search blitz into sourcegraph/sourcegraph so it has access to the internal streaming client\n```',
                ranges: [
                    [3, 37, 5],
                    [3, 49, 5],
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

export const commitHighlightResult: Partial<WebGraphQlOperations> = {
    highlightCode: ({ isLightTheme }) => ({
        highlightCode: isLightTheme
            ? '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#657b83;">search: Incorporate search blitz (#19567)\n</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#657b83;">Incorporates search blitz into sourcegraph/sourcegraph so it has access to the internal streaming client</span></div></td></tr></tbody></table>'
            : '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#c0c5ce;">search: Incorporate search blitz (#19567)\n</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#c0c5ce;">Incorporates search blitz into sourcegraph/sourcegraph so it has access to the internal streaming client</span></div></td></tr></tbody></table>',
    }),
}

export const mixedSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            { type: 'repo', repository: 'gitlab.sgdev.org/lg-test-private/lg-test' },
            {
                type: 'file',
                name: 'overridable/bool_or_string_test.go',
                repository: 'gitlab.sgdev.org/aharvey/batch-change-utils',
                branches: [''],
                version: '206c057cc03eea48300a4bd33f4dc4222d242114',
                lineMatches: [],
            },
            {
                type: 'file',
                name: 'src/main.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/lsif-cpp',
                branches: [''],
                version: '2e3569cf60646c9ce4e37a43e5cf698a00cbd41a',
                lineMatches: [
                    {
                        line: "test('does not emit items with duplicate IDs', async () =\u003E {",
                        lineNumber: 38,
                        offsetAndLengths: [[0, 4]],
                    },
                    { line: "test('five', async () =\u003E {", lineNumber: 63, offsetAndLengths: [[0, 4]] },
                ],
            },
        ],
    },
    {
        type: 'filters',
        data: [
            { value: 'lang:go', label: 'lang:go', count: 1092, limitHit: false, kind: 'lang' },
            {
                value: '-file:_test\\.go$',
                label: '-file:_test\\.go$',
                count: 663,
                limitHit: false,
                kind: 'file',
            },
            {
                value: 'lang:typescript',
                label: 'lang:typescript',
                count: 379,
                limitHit: false,
                kind: 'lang',
            },
            { value: 'lang:markdown', label: 'lang:markdown', count: 343, limitHit: false, kind: 'lang' },
            { value: 'lang:yaml', label: 'lang:yaml', count: 193, limitHit: false, kind: 'lang' },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/sourcegraph/src-cli$',
                label: 'gitlab.sgdev.org/sourcegraph/src-cli',
                count: 156,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^ghe\\.sgdev\\.org/sourcegraph/gorillalabs-sparkling$',
                label: 'ghe.sgdev.org/sourcegraph/gorillalabs-sparkling',
                count: 145,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/sourcegraph/java-langserver$',
                label: 'gitlab.sgdev.org/sourcegraph/java-langserver',
                count: 142,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/sourcegraph/go-jsonschema$',
                label: 'gitlab.sgdev.org/sourcegraph/go-jsonschema',
                count: 130,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/aharvey/batch-change-utils$',
                label: 'gitlab.sgdev.org/aharvey/batch-change-utils',
                count: 125,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/sourcegraph/about$',
                label: 'gitlab.sgdev.org/sourcegraph/about',
                count: 125,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^ghe\\.sgdev\\.org/sourcegraph/gorilla-websocket$',
                label: 'ghe.sgdev.org/sourcegraph/gorilla-websocket',
                count: 123,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^github\\.com/hashicorp/go-multierror$',
                label: 'github.com/hashicorp/go-multierror',
                count: 121,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/sourcegraph/sourcegraph$',
                label: 'gitlab.sgdev.org/sourcegraph/sourcegraph',
                count: 115,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^github\\.com/sourcegraph/sourcegraph$',
                label: 'github.com/sourcegraph/sourcegraph',
                count: 112,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^gitlab\\.sgdev\\.org/aharvey/sourcegraph$',
                label: 'gitlab.sgdev.org/aharvey/sourcegraph',
                count: 109,
                limitHit: false,
                kind: 'repo',
            },
            {
                value: 'repo:^ghe\\.sgdev\\.org/sourcegraph/gorilla-mux$',
                label: 'ghe.sgdev.org/sourcegraph/gorilla-mux',
                count: 108,
                limitHit: false,
                kind: 'repo',
            },
            { value: 'lang:java', label: 'lang:java', count: 95, limitHit: false, kind: 'lang' },
            { value: 'lang:json', label: 'lang:json', count: 77, limitHit: false, kind: 'lang' },
            { value: 'lang:graphql', label: 'lang:graphql', count: 70, limitHit: false, kind: 'lang' },
            { value: 'lang:text', label: 'lang:text', count: 50, limitHit: false, kind: 'lang' },
            { value: 'lang:clojure', label: 'lang:clojure', count: 45, limitHit: false, kind: 'lang' },
            { value: 'lang:css', label: 'lang:css', count: 32, limitHit: false, kind: 'lang' },
        ],
    },
    { type: 'done', data: {} },
]

export const highlightFileResult: Partial<WebGraphQlOperations> = {
    HighlightedFile: ({ isLightTheme }) =>
        isLightTheme
            ? {
                  repository: {
                      commit: {
                          file: {
                              isDirectory: false,
                              richHTML: '',
                              highlight: {
                                  aborted: false,
                                  lineRanges: [
                                      [
                                          '<tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">index </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./index</span><span style="color:#839496;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">Edge</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">Vertex </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">lsif-protocol</span><span style="color:#839496;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">_ </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">lodash</span><span style="color:#839496;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#b58900;">* </span><span style="color:#cb4b16;">as </span><span style="color:#268bd2;">path </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">path</span><span style="color:#839496;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#b58900;">* </span><span style="color:#cb4b16;">as </span><span style="color:#268bd2;">cp </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">child_process</span><span style="color:#839496;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#268bd2;">const GENERATE </span><span style="color:#657b83;">= </span><span style="color:#b58900;">false\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="9"></td><td class="code"><div><span style="color:#268bd2;">function </span><span style="color:#b58900;">generate</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">example</span><span style="color:#859900;">: string</span><span style="color:#657b83;">)</span><span style="color:#859900;">: void </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="10"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">cp</span><span style="color:#657b83;">.</span><span style="color:#b58900;">execFileSync</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./generate-csv</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">[</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">$CXX -c *.cpp</span><span style="color:#839496;">&#39;</span><span style="color:#268bd2;">]</span><span style="color:#657b83;">, {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="11"></td><td class="code"><div><span style="color:#657b83;">        env: {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="12"></td><td class="code"><div><span style="color:#657b83;">            ABSROOTDIR: </span><span style="color:#268bd2;">path</span><span style="color:#657b83;">.</span><span style="color:#859900;">resolve</span><span style="color:#657b83;">(</span><span style="color:#839496;">`</span><span style="color:#2aa198;">examples/${</span><span style="color:#268bd2;">example</span><span style="color:#2aa198;">}/root</span><span style="color:#839496;">`</span><span style="color:#657b83;">),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="13"></td><td class="code"><div><span style="color:#657b83;">            ABSOUTDIR: </span><span style="color:#268bd2;">path</span><span style="color:#657b83;">.</span><span style="color:#859900;">resolve</span><span style="color:#657b83;">(</span><span style="color:#839496;">`</span><span style="color:#2aa198;">examples/${</span><span style="color:#268bd2;">example</span><span style="color:#2aa198;">}/output</span><span style="color:#839496;">`</span><span style="color:#657b83;">),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="14"></td><td class="code"><div><span style="color:#657b83;">            CLEAN: </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">true</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="15"></td><td class="code"><div><span style="color:#657b83;">        },\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="16"></td><td class="code"><div><span style="color:#657b83;">    })\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="17"></td><td class="code"><div><span style="color:#657b83;">}\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="18"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="19"></td><td class="code"><div><span style="color:#586e75;">async </span><span style="color:#268bd2;">function </span><span style="color:#b58900;">indexExample</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">example</span><span style="color:#859900;">: string</span><span style="color:#657b83;">)</span><span style="color:#859900;">: </span><span style="color:#b58900;">Promise</span><span style="color:#657b83;">&lt;(</span><span style="color:#b58900;">Edge </span><span style="color:#859900;">| </span><span style="color:#b58900;">Vertex</span><span style="color:#657b83;">)</span><span style="color:#268bd2;">[]</span><span style="color:#657b83;">&gt; {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="20"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#859900;">if </span><span style="color:#657b83;">(</span><span style="color:#268bd2;">GENERATE</span><span style="color:#657b83;">) {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="21"></td><td class="code"><div><span style="color:#657b83;">        </span><span style="color:#b58900;">generate</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">example</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="22"></td><td class="code"><div><span style="color:#657b83;">    }\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="23"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="24"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const output</span><span style="color:#859900;">: </span><span style="color:#657b83;">(</span><span style="color:#b58900;">Edge </span><span style="color:#859900;">| </span><span style="color:#b58900;">Vertex</span><span style="color:#657b83;">)</span><span style="color:#268bd2;">[] </span><span style="color:#657b83;">= </span><span style="color:#268bd2;">[]\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="25"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="26"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#859900;">await </span><span style="color:#b58900;">index</span><span style="color:#657b83;">({\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="27"></td><td class="code"><div><span style="color:#657b83;">        csvFileGlob: </span><span style="color:#839496;">`</span><span style="color:#2aa198;">examples/${</span><span style="color:#268bd2;">example</span><span style="color:#2aa198;">}/output/*.csv</span><span style="color:#839496;">`</span><span style="color:#657b83;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="28"></td><td class="code"><div><span style="color:#657b83;">        root: </span><span style="color:#839496;">`</span><span style="color:#2aa198;">examples/${</span><span style="color:#268bd2;">example</span><span style="color:#2aa198;">}/root</span><span style="color:#839496;">`</span><span style="color:#657b83;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="29"></td><td class="code"><div><span style="color:#657b83;">        </span><span style="color:#b58900;">emit</span><span style="color:#657b83;">: </span><span style="color:#268bd2;">item =&gt;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="30"></td><td class="code"><div><span style="color:#657b83;">            </span><span style="color:#859900;">new </span><span style="color:#b58900;">Promise</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">resolve =&gt; </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="31"></td><td class="code"><div><span style="color:#657b83;">                </span><span style="color:#268bd2;">output</span><span style="color:#657b83;">.</span><span style="color:#859900;">push</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">item</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="32"></td><td class="code"><div><span style="color:#657b83;">                </span><span style="color:#b58900;">resolve</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="33"></td><td class="code"><div><span style="color:#657b83;">            }),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="34"></td><td class="code"><div><span style="color:#657b83;">    })\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="35"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="36"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#859900;">return </span><span style="color:#268bd2;">output\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="37"></td><td class="code"><div><span style="color:#657b83;">}\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="38"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="39"></td><td class="code"><div><span style="color:#b58900;">test</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">does not emit items with duplicate IDs</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">, </span><span style="color:#586e75;">async </span><span style="color:#657b83;">() </span><span style="color:#268bd2;">=&gt; </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="40"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const output </span><span style="color:#657b83;">= </span><span style="color:#859900;">await </span><span style="color:#b58900;">indexExample</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">five</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="41"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="42"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const setsOfDupes </span><span style="color:#657b83;">= </span><span style="color:#b58900;">_</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">output</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="43"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#b58900;">groupBy</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">item =&gt; item</span><span style="color:#657b83;">.</span><span style="color:#859900;">id</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="44"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#859900;">values</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="45"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">group =&gt; </span><span style="color:#657b83;">({ </span><span style="color:#268bd2;">group</span><span style="color:#657b83;">, count: </span><span style="color:#268bd2;">group</span><span style="color:#657b83;">.</span><span style="color:#859900;">length </span><span style="color:#657b83;">}))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="46"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#b58900;">value</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="47"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#b58900;">filter</span><span style="color:#657b83;">(({ </span><span style="color:#268bd2;">count </span><span style="color:#657b83;">}) </span><span style="color:#268bd2;">=&gt; count </span><span style="color:#859900;">&gt; </span><span style="color:#6c71c4;">1</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="48"></td><td class="code"><div><span style="color:#657b83;">        .</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(({ </span><span style="color:#268bd2;">group </span><span style="color:#657b83;">}) </span><span style="color:#268bd2;">=&gt; group</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="49"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="50"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#859900;">if </span><span style="color:#657b83;">(</span><span style="color:#268bd2;">setsOfDupes</span><span style="color:#657b83;">.</span><span style="color:#859900;">length &gt; </span><span style="color:#6c71c4;">0</span><span style="color:#657b83;">) {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="51"></td><td class="code"><div><span style="color:#657b83;">        </span><span style="color:#b58900;">fail</span><span style="color:#657b83;">(\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="52"></td><td class="code"><div><span style="color:#657b83;">            </span><span style="color:#859900;">new </span><span style="color:#b58900;">Error</span><span style="color:#657b83;">(\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="53"></td><td class="code"><div><span style="color:#657b83;">                </span><span style="color:#839496;">`</span><span style="color:#2aa198;">Sets of lines with duplicate IDs:</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">` </span><span style="color:#657b83;">+\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="54"></td><td class="code"><div><span style="color:#657b83;">                    </span><span style="color:#268bd2;">setsOfDupes\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="55"></td><td class="code"><div><span style="color:#657b83;">                        .</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">dupes =&gt;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="56"></td><td class="code"><div><span style="color:#657b83;">                            </span><span style="color:#268bd2;">dupes</span><span style="color:#657b83;">.</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">item =&gt; </span><span style="color:#859900;">JSON</span><span style="color:#657b83;">.</span><span style="color:#859900;">stringify</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">item</span><span style="color:#657b83;">)).</span><span style="color:#859900;">join</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="57"></td><td class="code"><div><span style="color:#657b83;">                        )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="58"></td><td class="code"><div><span style="color:#657b83;">                        .</span><span style="color:#859900;">join</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#dc322f;">\\n\\n</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="59"></td><td class="code"><div><span style="color:#657b83;">            )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="60"></td><td class="code"><div><span style="color:#657b83;">        )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="61"></td><td class="code"><div><span style="color:#657b83;">    }\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="62"></td><td class="code"><div><span style="color:#657b83;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="63"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="64"></td><td class="code"><div><span style="color:#b58900;">test</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">five</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">, </span><span style="color:#586e75;">async </span><span style="color:#657b83;">() </span><span style="color:#268bd2;">=&gt; </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="65"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const output </span><span style="color:#657b83;">= (</span><span style="color:#859900;">await </span><span style="color:#b58900;">indexExample</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">five</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v =&gt; </span><span style="color:#859900;">JSON</span><span style="color:#657b83;">.</span><span style="color:#859900;">stringify</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v</span><span style="color:#657b83;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="66"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="67"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#b58900;">expect</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">output</span><span style="color:#657b83;">.</span><span style="color:#859900;">join</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">toMatchSnapshot</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="68"></td><td class="code"><div><span style="color:#657b83;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="69"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="70"></td><td class="code"><div><span style="color:#b58900;">test</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">cross-app</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">, </span><span style="color:#586e75;">async </span><span style="color:#657b83;">() </span><span style="color:#268bd2;">=&gt; </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="71"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const app </span><span style="color:#657b83;">= (</span><span style="color:#859900;">await </span><span style="color:#b58900;">indexExample</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">cross-app</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v =&gt; </span><span style="color:#859900;">JSON</span><span style="color:#657b83;">.</span><span style="color:#859900;">stringify</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v</span><span style="color:#657b83;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="72"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#b58900;">expect</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">app</span><span style="color:#657b83;">.</span><span style="color:#859900;">join</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">toMatchSnapshot</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="73"></td><td class="code"><div><span style="color:#657b83;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="74"></td><td class="code"><div><span style="color:#657b83;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="75"></td><td class="code"><div><span style="color:#b58900;">test</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">cross-lib</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">, </span><span style="color:#586e75;">async </span><span style="color:#657b83;">() </span><span style="color:#268bd2;">=&gt; </span><span style="color:#657b83;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="76"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#268bd2;">const lib </span><span style="color:#657b83;">= (</span><span style="color:#859900;">await </span><span style="color:#b58900;">indexExample</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">cross-lib</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">map</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v =&gt; </span><span style="color:#859900;">JSON</span><span style="color:#657b83;">.</span><span style="color:#859900;">stringify</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">v</span><span style="color:#657b83;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="77"></td><td class="code"><div><span style="color:#657b83;">    </span><span style="color:#b58900;">expect</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">lib</span><span style="color:#657b83;">.</span><span style="color:#859900;">join</span><span style="color:#657b83;">(</span><span style="color:#839496;">&#39;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#39;</span><span style="color:#657b83;">)).</span><span style="color:#b58900;">toMatchSnapshot</span><span style="color:#657b83;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="78"></td><td class="code"><div><span style="color:#657b83;">})</span></div></td></tr>',
                                      ],
                                  ],
                              },
                          },
                      },
                  },
              }
            : {
                  repository: {
                      commit: {
                          file: {
                              isDirectory: false,
                              richHTML: '',
                              highlight: {
                                  aborted: false,
                                  lineRanges: [
                                      [
                                          '<tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#35a5ff;">import </span><span style="color:#c0c5ce;">{ </span><span style="color:#72c3fc;">index </span><span style="color:#c0c5ce;">} </span><span style="color:#35a5ff;">from </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">./index</span><span style="color:#bdd4e3;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#35a5ff;">import </span><span style="color:#c0c5ce;">{ </span><span style="color:#72c3fc;">Edge</span><span style="color:#c0c5ce;">, </span><span style="color:#72c3fc;">Vertex </span><span style="color:#c0c5ce;">} </span><span style="color:#35a5ff;">from </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">lsif-protocol</span><span style="color:#bdd4e3;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#35a5ff;">import </span><span style="color:#72c3fc;">_ </span><span style="color:#35a5ff;">from </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">lodash</span><span style="color:#bdd4e3;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#35a5ff;">import </span><span style="color:#329af0;">* </span><span style="color:#35a5ff;">as </span><span style="color:#72c3fc;">path </span><span style="color:#35a5ff;">from </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">path</span><span style="color:#bdd4e3;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#35a5ff;">import </span><span style="color:#329af0;">* </span><span style="color:#35a5ff;">as </span><span style="color:#72c3fc;">cp </span><span style="color:#35a5ff;">from </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">child_process</span><span style="color:#bdd4e3;">&#39;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#329af0;">const </span><span style="color:#72c3fc;">GENERATE </span><span style="color:#329af0;">= false\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="9"></td><td class="code"><div><span style="color:#329af0;">function </span><span style="color:#fff3bf;">generate</span><span style="color:#bdd4e3;">(</span><span style="color:#72c3fc;">example</span><span style="color:#329af0;">: </span><span style="color:#c0c5ce;">string</span><span style="color:#bdd4e3;">)</span><span style="color:#329af0;">: </span><span style="color:#c0c5ce;">void {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="10"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#72c3fc;">cp</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">execFileSync</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">./generate-csv</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">, [</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">$CXX -c *.cpp</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">], {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="11"></td><td class="code"><div><span style="color:#c0c5ce;">        env: {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="12"></td><td class="code"><div><span style="color:#c0c5ce;">            ABSROOTDIR: </span><span style="color:#72c3fc;">path</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">resolve</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">`</span><span style="color:#ffb0af;">examples/${</span><span style="color:#72c3fc;">example</span><span style="color:#ffb0af;">}/root</span><span style="color:#bdd4e3;">`</span><span style="color:#c0c5ce;">),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="13"></td><td class="code"><div><span style="color:#c0c5ce;">            ABSOUTDIR: </span><span style="color:#72c3fc;">path</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">resolve</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">`</span><span style="color:#ffb0af;">examples/${</span><span style="color:#72c3fc;">example</span><span style="color:#ffb0af;">}/output</span><span style="color:#bdd4e3;">`</span><span style="color:#c0c5ce;">),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="14"></td><td class="code"><div><span style="color:#c0c5ce;">            CLEAN: </span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">true</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="15"></td><td class="code"><div><span style="color:#c0c5ce;">        },\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="16"></td><td class="code"><div><span style="color:#c0c5ce;">    })\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="17"></td><td class="code"><div><span style="color:#c0c5ce;">}\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="18"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="19"></td><td class="code"><div><span style="color:#329af0;">async function </span><span style="color:#fff3bf;">indexExample</span><span style="color:#bdd4e3;">(</span><span style="color:#72c3fc;">example</span><span style="color:#329af0;">: </span><span style="color:#c0c5ce;">string</span><span style="color:#bdd4e3;">)</span><span style="color:#329af0;">: </span><span style="color:#c0c5ce;">Promise&lt;(Edge </span><span style="color:#329af0;">| </span><span style="color:#c0c5ce;">Vertex)[]&gt; {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="20"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#35a5ff;">if </span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">GENERATE</span><span style="color:#c0c5ce;">) {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="21"></td><td class="code"><div><span style="color:#c0c5ce;">        </span><span style="color:#fff3bf;">generate</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">example</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="22"></td><td class="code"><div><span style="color:#c0c5ce;">    }\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="23"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="24"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">output</span><span style="color:#329af0;">: </span><span style="color:#c0c5ce;">(Edge </span><span style="color:#329af0;">| </span><span style="color:#c0c5ce;">Vertex)[] </span><span style="color:#329af0;">= </span><span style="color:#c0c5ce;">[]\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="25"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="26"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#35a5ff;">await </span><span style="color:#fff3bf;">index</span><span style="color:#c0c5ce;">({\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="27"></td><td class="code"><div><span style="color:#c0c5ce;">        csvFileGlob: </span><span style="color:#bdd4e3;">`</span><span style="color:#ffb0af;">examples/${</span><span style="color:#72c3fc;">example</span><span style="color:#ffb0af;">}/output/*.csv</span><span style="color:#bdd4e3;">`</span><span style="color:#c0c5ce;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="28"></td><td class="code"><div><span style="color:#c0c5ce;">        root: </span><span style="color:#bdd4e3;">`</span><span style="color:#ffb0af;">examples/${</span><span style="color:#72c3fc;">example</span><span style="color:#ffb0af;">}/root</span><span style="color:#bdd4e3;">`</span><span style="color:#c0c5ce;">,\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="29"></td><td class="code"><div><span style="color:#c0c5ce;">        </span><span style="color:#fff3bf;">emit</span><span style="color:#c0c5ce;">: </span><span style="color:#72c3fc;">item </span><span style="color:#329af0;">=&gt;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="30"></td><td class="code"><div><span style="color:#c0c5ce;">            </span><span style="color:#329af0;">new </span><span style="color:#c0c5ce;">Promise(</span><span style="color:#72c3fc;">resolve </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="31"></td><td class="code"><div><span style="color:#c0c5ce;">                </span><span style="color:#72c3fc;">output</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">push</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">item</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="32"></td><td class="code"><div><span style="color:#c0c5ce;">                </span><span style="color:#fff3bf;">resolve</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="33"></td><td class="code"><div><span style="color:#c0c5ce;">            }),\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="34"></td><td class="code"><div><span style="color:#c0c5ce;">    })\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="35"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="36"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#35a5ff;">return </span><span style="color:#72c3fc;">output\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="37"></td><td class="code"><div><span style="color:#c0c5ce;">}\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="38"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="39"></td><td class="code"><div><span style="color:#fff3bf;">test</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">does not emit items with duplicate IDs</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">, </span><span style="color:#329af0;">async </span><span style="color:#bdd4e3;">() </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="40"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">output </span><span style="color:#329af0;">= </span><span style="color:#35a5ff;">await </span><span style="color:#fff3bf;">indexExample</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">five</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="41"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="42"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">setsOfDupes </span><span style="color:#329af0;">= </span><span style="color:#fff3bf;">_</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">output</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="43"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">groupBy</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">item </span><span style="color:#329af0;">=&gt; </span><span style="color:#72c3fc;">item</span><span style="color:#c0c5ce;">.id)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="44"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">values</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="45"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">group </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">({ </span><span style="color:#72c3fc;">group</span><span style="color:#c0c5ce;">, count: </span><span style="color:#72c3fc;">group</span><span style="color:#c0c5ce;">.length }))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="46"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">value</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="47"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">filter</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">(</span><span style="color:#c0c5ce;">{ </span><span style="color:#72c3fc;">count </span><span style="color:#c0c5ce;">}</span><span style="color:#bdd4e3;">) </span><span style="color:#329af0;">=&gt; </span><span style="color:#72c3fc;">count </span><span style="color:#329af0;">&gt; </span><span style="color:#d3f9d8;">1</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="48"></td><td class="code"><div><span style="color:#c0c5ce;">        .</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">(</span><span style="color:#c0c5ce;">{ </span><span style="color:#72c3fc;">group </span><span style="color:#c0c5ce;">}</span><span style="color:#bdd4e3;">) </span><span style="color:#329af0;">=&gt; </span><span style="color:#72c3fc;">group</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="49"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="50"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#35a5ff;">if </span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">setsOfDupes</span><span style="color:#c0c5ce;">.length </span><span style="color:#329af0;">&gt; </span><span style="color:#d3f9d8;">0</span><span style="color:#c0c5ce;">) {\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="51"></td><td class="code"><div><span style="color:#c0c5ce;">        </span><span style="color:#fff3bf;">fail</span><span style="color:#c0c5ce;">(\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="52"></td><td class="code"><div><span style="color:#c0c5ce;">            </span><span style="color:#329af0;">new </span><span style="color:#c0c5ce;">Error(\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="53"></td><td class="code"><div><span style="color:#c0c5ce;">                </span><span style="color:#bdd4e3;">`</span><span style="color:#ffb0af;">Sets of lines with duplicate IDs:</span><span style="color:#96b5b4;">\\n</span><span style="color:#bdd4e3;">` </span><span style="color:#329af0;">+\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="54"></td><td class="code"><div><span style="color:#c0c5ce;">                    </span><span style="color:#72c3fc;">setsOfDupes\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="55"></td><td class="code"><div><span style="color:#c0c5ce;">                        .</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">dupes </span><span style="color:#329af0;">=&gt;\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="56"></td><td class="code"><div><span style="color:#c0c5ce;">                            </span><span style="color:#72c3fc;">dupes</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">item </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">JSON.</span><span style="color:#fff3bf;">stringify</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">item</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">join</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#96b5b4;">\\n</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="57"></td><td class="code"><div><span style="color:#c0c5ce;">                        )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="58"></td><td class="code"><div><span style="color:#c0c5ce;">                        .</span><span style="color:#fff3bf;">join</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#96b5b4;">\\n\\n</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="59"></td><td class="code"><div><span style="color:#c0c5ce;">            )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="60"></td><td class="code"><div><span style="color:#c0c5ce;">        )\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="61"></td><td class="code"><div><span style="color:#c0c5ce;">    }\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="62"></td><td class="code"><div><span style="color:#c0c5ce;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="63"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="64"></td><td class="code"><div><span style="color:#fff3bf;">test</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">five</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">, </span><span style="color:#329af0;">async </span><span style="color:#bdd4e3;">() </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="65"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">output </span><span style="color:#329af0;">= </span><span style="color:#c0c5ce;">(</span><span style="color:#35a5ff;">await </span><span style="color:#fff3bf;">indexExample</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">five</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">JSON.</span><span style="color:#fff3bf;">stringify</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v</span><span style="color:#c0c5ce;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="66"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="67"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#fff3bf;">expect</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">output</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">join</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#96b5b4;">\\n</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">toMatchSnapshot</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="68"></td><td class="code"><div><span style="color:#c0c5ce;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="69"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="70"></td><td class="code"><div><span style="color:#fff3bf;">test</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">cross-app</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">, </span><span style="color:#329af0;">async </span><span style="color:#bdd4e3;">() </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="71"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">app </span><span style="color:#329af0;">= </span><span style="color:#c0c5ce;">(</span><span style="color:#35a5ff;">await </span><span style="color:#fff3bf;">indexExample</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">cross-app</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">JSON.</span><span style="color:#fff3bf;">stringify</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v</span><span style="color:#c0c5ce;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="72"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#fff3bf;">expect</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">app</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">join</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#96b5b4;">\\n</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">toMatchSnapshot</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="73"></td><td class="code"><div><span style="color:#c0c5ce;">})\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="74"></td><td class="code"><div><span style="color:#c0c5ce;">\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="75"></td><td class="code"><div><span style="color:#fff3bf;">test</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">cross-lib</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">, </span><span style="color:#329af0;">async </span><span style="color:#bdd4e3;">() </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">{\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="76"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#329af0;">const </span><span style="color:#72c3fc;">lib </span><span style="color:#329af0;">= </span><span style="color:#c0c5ce;">(</span><span style="color:#35a5ff;">await </span><span style="color:#fff3bf;">indexExample</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#ffb0af;">cross-lib</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">map</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v </span><span style="color:#329af0;">=&gt; </span><span style="color:#c0c5ce;">JSON.</span><span style="color:#fff3bf;">stringify</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">v</span><span style="color:#c0c5ce;">))\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="77"></td><td class="code"><div><span style="color:#c0c5ce;">    </span><span style="color:#fff3bf;">expect</span><span style="color:#c0c5ce;">(</span><span style="color:#72c3fc;">lib</span><span style="color:#c0c5ce;">.</span><span style="color:#fff3bf;">join</span><span style="color:#c0c5ce;">(</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#96b5b4;">\\n</span><span style="color:#bdd4e3;">&#39;</span><span style="color:#c0c5ce;">)).</span><span style="color:#fff3bf;">toMatchSnapshot</span><span style="color:#c0c5ce;">()\n</span></div></td></tr>',
                                          '<tr><td class="line" data-line="78"></td><td class="code"><div><span style="color:#c0c5ce;">})</span></div></td></tr>',
                                      ],
                                  ],
                              },
                          },
                      },
                  },
              },
}

export const symbolSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'symbol',
                name: 'website/src/components/TestimonialCarousel.tsx',
                repository: 'gitlab.sgdev.org/sourcegraph/about',
                branches: [''],
                version: 'b1812108c8c8f0d24c03d69a883060159ebe1ae3',
                symbols: [
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/about/-/blob/website/src/components/TestimonialCarousel.tsx#L22:18-22:29',
                        name: 'Testimonial',
                        containerName: '',
                        kind: 'INTERFACE',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/about/-/blob/website/src/components/TestimonialCarousel.tsx#L36:14-36:33',
                        name: 'TestimonialCarousel',
                        containerName: '',
                        kind: 'VARIABLE',
                    },
                ],
            },
            {
                type: 'symbol',
                name: 'src/characters.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/event-positions',
                branches: [''],
                version: '03f7c3714a1eefe96fdaca48dd234ea3a19224ff',
                symbols: [
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/characters.test.ts#L43:9-43:18',
                        name: 'testcases',
                        containerName: '',
                        kind: 'VARIABLE',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/characters.test.ts#L153:15-153:20',
                        name: 'tests',
                        containerName: '',
                        kind: 'CONSTANT',
                    },
                ],
            },
            {
                type: 'symbol',
                name: 'src/positions_events.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/event-positions',
                branches: [''],
                version: '03f7c3714a1eefe96fdaca48dd234ea3a19224ff',
                symbols: [
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/positions_events.test.ts#L15:9-15:18',
                        name: 'testcases',
                        containerName: '',
                        kind: 'VARIABLE',
                    },
                ],
            },
            {
                type: 'symbol',
                name: 'src/typings/SourcegraphGQL.d.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions',
                branches: [''],
                version: 'f8c71486372087822b7995f0d572c6422b7ae0e5',
                symbols: [
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L493:9-493:31',
                        name: 'lastIndexedRevOrLatest',
                        containerName: 'SourcegraphGQL.IRepository',
                        kind: 'CLASS',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1046:9-1046:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IUser',
                        kind: 'FUNCTION',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1170:9-1170:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IConfigurationSubject',
                        kind: 'ENUM',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1317:9-1317:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IOrg',
                        kind: 'PROPERTY',
                    },
                    {
                        url:
                            '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L4633:9-4633:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.ISite',
                        kind: 'PROPERTY',
                    },
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]
