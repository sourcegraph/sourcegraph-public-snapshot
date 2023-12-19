/* eslint-disable no-template-curly-in-string */
import type {
    SymbolKind,
    HighlightedFileResult,
    SharedGraphQlOperations,
    HighlightedFileVariables,
} from '../../graphql-operations'
import type { SearchEvent } from '../stream'

export const diffSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'commit',
                oid: '65dba23797be9e0ce1941f92c5385a7856bc5a42',
                message: 'build: set up test deps and scripts\n',
                authorName: 'Quinn Slack',
                authorDate: '2019-10-29T20:59:15Z',
                committerName: 'Committer Slack',
                committerDate: '2020-10-29T20:59:15Z',
                url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-lightstep/-/commit/65dba23797be9e0ce1941f92c5385a7856bc5a42',
                repository: 'gitlab.sgdev.org/sourcegraph/sourcegraph-lightstep',
                content:
                    '```diff\nmocha.opts mocha.opts\n@@ -0,0 +3,2 @@\n+--timeout 200\n+src/**/*.test.ts\n\\ No newline at end of file\npackage.json package.json\n@@ -50,0 +54,3 @@\n+    "exclude": [\n+      "**/*.test.ts"\n+    ],\n@@ -54,1 +64,2 @@\n-    "serve": "parcel serve --no-hmr --out-file dist/extension.js src/extension.ts",\n+    "test": "TS_NODE_COMPILER_OPTIONS=\'{\\"module\\":\\"commonjs\\"}\' mocha --require ts-node/register --require source-map-support/register --opts mocha.opts",\n+    "cover": "TS_NODE_COMPILER_OPTIONS=\'{\\"module\\":\\"commonjs\\"}\' nyc --require ts-node/register --require source-map-support/register --all mocha --opts mocha.opts --timeout 10000",\n@@ -57,2 +70,2 @@\n-    "sourcegraph:prepublish": "parcel build src/extension.ts"\n+    "sourcegraph:prepublish": "yarn typecheck && yarn test && yarn build"\n   },\nyarn.lock yarn.lock\n@@ -3736,0 +4204,3 @@ number-is-nan@^1.0.0:\n+    spawn-wrap "^1.4.2"\n+    test-exclude "^5.1.0"\n+    uuid "^3.3.2"\n@@ -5550,1 +6166,5 @@ terser@^3.7.3, terser@^3.8.1:\n \n+test-exclude@^5.1.0:\n+  version "5.1.0"\n+  resolved "https://registry.yarnpkg.com/test-exclude/-/test-exclude-5.1.0.tgz#6ba6b25179d2d38724824661323b73e03c0c1de1"\n+  integrity sha512-gwf0S2fFsANC55fSeSqpb8BYk6w3FDvwZxfNjeF6FRgvFa43r+7wRiA/Q0IxoRU37wB/LE8IQ4221BsNucTaCA==\n```',
                ranges: [
                    [4, 10, 4],
                    [9, 13, 4],
                    [13, 6, 4],
                    [17, 55, 4],
                    [22, 5, 4],
                    [26, 1, 4],
                    [28, 42, 4],
                    [28, 57, 4],
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

/*
    a11y-ignore
    Rule: "color-contrast" (Elements must have sufficient color contrast) for all changes in this file
    GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
*/
export const diffHighlightResult: Partial<SharedGraphQlOperations> = {
    highlightCode: () => ({
        highlightCode:
            '<table><tbody><tr><td class="line" data-line="1"/><td class="code"><div><span class="a11y-ignore hl-source hl-diff">mocha.opts mocha.opts\n</span></div></td></tr><tr><td class="line" data-line="2"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-0,0 +3,2</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="3"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>--timeout 200\n</span></span></div></td></tr><tr><td class="line" data-line="4"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>src/**/*.test.ts\n</span></span></div></td></tr><tr><td class="line" data-line="5"/><td class="code"><div><span class="a11y-ignore hl-source hl-diff">\\ No newline at end of file\n</span></div></td></tr><tr><td class="line" data-line="6"/><td class="code"><div><span class="a11y-ignore hl-source hl-diff">package.json package.json\n</span></div></td></tr><tr><td class="line" data-line="7"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-50,0 +54,3</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="8"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    &quot;exclude&quot;: [\n</span></span></div></td></tr><tr><td class="line" data-line="9"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>      &quot;**/*.test.ts&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="10"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    ],\n</span></span></div></td></tr><tr><td class="line" data-line="11"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-54,1 +64,2</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="12"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-deleted hl-diff"><span class="hl-punctuation hl-definition hl-deleted hl-diff">-</span>    &quot;serve&quot;: &quot;parcel serve --no-hmr --out-file dist/extension.js src/extension.ts&quot;,\n</span></span></div></td></tr><tr><td class="line" data-line="13"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    &quot;test&quot;: &quot;TS_NODE_COMPILER_OPTIONS=&#39;{\\&quot;module\\&quot;:\\&quot;commonjs\\&quot;}&#39; mocha --require ts-node/register --require source-map-support/register --opts mocha.opts&quot;,\n</span></span></div></td></tr><tr><td class="line" data-line="14"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    &quot;cover&quot;: &quot;TS_NODE_COMPILER_OPTIONS=&#39;{\\&quot;module\\&quot;:\\&quot;commonjs\\&quot;}&#39; nyc --require ts-node/register --require source-map-support/register --all mocha --opts mocha.opts --timeout 10000&quot;,\n</span></span></div></td></tr><tr><td class="line" data-line="15"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-57,2 +70,2</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="16"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-deleted hl-diff"><span class="hl-punctuation hl-definition hl-deleted hl-diff">-</span>    &quot;sourcegraph:prepublish&quot;: &quot;parcel build src/extension.ts&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="17"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    &quot;sourcegraph:prepublish&quot;: &quot;yarn typecheck &amp;&amp; yarn test &amp;&amp; yarn build&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="18"/><td class="code"><div><span class="hl-source hl-diff">   },\n</span></div></td></tr><tr><td class="line" data-line="19"/><td class="code"><div><span class="a11y-ignore hl-source hl-diff">yarn.lock yarn.lock\n</span></div></td></tr><tr><td class="line" data-line="20"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-3736,0 +4204,3</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-entity hl-name hl-section hl-diff">number-is-nan@^1.0.0:</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="21"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    spawn-wrap &quot;^1.4.2&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="22"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    test-exclude &quot;^5.1.0&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="23"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>    uuid &quot;^3.3.2&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="24"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-meta hl-diff hl-range hl-unified"><span class="hl-meta hl-range hl-unified hl-diff"><span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-meta hl-toc-list hl-line-number hl-diff">-5550,1 +6166,5</span> <span class="hl-punctuation hl-definition hl-range hl-diff">@@</span> <span class="a11y-ignore hl-entity hl-name hl-section hl-diff">terser@^3.7.3, terser@^3.8.1:</span>\n</span></span></span></div></td></tr><tr><td class="line" data-line="25"/><td class="code"><div><span class="hl-source hl-diff"> \n</span></div></td></tr><tr><td class="line" data-line="26"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>test-exclude@^5.1.0:\n</span></span></div></td></tr><tr><td class="line" data-line="27"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>  version &quot;5.1.0&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="28"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>  resolved &quot;https://registry.yarnpkg.com/test-exclude/-/test-exclude-5.1.0.tgz#6ba6b25179d2d38724824661323b73e03c0c1de1&quot;\n</span></span></div></td></tr><tr><td class="line" data-line="29"/><td class="code"><div><span class="hl-source hl-diff"><span class="hl-markup hl-inserted hl-diff"><span class="hl-punctuation hl-definition hl-inserted hl-diff">+</span>  integrity sha512-gwf0S2fFsANC55fSeSqpb8BYk6w3FDvwZxfNjeF6FRgvFa43r+7wRiA/Q0IxoRU37wB/LE8IQ4221BsNucTaCA==</span></span></div></td></tr></tbody></table>',
    }),
}

export const commitSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'commit',
                oid: '7e69ceb49adc30cb46bbe50335e1a371a0f2f6b1',
                message: 'add more tests, use the Sourcegraph stubs api and improve repo matching. (#13)',
                authorName: 'Vanesa',
                authorDate: '2019-10-29T20:59:15Z',
                committerName: 'Committer Vanesa',
                committerDate: '2020-10-29T20:59:15Z',
                url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-sentry/-/commit/7e69ceb49adc30cb46bbe50335e1a371a0f2f6b1',
                repository: 'gitlab.sgdev.org/sourcegraph/sourcegraph-sentry',
                content:
                    '```COMMIT_EDITMSG\nadd more tests, use the Sourcegraph stubs api and improve repo matching. (#13)\n\n* add more tests, refactor to use extension api stubs\r\n* improve repo matching\r\nCo-Authored-By: Felix Becker <felix.b@outlook.com>\n```',
                ranges: [
                    [1, 9, 4],
                    [3, 11, 4],
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

export const commitHighlightResult: Partial<SharedGraphQlOperations> = {
    highlightCode: () => ({
        highlightCode:
            '<table><tbody><tr><td class="line" data-line="1"/><td class="code"><div><span class="a11y-ignore hl-text hl-plain">add more tests, use the Sourcegraph stubs api and improve repo matching. (#13)\n</span></div></td></tr><tr><td class="line" data-line="2"/><td class="code"><div><span class="hl-text hl-plain">\n</span></div></td></tr><tr><td class="line" data-line="3"/><td class="code"><div><span class="hl-text hl-plain">* add more tests, refactor to use extension api stubs\n</span></div></td></tr><tr><td class="line" data-line="4"/><td class="code"><div><span class="a11y-ignore hl-text hl-plain">* improve repo matching\n</span></div></td></tr><tr><td class="line" data-line="5"/><td class="code"><div><span class="a11y-ignore hl-text hl-plain">Co-Authored-By: Felix Becker</span></div></td></tr></tbody></table>',
    }),
}

export const mixedSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            { type: 'repo', repository: 'gitlab.sgdev.org/lg-test-private/lg-test' },
            {
                type: 'path',
                path: 'overridable/bool_or_string_test.go',
                repository: 'gitlab.sgdev.org/aharvey/batch-change-utils',
                branches: [''],
                commit: '206c057cc03eea48300a4bd33f4dc4222d242114',
            },
            {
                type: 'content',
                path: 'src/main.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/lsif-cpp',
                branches: [''],
                commit: '2e3569cf60646c9ce4e37a43e5cf698a00cbd41a',
                chunkMatches: [
                    {
                        content:
                            "\ntest('does not emit items with duplicate IDs', async () => {\n\tconst output = await indexExample('five')",
                        contentStart: {
                            offset: 938,
                            line: 37,
                            column: 0,
                        },
                        ranges: [
                            {
                                start: {
                                    offset: 939,
                                    line: 38,
                                    column: 0,
                                },
                                end: {
                                    offset: 943,
                                    line: 38,
                                    column: 4,
                                },
                            },
                        ],
                    },
                    {
                        content:
                            "\ntest('five', async () => {\n\tconst output = (await indexExample('five')).map(v => JSON.stringify(v))",
                        contentStart: {
                            offset: 1657,
                            line: 62,
                            column: 0,
                        },
                        ranges: [
                            {
                                start: {
                                    offset: 1658,
                                    line: 63,
                                    column: 0,
                                },
                                end: {
                                    offset: 1662,
                                    line: 63,
                                    column: 4,
                                },
                            },
                        ],
                    },
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

export const highlightFileResult = {
    HighlightedFile: ((parameters: HighlightedFileVariables): HighlightedFileResult => {
        const allLines = [
            '<tr><td class="line" data-line="1"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-import hl-ts"><span class="hl-keyword hl-control hl-import hl-ts">import</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">index</span> <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span> <span class="hl-keyword hl-control hl-from hl-ts">from</span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>./index<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="2"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-import hl-ts"><span class="hl-keyword hl-control hl-import hl-ts">import</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">Edge</span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">Vertex</span> <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span> <span class="hl-keyword hl-control hl-from hl-ts">from</span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>lsif-protocol<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="3"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-import hl-ts"><span class="hl-keyword hl-control hl-import hl-ts">import</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">_</span> <span class="hl-keyword hl-control hl-from hl-ts">from</span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>lodash<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="4"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-import hl-ts"><span class="hl-keyword hl-control hl-import hl-ts">import</span> <span class="hl-constant hl-language hl-import-export-all hl-ts">*</span> <span class="hl-keyword hl-control hl-as hl-ts">as</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">path</span> <span class="hl-keyword hl-control hl-from hl-ts">from</span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>path<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="5"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-import hl-ts"><span class="hl-keyword hl-control hl-import hl-ts">import</span> <span class="hl-constant hl-language hl-import-export-all hl-ts">*</span> <span class="hl-keyword hl-control hl-as hl-ts">as</span> <span class="hl-variable hl-other hl-readwrite hl-alias hl-ts">cp</span> <span class="hl-keyword hl-control hl-from hl-ts">from</span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>child_process<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="6"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="7"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts">GENERATE</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-constant hl-language hl-boolean hl-false hl-ts">false</span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="8"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="9"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-storage hl-type hl-function hl-ts">function</span> <span class="hl-meta hl-definition hl-function hl-ts"><span class="hl-entity hl-name hl-function hl-ts">generate</span></span><span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-variable hl-parameter hl-ts">example</span><span class="hl-meta hl-type hl-annotation hl-ts"><span class="hl-keyword hl-operator hl-type hl-annotation hl-ts">:</span> <span class="hl-support hl-type hl-primitive hl-ts">string</span></span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span><span class="hl-meta hl-return hl-type hl-ts"><span class="hl-keyword hl-operator hl-type hl-annotation hl-ts">:</span> <span class="hl-support hl-type hl-primitive hl-ts">void</span> </span><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="10"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">cp</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">execFileSync</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>./generate-csv<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span><span class="hl-meta hl-array hl-literal hl-ts"> <span class="hl-meta hl-brace hl-square hl-ts">[</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>$CXX -c *.cpp<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-square hl-ts">]</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-objectliteral hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="11"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">        <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">env</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-meta hl-objectliteral hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="12"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-objectliteral hl-ts">            <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">ABSROOTDIR</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">path</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-promise hl-ts">resolve</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-template hl-ts"><span class="hl-punctuation hl-definition hl-string hl-template hl-begin hl-ts">`</span>examples/<span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-punctuation hl-definition hl-template-expression hl-begin hl-ts">${</span></span><span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-meta hl-embedded hl-line hl-ts"><span class="hl-variable hl-other hl-readwrite hl-ts">example</span></span><span class="hl-punctuation hl-definition hl-template-expression hl-end hl-ts">}</span></span>/root<span class="hl-punctuation hl-definition hl-string hl-template hl-end hl-ts">`</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="13"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-objectliteral hl-ts">            <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">ABSOUTDIR</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">path</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-promise hl-ts">resolve</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-template hl-ts"><span class="hl-punctuation hl-definition hl-string hl-template hl-begin hl-ts">`</span>examples/<span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-punctuation hl-definition hl-template-expression hl-begin hl-ts">${</span></span><span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-meta hl-embedded hl-line hl-ts"><span class="hl-variable hl-other hl-readwrite hl-ts">example</span></span><span class="hl-punctuation hl-definition hl-template-expression hl-end hl-ts">}</span></span>/output<span class="hl-punctuation hl-definition hl-string hl-template hl-end hl-ts">`</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="14"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-objectliteral hl-ts">            <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">CLEAN</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>true<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="15"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-objectliteral hl-ts">        <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="16"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">    <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="17"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="18"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="19"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-storage hl-modifier hl-async hl-ts">async</span> <span class="hl-storage hl-type hl-function hl-ts">function</span> <span class="hl-meta hl-definition hl-function hl-ts"><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">indexExample</span></span><span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-variable hl-parameter hl-ts">example</span><span class="hl-meta hl-type hl-annotation hl-ts"><span class="hl-keyword hl-operator hl-type hl-annotation hl-ts">:</span> <span class="hl-support hl-type hl-primitive hl-ts">string</span></span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span><span class="hl-meta hl-return hl-type hl-ts"><span class="hl-keyword hl-operator hl-type hl-annotation hl-ts">:</span> <span class="hl-entity hl-name hl-type hl-ts">Promise</span><span class="hl-meta hl-type hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-typeparameters hl-begin hl-ts">&lt;</span></span><span class="hl-meta hl-type hl-parameters hl-ts"><span class="hl-meta hl-type hl-paren hl-cover hl-ts"><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-entity hl-name hl-type hl-ts">Edge</span> <span class="hl-keyword hl-operator hl-type hl-ts">|</span> <span class="hl-entity hl-name hl-type hl-ts">Vertex</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span><span class="hl-meta hl-type hl-tuple hl-ts"><span class="hl-meta hl-brace hl-square hl-ts">[</span><span class="hl-meta hl-brace hl-square hl-ts">]</span></span></span><span class="hl-meta hl-type hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-typeparameters hl-end hl-ts">&gt;</span></span> </span><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="20"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-keyword hl-control hl-conditional hl-ts">if</span> <span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-constant hl-ts">GENERATE</span><span class="hl-meta hl-brace hl-round hl-ts">)</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="21"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">generate</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">example</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="22"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="23"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="24"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts a11y-ignore">output</span></span><span class="hl-meta hl-type hl-annotation hl-ts"><span class="hl-keyword hl-operator hl-type hl-annotation hl-ts">:</span> <span class="hl-meta hl-type hl-paren hl-cover hl-ts"><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-entity hl-name hl-type hl-ts">Edge</span> <span class="hl-keyword hl-operator hl-type hl-ts">|</span> <span class="hl-entity hl-name hl-type hl-ts">Vertex</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span><span class="hl-meta hl-type hl-tuple hl-ts"><span class="hl-meta hl-brace hl-square hl-ts">[</span><span class="hl-meta hl-brace hl-square hl-ts">]</span></span> </span></span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span><span class="hl-meta hl-array hl-literal hl-ts"> <span class="hl-meta hl-brace hl-square hl-ts">[</span><span class="hl-meta hl-brace hl-square hl-ts">]</span></span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="25"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="26"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-keyword hl-control hl-flow hl-ts a11y-ignore">await</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">index</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="27"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">        <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">csvFileGlob</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-string hl-template hl-ts"><span class="hl-punctuation hl-definition hl-string hl-template hl-begin hl-ts">`</span>examples/<span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-punctuation hl-definition hl-template-expression hl-begin hl-ts">${</span></span><span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-meta hl-embedded hl-line hl-ts"><span class="hl-variable hl-other hl-readwrite hl-ts">example</span></span><span class="hl-punctuation hl-definition hl-template-expression hl-end hl-ts">}</span></span>/output/*.csv<span class="hl-punctuation hl-definition hl-string hl-template hl-end hl-ts">`</span></span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="28"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">        <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">root</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-string hl-template hl-ts"><span class="hl-punctuation hl-definition hl-string hl-template hl-begin hl-ts">`</span>examples/<span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-punctuation hl-definition hl-template-expression hl-begin hl-ts">${</span></span><span class="hl-meta hl-template hl-expression hl-ts"><span class="hl-meta hl-embedded hl-line hl-ts"><span class="hl-variable hl-other hl-readwrite hl-ts">example</span></span><span class="hl-punctuation hl-definition hl-template-expression hl-end hl-ts">}</span></span>/root<span class="hl-punctuation hl-definition hl-string hl-template hl-end hl-ts">`</span></span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="29"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">        <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-entity hl-name hl-function hl-ts">emit</span></span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">item</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="30"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-arrow hl-ts">            </span><span class="hl-new hl-expr hl-ts"><span class="hl-keyword hl-operator hl-new hl-ts">new</span> <span class="hl-entity hl-name hl-type hl-ts">Promise</span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">resolve</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="31"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-new hl-expr hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">                <span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">output</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">push</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">item</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="32"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-new hl-expr hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">                <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">resolve</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="33"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-new hl-expr hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">            <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="34"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-objectliteral hl-ts">    <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="35"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="36"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-keyword hl-control hl-flow hl-ts">return</span> <span class="hl-variable hl-other hl-readwrite hl-ts">output</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="37"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="38"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="39"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">test</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>does not emit items with duplicate IDs<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-modifier hl-async hl-ts">async</span> <span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="40"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts a11y-ignore">output</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-keyword hl-control hl-flow hl-ts a11y-ignore">await</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">indexExample</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>five<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="41"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="42"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts">setsOfDupes</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">_</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">output</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="43"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">groupBy</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">item</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-variable hl-other hl-object hl-ts">item</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-variable hl-property hl-dom hl-ts">id</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="44"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-dom hl-ts">values</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="45"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">group</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-objectliteral hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span> <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-variable hl-other hl-readwrite hl-ts">group</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts">count</span></span><span class="hl-meta hl-object hl-member hl-ts"><span class="hl-meta hl-object-literal hl-key hl-ts"><span class="hl-punctuation hl-separator hl-key-value hl-ts">:</span></span> <span class="hl-variable hl-other hl-object hl-ts">group</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-variable hl-property hl-ts">length</span> </span><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="46"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">value</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="47"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">filter</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-meta hl-parameter hl-object-binding-pattern hl-ts"><span class="hl-punctuation hl-definition hl-binding-pattern hl-object hl-ts">{</span> <span class="hl-variable hl-parameter hl-ts">count</span> <span class="hl-punctuation hl-definition hl-binding-pattern hl-object hl-ts">}</span></span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-variable hl-other hl-readwrite hl-ts">count</span> <span class="hl-keyword hl-operator hl-relational hl-ts">&gt;</span> <span class="hl-constant hl-numeric hl-decimal hl-ts">1</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="48"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-meta hl-parameter hl-object-binding-pattern hl-ts"><span class="hl-punctuation hl-definition hl-binding-pattern hl-object hl-ts">{</span> <span class="hl-variable hl-parameter hl-ts">group</span> <span class="hl-punctuation hl-definition hl-binding-pattern hl-object hl-ts">}</span></span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-variable hl-other hl-readwrite hl-ts">group</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="49"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="50"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-keyword hl-control hl-conditional hl-ts">if</span> <span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-object hl-ts">setsOfDupes</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-variable hl-property hl-ts">length</span> <span class="hl-keyword hl-operator hl-relational hl-ts">&gt;</span> <span class="hl-constant hl-numeric hl-decimal hl-ts">0</span><span class="hl-meta hl-brace hl-round hl-ts">)</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="51"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">fail</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="52"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">            <span class="hl-new hl-expr hl-ts"><span class="hl-keyword hl-operator hl-new hl-ts">new</span> <span class="hl-entity hl-name hl-type hl-ts">Error</span><span class="hl-meta hl-brace hl-round hl-ts">(</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="53"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">                <span class="hl-string hl-template hl-ts"><span class="hl-punctuation hl-definition hl-string hl-template hl-begin hl-ts">`</span>Sets of lines with duplicate IDs:<span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-template hl-end hl-ts">`</span></span> <span class="hl-keyword hl-operator hl-arithmetic hl-ts">+</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="54"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">                    <span class="hl-variable hl-other hl-readwrite hl-ts">setsOfDupes</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="55"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">                        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">dupes</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span>\n</span></span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="56"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts"><span class="hl-meta hl-arrow hl-ts">                            </span><span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">dupes</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">item</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-meta hl-function-call hl-ts"><span class="hl-support hl-constant hl-json hl-ts a11y-ignore">JSON</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-json hl-ts a11y-ignore">stringify</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">item</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">join</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="57"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">                        <span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="58"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">                        <span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">join</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="59"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-new hl-expr hl-ts">            <span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="60"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">        <span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="61"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="62"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="63"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="64"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">test</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>five<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-modifier hl-async hl-ts">async</span> <span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="65"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts a11y-ignore">output</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-keyword hl-control hl-flow hl-ts a11y-ignore">await</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">indexExample</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>five<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">v</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-meta hl-function-call hl-ts"><span class="hl-support hl-constant hl-json hl-ts a11y-ignore">JSON</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-json hl-ts a11y-ignore">stringify</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">v</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="66"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="67"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">expect</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">output</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">join</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">toMatchSnapshot</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="68"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="69"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="70"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">test</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>cross-app<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-modifier hl-async hl-ts">async</span> <span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="71"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts">app</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-keyword hl-control hl-flow hl-ts a11y-ignore">await</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">indexExample</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>cross-app<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">v</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-meta hl-function-call hl-ts"><span class="hl-support hl-constant hl-json hl-ts a11y-ignore">JSON</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-json hl-ts a11y-ignore">stringify</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">v</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="72"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">expect</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">app</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">join</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">toMatchSnapshot</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="73"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></div></td></tr>',
            '<tr><td class="line" data-line="74"></td><td class="code"><div><span class="hl-source hl-ts">\n</span></div></td></tr>',
            '<tr><td class="line" data-line="75"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">test</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>cross-lib<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-punctuation hl-separator hl-comma hl-ts">,</span> <span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-modifier hl-async hl-ts">async</span> <span class="hl-meta hl-parameters hl-ts"><span class="hl-punctuation hl-definition hl-parameters hl-begin hl-ts">(</span><span class="hl-punctuation hl-definition hl-parameters hl-end hl-ts">)</span></span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> <span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">{</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="76"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-var hl-expr hl-ts"><span class="hl-storage hl-type hl-ts a11y-ignore">const</span> <span class="hl-meta hl-var-single-variable hl-expr hl-ts"><span class="hl-meta hl-definition hl-variable hl-ts"><span class="hl-variable hl-other hl-constant hl-ts">lib</span></span> </span><span class="hl-keyword hl-operator hl-assignment hl-ts">=</span> <span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-keyword hl-control hl-flow hl-ts a11y-ignore">await</span> <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">indexExample</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span>cross-lib<span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts a11y-ignore">map</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-arrow hl-ts"><span class="hl-variable hl-parameter hl-ts">v</span> </span><span class="hl-meta hl-arrow hl-ts"><span class="hl-storage hl-type hl-function hl-arrow hl-ts">=&gt;</span> </span><span class="hl-meta hl-function-call hl-ts"><span class="hl-support hl-constant hl-json hl-ts a11y-ignore">JSON</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-json hl-ts a11y-ignore">stringify</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-variable hl-other hl-readwrite hl-ts">v</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="77"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts">    <span class="hl-meta hl-function-call hl-ts"><span class="hl-entity hl-name hl-function hl-ts">expect</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-variable hl-other hl-object hl-ts">lib</span><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-support hl-function hl-ts">join</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-string hl-quoted hl-single hl-ts a11y-ignore"><span class="hl-punctuation hl-definition hl-string hl-begin hl-ts">&#39;</span><span class="hl-constant hl-character hl-escape hl-ts">\\n</span><span class="hl-punctuation hl-definition hl-string hl-end hl-ts">&#39;</span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-brace hl-round hl-ts">)</span><span class="hl-meta hl-function-call hl-ts"><span class="hl-punctuation hl-accessor hl-ts">.</span><span class="hl-entity hl-name hl-function hl-ts">toMatchSnapshot</span></span><span class="hl-meta hl-brace hl-round hl-ts">(</span><span class="hl-meta hl-brace hl-round hl-ts">)</span>\n</span></span></span></div></td></tr>',
            '<tr><td class="line" data-line="78"></td><td class="code"><div><span class="hl-source hl-ts"><span class="hl-meta hl-arrow hl-ts"><span class="hl-meta hl-block hl-ts"><span class="hl-punctuation hl-definition hl-block hl-ts">}</span></span></span><span class="hl-meta hl-brace hl-round hl-ts">)</span></span></div></td></tr>',
        ]

        const lineRanges = parameters.ranges.map(range => allLines.slice(range.startLine, range.endLine))

        return {
            repository: {
                commit: {
                    file: {
                        isDirectory: false,
                        richHTML: '',
                        highlight: {
                            aborted: false,
                            /*
                            a11y-ignore
                            Rule: "color-contrast" (Elements must have sufficient color contrast)
                            GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                        */
                            lineRanges,
                        },
                    },
                },
            },
        }
    }) as SharedGraphQlOperations['HighlightedFile'],
}

export const symbolSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'symbol',
                path: 'website/src/components/TestimonialCarousel.tsx',
                repository: 'gitlab.sgdev.org/sourcegraph/about',
                branches: [''],
                commit: 'b1812108c8c8f0d24c03d69a883060159ebe1ae3',
                symbols: [
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/about/-/blob/website/src/components/TestimonialCarousel.tsx#L22:18-22:29',
                        name: 'Testimonial',
                        containerName: '',
                        kind: 'INTERFACE' as SymbolKind,
                        line: 22,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/about/-/blob/website/src/components/TestimonialCarousel.tsx#L36:14-36:33',
                        name: 'TestimonialCarousel',
                        containerName: '',
                        kind: 'VARIABLE' as SymbolKind,
                        line: 36,
                    },
                ],
            },
            {
                type: 'symbol',
                path: 'src/characters.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/event-positions',
                branches: [''],
                commit: '03f7c3714a1eefe96fdaca48dd234ea3a19224ff',
                symbols: [
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/characters.test.ts#L43:9-43:18',
                        name: 'testcases',
                        containerName: '',
                        kind: 'VARIABLE' as SymbolKind,
                        line: 43,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/characters.test.ts#L153:15-153:20',
                        name: 'tests',
                        containerName: '',
                        kind: 'CONSTANT' as SymbolKind,
                        line: 153,
                    },
                ],
            },
            {
                type: 'symbol',
                path: 'src/positions_events.test.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/event-positions',
                branches: [''],
                commit: '03f7c3714a1eefe96fdaca48dd234ea3a19224ff',
                symbols: [
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/event-positions/-/blob/src/positions_events.test.ts#L15:9-15:18',
                        name: 'testcases',
                        containerName: '',
                        kind: 'VARIABLE' as SymbolKind,
                        line: 15,
                    },
                ],
            },
            {
                type: 'symbol',
                path: 'src/typings/SourcegraphGQL.d.ts',
                repository: 'gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions',
                branches: [''],
                commit: 'f8c71486372087822b7995f0d572c6422b7ae0e5',
                symbols: [
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L493:9-493:31',
                        name: 'lastIndexedRevOrLatest',
                        containerName: 'SourcegraphGQL.IRepository',
                        kind: 'CLASS' as SymbolKind,
                        line: 493,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1046:9-1046:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IUser',
                        kind: 'FUNCTION' as SymbolKind,
                        line: 1046,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1170:9-1170:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IConfigurationSubject',
                        kind: 'ENUM' as SymbolKind,
                        line: 1170,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L1317:9-1317:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.IOrg',
                        kind: 'PROPERTY' as SymbolKind,
                        line: 1317,
                    },
                    {
                        url: '/gitlab.sgdev.org/sourcegraph/sourcegraph-code-discussions/-/blob/src/typings/SourcegraphGQL.d.ts#L4633:9-4633:23',
                        name: 'latestSettings',
                        containerName: 'SourcegraphGQL.ISite',
                        kind: 'PROPERTY' as SymbolKind,
                        line: 4633,
                    },
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

export const ownerSearchStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'person',
                handle: 'example-unknown',
            },
            {
                type: 'person',
                email: 'test@example.com',
            },
            {
                type: 'person',
                handle: 'example-user',
                email: 'user@example.com',
                user: {
                    username: 'example-user',
                    displayName: 'Example User',
                    avatarURL:
                        'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAaQAAAGkAQMAAABEgsN2AAAABlBMVEWD1qfw8PD3jxSNAAAAiklEQVR42u3cMQrAIAwFUKEH8/638AgdO9UOkl0qosL7Y+CR+RNIqj9yUxRFURTVoUrqT6YoiqIoiqIoiqKoc1WuLW8MrqjIhaIoiqIoiqIoiqJG1LzmGyseiqIoiqIoiqIoihpRu11vKYqiKIqiKIqiKIqiKIqiKIqiKIqiqKYO+lVFURRFUcvUB5Q0nNnSPJFEAAAAAElFTkSuQmCC',
                },
            },
            {
                type: 'team',
                name: 'example-team',
                displayName: 'Example Team',
            },
        ],
    },
    { type: 'done', data: {} },
]
