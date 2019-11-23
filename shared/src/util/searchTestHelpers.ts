import { of } from 'rxjs'
import sinon from 'sinon'
import { Controller } from '../extensions/controller'

export const RESULT = {
    __typename: 'FileMatch',
    file: {
        path: '.travis.yml',
        url: '/github.com/golang/oauth2/-/blob/.travis.yml',
        commit: { oid: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3' },
    },
    repository: { name: 'github.com/golang/oauth2', url: '/github.com/golang/oauth2' },
    limitHit: false,
    symbols: [],
    lineMatches: [
        {
            preview: '  - go test -v golang.org/x/oauth2/...',
            lineNumber: 12,
            offsetAndLengths: [[7, 4]],
            limitHit: false,
        },
    ],
}

export const MULTIPLE_MATCH_RESULT = {
    __typename: 'FileMatch',
    file: {
        path: 'clientcredentials/clientcredentials_test.go',
        url: '/github.com/golang/oauth2/-/blob/clientcredentials/clientcredentials_test.go',
        commit: {
            oid: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
        },
    },
    repository: {
        name: 'github.com/golang/oauth2',
        url: '/github.com/golang/oauth2',
    },
    limitHit: false,
    symbols: [],
    lineMatches: [
        {
            preview: '\t"net/http/httptest"',
            lineNumber: 11,
            offsetAndLengths: [[15, 4]],
            limitHit: false,
        },
        {
            preview: '\t"testing"',
            lineNumber: 13,
            offsetAndLengths: [[2, 4]],
            limitHit: false,
        },
        {
            preview: 'func TestTokenSourceGrantTypeOverride(t *testing.T) {',
            lineNumber: 36,
            offsetAndLengths: [
                [5, 4],
                [41, 4],
            ],
            limitHit: false,
        },
        {
            preview: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            lineNumber: 39,
            offsetAndLengths: [[11, 4]],
            limitHit: false,
        },
        {
            preview: 'func TestTokenRequest(t *testing.T) {',
            lineNumber: 73,
            offsetAndLengths: [
                [5, 4],
                [25, 4],
            ],
            limitHit: false,
        },
        {
            preview: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            lineNumber: 74,
            offsetAndLengths: [[11, 4]],
            limitHit: false,
        },
        {
            preview: 'func TestTokenRefreshRequest(t *testing.T) {',
            lineNumber: 115,
            offsetAndLengths: [
                [5, 4],
                [32, 4],
            ],
            limitHit: false,
        },
        {
            preview: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
            lineNumber: 117,
            offsetAndLengths: [[11, 4]],
            limitHit: false,
        },
        {
            preview: '\t\tio.WriteString(w, `{"access_token": "foo", "refresh_token": "bar"}`)',
            lineNumber: 134,
            offsetAndLengths: [[8, 4]],
            limitHit: false,
        },
    ],
}

export const SEARCH_RESULT = {
    __typename: 'SearchResults',
    limitHit: false,
    resultCount: 1,
    approximateResultCount: '1',
    missing: [],
    cloning: [],
    timedout: [],
    indexUnavailable: false,
    dynamicFilters: [
        { value: 'file:\\.yml$', label: 'file:\\.yml$', count: 1, limitHit: false, kind: 'file' },
        { value: 'case:yes', label: 'case:yes', count: 0, limitHit: false, kind: 'case' },
        {
            value: 'repo:^github\\.com/golang/oauth2$',
            label: 'github.com/golang/oauth2',
            count: 1,
            limitHit: false,
            kind: 'repo',
        },
    ],
    results: [RESULT],
    alert: null,
    elapsedMilliseconds: 78,
}

export const MULTIPLE_SEARCH_RESULT = {
    ...SEARCH_RESULT,
    limitHit: false,
    resultCount: 136,
    approximateResultCount: '136',
    results: [
        RESULT,
        MULTIPLE_MATCH_RESULT,
        {
            __typename: 'FileMatch',
            file: {
                path: 'example_test.go',
                url: '/github.com/golang/oauth2/-/blob/example_test.go',
                commit: {
                    oid: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
                },
            },
            repository: {
                name: 'github.com/golang/oauth2',
                url: '/github.com/golang/oauth2',
            },
            limitHit: false,
            symbols: [],
            lineMatches: [
                {
                    preview: 'package oauth2_test',
                    lineNumber: 4,
                    offsetAndLengths: [[15, 4]],
                },
            ],
        },
    ],
}

// Result from query: repo:^github\.com/golang/oauth2$ test f:travis
export const SEARCH_REQUEST = sinon.fake.returns(SEARCH_RESULT)
export const MULTIPLE_SEARCH_REQUEST = sinon.fake.returns(MULTIPLE_SEARCH_RESULT)
export const OBSERVABLE_SEARCH_REQUEST = sinon.fake.returns(of(SEARCH_RESULT))

export const HIGHLIGHTED_FILE_LINES = [
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
]

export const HIGHLIGHTED_FILE_LINES_SIMPLE = [
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
]

export const HIGHLIGHTED_FILE_LINES_REQUEST = sinon.fake.returns(of(HIGHLIGHTED_FILE_LINES))
export const HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST = sinon.fake.returns(of(HIGHLIGHTED_FILE_LINES_SIMPLE))

export const NOOP_SETTINGS_CASCADE = {
    subjects: null,
    final: null,
}

const services = {
    contribution: {
        getContributions: () => of({}),
    },
}

export const extensionsController: Pick<Controller, 'executeCommand' | 'services'> = {
    executeCommand: () => Promise.resolve(),
    services: services as any,
}
