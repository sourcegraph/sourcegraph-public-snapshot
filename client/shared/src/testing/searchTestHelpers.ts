import { noop } from 'lodash'
import { EMPTY, NEVER, of, Subscription } from 'rxjs'
import sinon from 'sinon'

import { FetchFileParameters } from '@sourcegraph/search-ui'

import { FlatExtensionHostAPI } from '../api/contract'
import { pretendProxySubscribable, pretendRemote } from '../api/util'
import { Controller } from '../extensions/controller'
import { PlatformContext } from '../platform/context'
import { AggregateStreamingSearchResults, ContentMatch, RepositoryMatch } from '../search/stream'

export const RESULT: ContentMatch = {
    type: 'content',
    path: '.travis.yml',
    repository: 'github.com/golang/oauth2',
    chunkMatches: [
        {
            content: '  - go test -v golang.org/x/oauth2/...',
            contentStart: {
                line: 12,
                offset: 12,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 12,
                        offset: 7,
                        column: 7,
                    },
                    end: {
                        line: 12,
                        offset: 11,
                        column: 11,
                    },
                }
            ]
        }
    ],
}

export const REPO_MATCH_RESULT: RepositoryMatch = {
    type: 'repo',
    repository: 'github.com/golang/oauth2',
}

export const REPO_MATCH_RESULTS_WITH_METADATA: RepositoryMatch[] = [
    {
        type: 'repo',
        repository: 'github.com/golang/oauth2',
        description: 'The Go package for OAuth2.',
    },
    {
        type: 'repo',
        repository: 'github.com/sourcegraph/sourcegraph',
        description: 'Universtal code search',
        repoStars: 123,
        repoLastFetched: '2017-01-01T00:00:00Z',
        private: true,
    },
    {
        type: 'repo',
        repository: 'github.com/sourcegraph/go-langserver',
        description: 'Go language server',
        repoStars: 9000,
        fork: true,
        archived: true,
    },
]

export const MULTIPLE_MATCH_RESULT: ContentMatch = {
    type: 'content',
    path: 'clientcredentials/clientcredentials_test.go',
    repository: 'github.com/golang/oauth2',
    chunkMatches: [
        {
            content: '\t"net/http/httptest"',
            contentStart: {
                line: 11,
                offset: 11,
                column: 0,
            },
            ranges: [
                {
                    start: {
                       line: 11,
                       offset: 15,
                       column: 15,
                    },
                    end: {
                        line: 11,
                        offset: 19,
                        column: 19,
                    },
                },
            ],
        },
        {
            content: '\t"testing"',
            contentStart: {
                line: 13,
                offset: 13,
                column: 0,
            },
            ranges: [
                {
                    start: {
                       line: 13,
                       offset: 2,
                       column: 2,
                    },
                    end: {
                        line: 13,
                        offset: 6,
                        column: 6,
                    },
                },
            ],
        },
        {
            content: 'func TestTokenSourceGrantTypeOverride(t *testing.T) {',
            contentStart: {
                line: 36,
                offset: 36,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 36,
                        offset: 5,
                        column: 5,
                    },
                    end: {
                        line: 36,
                        offset: 9,
                        column: 9,
                    },
                },
                {
                    start: {
                        line: 36,
                        offset: 41,
                        column: 41,
                    },
                    end: {
                        line: 36,
                        offset: 45,
                        column: 45,
                    },
                },
            ],
        },
        {
            content: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            contentStart: {
                line: 39,
                offset: 39,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 39,
                        offset: 11,
                        column: 11,
                    },
                    end: {
                        line: 39,
                        offset: 15,
                        column: 15,
                    },
                },
            ],
        },
        {
           content: 'func TestTokenRequest(t *testing.T) {',
            contentStart: {
               line: 73,
                offset: 73,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 73,
                        offset: 5,
                        column: 5,
                    },
                    end: {
                        line: 73,
                        offset: 9,
                        column: 9,
                    },
                },
                {
                    start: {
                        line: 73,
                        offset: 25,
                        column: 25,
                    },
                    end: {
                        line: 73,
                        offset: 29,
                        column: 29,
                    },
                },
            ],
        },
        {
            content: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            contentStart: {
                line: 74,
                offset: 74,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 74,
                        offset: 11,
                        column: 11,
                    },
                    end: {
                        line: 74,
                        offset: 15,
                        column: 15,
                    },
                },
            ],
        },
        {
            content: 'func TestTokenRefreshRequest(t *testing.T) {',
            contentStart: {
                line: 115,
                offset: 115,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 115,
                        offset: 5,
                        column: 5,
                    },
                    end: {
                        line: 115,
                        offset: 9,
                        column: 9,
                    },
                },
            ],
        },
        {
            content: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            contentStart: {
                line: 117,
                offset: 117,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 117,
                        offset: 11,
                        column: 11,
                    },
                    end: {
                        line: 117,
                        offset: 15,
                        column: 15,
                    },
                },
            ],
        },
        {
            content: '\t\tio.WriteString(w, `{"access_token": "foo", "refresh_token": "bar"}`)',
            contentStart: {
                line: 134,
                offset: 134,
                column: 0,
            },
            ranges: [
                {
                    start: {
                        line: 134,
                        offset: 8,
                        column: 8,
                    },
                    end: {
                        line: 134,
                        offset: 12,
                        column: 12,
                    },
                },
            ],
        },
    ],
}

export const SEARCH_RESULT: AggregateStreamingSearchResults = {
    state: 'complete',
    progress: {
        durationMs: 78,
        matchCount: 1,
        skipped: [],
    },
    filters: [
        { value: 'file:\\.yml$', label: 'YAML', count: 1, limitHit: false, kind: 'file' },
        { value: 'case:yes', label: 'Make search case sensitive', count: 0, limitHit: false, kind: 'utility' },
        {
            value: 'repo:^github\\.com/golang/oauth2$',
            label: 'github.com/golang/oauth2',
            count: 1,
            limitHit: false,
            kind: 'repo',
        },
    ],
    results: [RESULT],
}

export const MULTIPLE_SEARCH_RESULT: AggregateStreamingSearchResults = {
    ...SEARCH_RESULT,
    progress: {
        durationMs: 78,
        matchCount: 136,
        skipped: [],
    },
    results: [
        RESULT,
        MULTIPLE_MATCH_RESULT,
        {
            type: 'content',
            path: 'example_test.go',
            commit: 'abcd1234',
            repository: 'github.com/golang/oauth2',
            chunkMatches: [
                {
                    content: 'package oauth2_test',
                    contentStart: {
                        line: 4,
                        offset: 4,
                        column: 0,
                    },
                    ranges: [
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                    ],
                },
            ],
            repoStars: 42,
            repoLastFetched: '2017-01-01T00:00:00Z',
        },
    ],
}

export const COLLAPSABLE_SEARCH_RESULT: AggregateStreamingSearchResults = {
    ...SEARCH_RESULT,
    progress: {
        durationMs: 78,
        matchCount: 136,
        skipped: [],
    },
    results: [
        RESULT,
        MULTIPLE_MATCH_RESULT,
        {
            type: 'content',
            path: 'example_test.go',
            commit: 'abcd1234',
            repository: 'github.com/golang/oauth2',
            chunkMatches: [
                {
                    content: 'package oauth2_test',
                    contentStart: {
                        line: 4,
                        offset: 4,
                        column: 0,
                    },
                    ranges: [
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                        {
                            start: {
                                line: 4,
                                offset: 15,
                                column: 15,
                            },
                            end: {
                                line: 4,
                                offset: 19,
                                column: 19,
                            },
                        },
                    ],
                },
            ],
            repoStars: 42,
            repoLastFetched: '2017-01-01T00:00:00Z',
        },
    ],
}

export const HIGHLIGHTED_FILE_LINES = [
    [
        '<tr><td class="line" data-line="1"></td><td class="code"><span style="color:#268bd2;">language</span><span style="color:#657b83;">: </span><span style="color:#2aa198;">go↵</span></td></tr>',
        '<tr><td class="line" data-line="2"></td><td class="code"><span style="color:#657b83;">↵</span></td></tr>',
        '<tr><td class="line" data-line="3"></td><td class="code"><span style="color:#268bd2;">go</span><span style="color:#657b83;">:↵</span></td></tr>',
        '<tr><td class="line" data-line="4"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">tip↵</span></td></tr>',
        '<tr><td class="line" data-line="5"></td><td class="code"><span style="color:#657b83;">↵</span></td></tr>',
        '<tr><td class="line" data-line="6"></td><td class="code"><span style="color:#268bd2;">install</span><span style="color:#657b83;">:↵</span></td></tr>',
        '<tr><td class="line" data-line="7"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">export GOPATH=&#34;$HOME/gopath&#34;↵</span></td></tr>',
        '<tr><td class="line" data-line="8"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">mkdir -p &#34;$GOPATH/src/golang.org/x&#34;↵</span></td></tr>',
        '<tr><td class="line" data-line="9"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">mv &#34;$TRAVIS_BUILD_DIR&#34; &#34;$GOPATH/src/golang.org/x/oauth2&#34;↵</span></td></tr>',
        '<tr><td class="line" data-line="10"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">go get -v -t -d golang.org/x/oauth2/...↵</span></td></tr><tr>',
        '<td class="line" data-line="11"></td><td class="code"><span style="color:#657b83;">↵</span></td></tr>',
        '<tr><td class="line" data-line="12"></td><td class="code"><span style="color:#268bd2;">script</span><span style="color:#657b83;">:↵</span></td></tr><tr>',
        '<td class="line" data-line="13"></td><td class="code"><span style="color:#657b83;">  - </span><span style="color:#2aa198;">go test -v golang.org/x/oauth2/...</span></td></tr>',
    ],
]

export const HIGHLIGHTED_FILE_LINES_SIMPLE = [
    [
        '<tr><td class="line" data-line="1"></td><td class="code">first line of code</td></tr>',
        '<tr><td class="line" data-line="2"></td><td class="code">second line of code</td></tr>',
        '<tr><td class="line" data-line="3"></td><td class="code">third line of code</td></tr>',
        '<tr><td class="line" data-line="4"></td><td class="code">fourth</td></tr>',
        '<tr><td class="line" data-line="5"></td><td class="code">fifth</td></tr>',
        '<tr><td class="line" data-line="6"></td><td class="code">sixth</td></tr>',
        '<tr><td class="line" data-line="7"></td><td class="code">seventh</td></tr>',
        '<tr><td class="line" data-line="8"></td><td class="code">eighth</td></tr>',
        '<tr><td class="line" data-line="9"></td><td class="code">ninth</td></tr>',
        '<tr><td class="line" data-line="10"></td><td class="code">tenth</td></tr>',
    ],
]

export const HIGHLIGHTED_FILE_LINES_LONG = [
    [
        '<tr><td class="line" data-line="1"></td><td class="code"><div><span style="color:#93a1a1;">// Copyright 2014 The Go Authors. All rights reserved.↵</span></div></td></tr>',
        '<tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#93a1a1;">// Use of this source code is governed by a BSD-style↵</span></div></td></tr>',
        '<tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#93a1a1;">// license that can be found in the LICENSE file.↵</span></div></td></tr>',
        '<tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#657b83;">↵</span></div></td></tr>',
        '<tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#859900;">package</span><span style="color:#657b83;"> oauth2_test↵</span></div></td></tr>',
        '<tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#657b83;">↵</span></div></td></tr>',
        '<tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">(↵</span></div></td></tr>',
        '<tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">context</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="9"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">fmt</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="10"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">log</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="11"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">net/http</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="12"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">time</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="13"></td><td class="code"><div><span style="color:#657b83;">↵</span></div></td></tr>',
        '<tr><td class="line" data-line="14"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">golang.org/x/oauth2</span><span style="color:#839496;">&#34;↵</span></div></td></tr>',
        '<tr><td class="line" data-line="15"></td><td class="code"><div><span style="color:#657b83;">)↵</span></div></td></tr>',
        '<tr><td class="line" data-line="16"></td><td class="code"><div><span style="color:#657b83;">↵</span></div></td></tr>',
        '<tr><td class="line" data-line="17"></td><td class="code"><div><span style="color:#268bd2;">func </span><span style="color:#b58900;">ExampleConfig</span><span style="color:#657b83;">() {↵</span></div></td></tr>',
        '<tr><td class="line" data-line="18"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#268bd2;">ctx </span><span style="color:#859900;">:=</span><span style="color:#657b83;"> context.</span><span style="color:#b58900;">Background</span><span style="color:#657b83;">()↵</span></div></td></tr>',
        '<tr><td class="line" data-line="19"></td><td class="code"><div><span style="color:#657b83;">	</span><span style="color:#268bd2;">conf </span><span style="color:#859900;">:= &amp;</span><span style="color:#657b83;">oauth2.</span><span style="color:#268bd2;">Config</span><span style="color:#657b83;">{↵</span></div></td></tr>',
        '<tr><td class="line" data-line="20"></td><td class="code"><div><span style="color:#657b83;">		</span><span style="color:#b58900;">ClientID</span><span style="color:#657b83;">:     </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">YOUR_CLIENT_ID</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">,↵</span></div></td></tr>',
        '<tr><td class="line" data-line="21"></td><td class="code"><div><span style="color:#657b83;">		</span><span style="color:#b58900;">ClientSecret</span><span style="color:#657b83;">: </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">YOUR_CLIENT_SECRET</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">,↵</span></div></td></tr>',
        '<tr><td class="line" data-line="22"></td><td class="code"><div><span style="color:#657b83;">		</span><span style="color:#b58900;">Scopes</span><span style="color:#657b83;">:       []</span><span style="color:#268bd2;">string</span><span style="color:#657b83;">{</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">SCOPE1</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">SCOPE2</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">},↵</span></div></td></tr>',
        '<tr><td class="line" data-line="23"></td><td class="code"><div><span style="color:#657b83;">		</span><span style="color:#b58900;">Endpoint</span><span style="color:#657b83;">: oauth2.</span><span style="color:#268bd2;">Endpoint</span><span style="color:#657b83;">{↵</span></div></td></tr>',
        '<tr><td class="line" data-line="24"></td><td class="code"><div><span style="color:#657b83;">			</span><span style="color:#b58900;">AuthURL</span><span style="color:#657b83;">:  </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">https://provider.com/o/oauth2/auth</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">,↵</span></div></td></tr>',
    ],
]

export const HIGHLIGHTED_FILE_LINES_REQUEST = sinon.fake((parameters: FetchFileParameters) =>
    of(parameters.ranges.map(range => HIGHLIGHTED_FILE_LINES[0].slice(range.startLine, range.endLine)))
)
export const HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST = sinon.fake((parameters: FetchFileParameters) =>
    of(parameters.ranges.map(range => HIGHLIGHTED_FILE_LINES_SIMPLE[0].slice(range.startLine, range.endLine)))
)
export const HIGHLIGHTED_FILE_LINES_LONG_REQUEST = sinon.fake((parameters: FetchFileParameters) =>
    of(parameters.ranges.map(range => HIGHLIGHTED_FILE_LINES_LONG[0].slice(range.startLine, range.endLine)))
)

export const NOOP_SETTINGS_CASCADE = {
    subjects: null,
    final: null,
}

export const extensionsController: Controller = {
    executeCommand: () => Promise.resolve(),
    registerCommand: () => new Subscription(),
    extHostAPI: Promise.resolve(
        pretendRemote<FlatExtensionHostAPI>({
            getContributions: () => pretendProxySubscribable(NEVER),
            registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
            haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
        })
    ),
    commandErrors: EMPTY,
    unsubscribe: noop,
}

export const NOOP_PLATFORM_CONTEXT: Pick<
    PlatformContext,
    'sourcegraphURL' | 'requestGraphQL' | 'urlToFile' | 'settings'
> = {
    requestGraphQL: () => EMPTY,
    urlToFile: () => '',
    settings: of(NOOP_SETTINGS_CASCADE),
    sourcegraphURL: '',
}
