import { describe, expect, test } from '@jest/globals'

import type { SearchMatch, SearchType, SymbolMatch } from '@sourcegraph/shared/src/search/stream'

import { searchResultsToFileContent, buildFileName } from './searchResultsExport'

describe('searchResultsToFileContent', () => {
    const sourcegraphURL = 'http://localhost:3443'
    const data: [SearchType | 'owner' | null, SearchMatch[], string][] = [
        [
            null,
            [
                {
                    type: 'content',
                    path: 'jssource/src_files/include/javascript/sugar_3.js',
                    repository: 'github.com/salesagility/SuiteCRM',
                    repoStars: 3254,
                    repoLastFetched: '2022-12-15T13:46:29.705328Z',
                    branches: [''],
                    commit: 'cea3e00eb4990d143c91935013058488c0964ee6',
                    chunkMatches: [
                        {
                            content: '    safari: {min: 534}, // Safari 5.1',
                            contentStart: {
                                offset: 5649,
                                line: 191,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 5653,
                                        line: 191,
                                        column: 4,
                                    },
                                    end: {
                                        offset: 5659,
                                        line: 191,
                                        column: 10,
                                    },
                                },
                                {
                                    start: {
                                        offset: 5676,
                                        line: 191,
                                        column: 27,
                                    },
                                    end: {
                                        offset: 5682,
                                        line: 191,
                                        column: 33,
                                    },
                                },
                            ],
                        },
                        {
                            content: "var isSafari = (navigator.userAgent.toLowerCase().indexOf('safari') != -1);",
                            contentStart: {
                                offset: 6969,
                                line: 233,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6975,
                                        line: 233,
                                        column: 6,
                                    },
                                    end: {
                                        offset: 6981,
                                        line: 233,
                                        column: 12,
                                    },
                                },
                                {
                                    start: {
                                        offset: 7028,
                                        line: 233,
                                        column: 59,
                                    },
                                    end: {
                                        offset: 7034,
                                        line: 233,
                                        column: 65,
                                    },
                                },
                            ],
                        },
                        {
                            content: '    else if ($.browser.safari) { // Safari',
                            contentStart: {
                                offset: 6309,
                                line: 209,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6332,
                                        line: 209,
                                        column: 23,
                                    },
                                    end: {
                                        offset: 6338,
                                        line: 209,
                                        column: 29,
                                    },
                                },
                                {
                                    start: {
                                        offset: 6345,
                                        line: 209,
                                        column: 36,
                                    },
                                    end: {
                                        offset: 6351,
                                        line: 209,
                                        column: 42,
                                    },
                                },
                            ],
                        },
                        {
                            content: "      supported = supportedBrowsers['safari'];",
                            contentStart: {
                                offset: 6352,
                                line: 210,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6389,
                                        line: 210,
                                        column: 37,
                                    },
                                    end: {
                                        offset: 6395,
                                        line: 210,
                                        column: 43,
                                    },
                                },
                            ],
                        },
                        {
                            content: '            //Chrome or safari on a mac',
                            contentStart: {
                                offset: 122918,
                                line: 3496,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 122942,
                                        line: 3496,
                                        column: 24,
                                    },
                                    end: {
                                        offset: 122948,
                                        line: 3496,
                                        column: 30,
                                    },
                                },
                            ],
                        },
                        {
                            content: '            //this is the default for safari, IE and Chrome',
                            contentStart: {
                                offset: 123607,
                                line: 3515,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 123645,
                                        line: 3515,
                                        column: 38,
                                    },
                                    end: {
                                        offset: 123651,
                                        line: 3515,
                                        column: 44,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                '            //if this is webkit and is NOT google, then we are most likely looking at Safari',
                            contentStart: {
                                offset: 123667,
                                line: 3516,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 123753,
                                        line: 3516,
                                        column: 86,
                                    },
                                    end: {
                                        offset: 123759,
                                        line: 3516,
                                        column: 92,
                                    },
                                },
                            ],
                        },
                    ],
                },
                {
                    type: 'content',
                    path: 'include/javascript/sugar_3.js',
                    repository: 'github.com/salesagility/SuiteCRM',
                    repoStars: 3254,
                    repoLastFetched: '2022-12-15T13:46:29.705328Z',
                    branches: [''],
                    commit: 'cea3e00eb4990d143c91935013058488c0964ee6',
                    chunkMatches: [
                        {
                            content:
                                'SUGAR.isSupportedBrowser=function(){var supportedBrowsers={msie:{min:9,max:11},safari:{min:534},mozilla:{min:31.0},chrome:{min:37}};var current=String($.browser.version);var supported;if($.browser.msie||(!(window.ActiveXObject)&&"ActiveXObject"in window)){supported=supportedBrowsers[\'msie\'];}',
                            contentStart: {
                                offset: 4147,
                                line: 44,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 4226,
                                        line: 44,
                                        column: 79,
                                    },
                                    end: {
                                        offset: 4232,
                                        line: 44,
                                        column: 85,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                "SUGAR.isIE=isSupportedIE();var isSafari=(navigator.userAgent.toLowerCase().indexOf('safari')!=-1);RegExp.escape=function(text){if(!arguments.callee.sRE){var specials=['/','.','*','+','?','|','(',')','[',']','{','}','\\\\'];arguments.callee.sRE=new RegExp('(\\\\'+specials.join('|\\\\')+')','g');}",
                            contentStart: {
                                offset: 5153,
                                line: 53,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 5186,
                                        line: 53,
                                        column: 33,
                                    },
                                    end: {
                                        offset: 5192,
                                        line: 53,
                                        column: 39,
                                    },
                                },
                                {
                                    start: {
                                        offset: 5237,
                                        line: 53,
                                        column: 84,
                                    },
                                    end: {
                                        offset: 5243,
                                        line: 53,
                                        column: 90,
                                    },
                                },
                            ],
                        },
                        {
                            content: "else if($.browser.safari){supported=supportedBrowsers['safari'];}}",
                            contentStart: {
                                offset: 4696,
                                line: 47,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 4714,
                                        line: 47,
                                        column: 18,
                                    },
                                    end: {
                                        offset: 4720,
                                        line: 47,
                                        column: 24,
                                    },
                                },
                                {
                                    start: {
                                        offset: 4751,
                                        line: 47,
                                        column: 55,
                                    },
                                    end: {
                                        offset: 4757,
                                        line: 47,
                                        column: 61,
                                    },
                                },
                            ],
                        },
                    ],
                },
                {
                    type: 'content',
                    path: 'src/scripts/safari/Safari.ts',
                    pathMatches: [
                        {
                            start: {
                                offset: 12,
                                line: 0,
                                column: 12,
                            },
                            end: {
                                offset: 18,
                                line: 0,
                                column: 18,
                            },
                        },
                        {
                            start: {
                                offset: 19,
                                line: 0,
                                column: 19,
                            },
                            end: {
                                offset: 25,
                                line: 0,
                                column: 25,
                            },
                        },
                    ],
                    repository: 'github.com/pokeclicker/pokeclicker',
                    repoStars: 386,
                    repoLastFetched: '2022-12-15T05:58:01.450798Z',
                    branches: [''],
                    commit: '8cd76ead3182b39647ae1c7af2dc9b912f7d6e9e',
                    chunkMatches: [
                        {
                            content: 'class Safari {',
                            contentStart: {
                                offset: 0,
                                line: 0,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6,
                                        line: 0,
                                        column: 6,
                                    },
                                    end: {
                                        offset: 12,
                                        line: 0,
                                        column: 12,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                '    static pokemonGrid: KnockoutObservableArray<SafariPokemon> = ko.observableArray([]);',
                            contentStart: {
                                offset: 54,
                                line: 2,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 102,
                                        line: 2,
                                        column: 48,
                                    },
                                    end: {
                                        offset: 108,
                                        line: 2,
                                        column: 54,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                "        return Math.floor(document.querySelector('#safariModal .modal-dialog').scrollWidth / 32);",
                            contentStart: {
                                offset: 697,
                                line: 17,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 748,
                                        line: 17,
                                        column: 51,
                                    },
                                    end: {
                                        offset: 754,
                                        line: 17,
                                        column: 57,
                                    },
                                },
                            ],
                        },
                        {
                            content: '        Safari.grid = [];',
                            contentStart: {
                                offset: 932,
                                line: 25,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 940,
                                        line: 25,
                                        column: 8,
                                    },
                                    end: {
                                        offset: 946,
                                        line: 25,
                                        column: 14,
                                    },
                                },
                            ],
                        },
                        {
                            content: '        Safari.pokemonGrid([]);',
                            contentStart: {
                                offset: 958,
                                line: 26,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 966,
                                        line: 26,
                                        column: 8,
                                    },
                                    end: {
                                        offset: 972,
                                        line: 26,
                                        column: 14,
                                    },
                                },
                            ],
                        },
                    ],
                },
            ],

            'Match type,Repository,Repository external URL,File path,File URL,Path matches [path [start end]],Chunk matches [line [start end]]\ncontent,github.com/salesagility/SuiteCRM,http://localhost:3443/github.com/salesagility/SuiteCRM,jssource/src_files/include/javascript/sugar_3.js,http://localhost:3443/github.com/salesagility/SuiteCRM/-/blob/jssource/src_files/include/javascript/sugar_3.js,,"[191, [[4, 10] [27, 33]]]; [233, [[6, 12] [59, 65]]]; [209, [[23, 29] [36, 42]]]; [210, [[37, 43]]]; [3496, [[24, 30]]]; [3515, [[38, 44]]]; [3516, [[86, 92]]]"\ncontent,github.com/salesagility/SuiteCRM,http://localhost:3443/github.com/salesagility/SuiteCRM,include/javascript/sugar_3.js,http://localhost:3443/github.com/salesagility/SuiteCRM/-/blob/include/javascript/sugar_3.js,,"[44, [[79, 85]]]; [53, [[33, 39] [84, 90]]]; [47, [[18, 24] [55, 61]]]"\ncontent,github.com/pokeclicker/pokeclicker,http://localhost:3443/github.com/pokeclicker/pokeclicker,src/scripts/safari/Safari.ts,http://localhost:3443/github.com/pokeclicker/pokeclicker/-/blob/src/scripts/safari/Safari.ts,"[src/scripts/safari/Safari.ts, [[12, 18] [19, 25]]]","[0, [[6, 12]]]; [2, [[48, 54]]]; [17, [[51, 57]]]; [25, [[8, 14]]]; [26, [[8, 14]]]"',
        ],
        [
            'file',
            [
                {
                    type: 'content',
                    path: 'tests/wpt/web-platform-tests/tools/wpt/run.py',
                    repository: 'github.com/servo/servo',
                    repoStars: 21921,
                    repoLastFetched: '2022-12-15T11:36:58.32248Z',
                    branches: [''],
                    commit: '2a45f247c44117e4a13603a87a3b6304905f8d1f',
                    chunkMatches: [
                        {
                            content: 'class Safari(BrowserSetup):',
                            contentStart: {
                                offset: 24181,
                                line: 623,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 24187,
                                        line: 623,
                                        column: 6,
                                    },
                                    end: {
                                        offset: 24193,
                                        line: 623,
                                        column: 12,
                                    },
                                },
                            ],
                        },
                        {
                            content: '    name = "safari"',
                            contentStart: {
                                offset: 24209,
                                line: 624,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 24221,
                                        line: 624,
                                        column: 12,
                                    },
                                    end: {
                                        offset: 24227,
                                        line: 624,
                                        column: 18,
                                    },
                                },
                            ],
                        },
                        {
                            content: '    browser_cls = browser.Safari',
                            contentStart: {
                                offset: 24229,
                                line: 625,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 24255,
                                        line: 625,
                                        column: 26,
                                    },
                                    end: {
                                        offset: 24261,
                                        line: 625,
                                        column: 32,
                                    },
                                },
                            ],
                        },
                        {
                            content: '    "safari": Safari,',
                            contentStart: {
                                offset: 28097,
                                line: 751,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 28102,
                                        line: 751,
                                        column: 5,
                                    },
                                    end: {
                                        offset: 28108,
                                        line: 751,
                                        column: 11,
                                    },
                                },
                                {
                                    start: {
                                        offset: 28111,
                                        line: 751,
                                        column: 14,
                                    },
                                    end: {
                                        offset: 28117,
                                        line: 751,
                                        column: 20,
                                    },
                                },
                            ],
                        },
                        {
                            content: '                raise WptrunError("Unable to locate safaridriver binary")',
                            contentStart: {
                                offset: 24554,
                                line: 635,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 24606,
                                        line: 635,
                                        column: 52,
                                    },
                                    end: {
                                        offset: 24612,
                                        line: 635,
                                        column: 58,
                                    },
                                },
                            ],
                        },
                    ],
                },
                {
                    type: 'content',
                    path: 'jssource/src_files/include/javascript/sugar_3.js',
                    repository: 'github.com/salesagility/SuiteCRM',
                    repoStars: 3254,
                    repoLastFetched: '2022-12-15T14:08:32.423002Z',
                    branches: [''],
                    commit: 'cea3e00eb4990d143c91935013058488c0964ee6',
                    chunkMatches: [
                        {
                            content: '    safari: {min: 534}, // Safari 5.1',
                            contentStart: {
                                offset: 5649,
                                line: 191,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 5653,
                                        line: 191,
                                        column: 4,
                                    },
                                    end: {
                                        offset: 5659,
                                        line: 191,
                                        column: 10,
                                    },
                                },
                                {
                                    start: {
                                        offset: 5676,
                                        line: 191,
                                        column: 27,
                                    },
                                    end: {
                                        offset: 5682,
                                        line: 191,
                                        column: 33,
                                    },
                                },
                            ],
                        },
                        {
                            content: "var isSafari = (navigator.userAgent.toLowerCase().indexOf('safari') != -1);",
                            contentStart: {
                                offset: 6969,
                                line: 233,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6975,
                                        line: 233,
                                        column: 6,
                                    },
                                    end: {
                                        offset: 6981,
                                        line: 233,
                                        column: 12,
                                    },
                                },
                                {
                                    start: {
                                        offset: 7028,
                                        line: 233,
                                        column: 59,
                                    },
                                    end: {
                                        offset: 7034,
                                        line: 233,
                                        column: 65,
                                    },
                                },
                            ],
                        },
                        {
                            content: '    else if ($.browser.safari) { // Safari',
                            contentStart: {
                                offset: 6309,
                                line: 209,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6332,
                                        line: 209,
                                        column: 23,
                                    },
                                    end: {
                                        offset: 6338,
                                        line: 209,
                                        column: 29,
                                    },
                                },
                                {
                                    start: {
                                        offset: 6345,
                                        line: 209,
                                        column: 36,
                                    },
                                    end: {
                                        offset: 6351,
                                        line: 209,
                                        column: 42,
                                    },
                                },
                            ],
                        },
                        {
                            content: "      supported = supportedBrowsers['safari'];",
                            contentStart: {
                                offset: 6352,
                                line: 210,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 6389,
                                        line: 210,
                                        column: 37,
                                    },
                                    end: {
                                        offset: 6395,
                                        line: 210,
                                        column: 43,
                                    },
                                },
                            ],
                        },
                        {
                            content: '            //Chrome or safari on a mac',
                            contentStart: {
                                offset: 122918,
                                line: 3496,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 122942,
                                        line: 3496,
                                        column: 24,
                                    },
                                    end: {
                                        offset: 122948,
                                        line: 3496,
                                        column: 30,
                                    },
                                },
                            ],
                        },
                        {
                            content: '            //this is the default for safari, IE and Chrome',
                            contentStart: {
                                offset: 123607,
                                line: 3515,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 123645,
                                        line: 3515,
                                        column: 38,
                                    },
                                    end: {
                                        offset: 123651,
                                        line: 3515,
                                        column: 44,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                '            //if this is webkit and is NOT google, then we are most likely looking at Safari',
                            contentStart: {
                                offset: 123667,
                                line: 3516,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 123753,
                                        line: 3516,
                                        column: 86,
                                    },
                                    end: {
                                        offset: 123759,
                                        line: 3516,
                                        column: 92,
                                    },
                                },
                            ],
                        },
                    ],
                },
                {
                    type: 'content',
                    path: 'include/javascript/sugar_3.js',
                    repository: 'github.com/salesagility/SuiteCRM',
                    repoStars: 3254,
                    repoLastFetched: '2022-12-15T14:08:32.423002Z',
                    branches: [''],
                    commit: 'cea3e00eb4990d143c91935013058488c0964ee6',
                    chunkMatches: [
                        {
                            content:
                                'SUGAR.isSupportedBrowser=function(){var supportedBrowsers={msie:{min:9,max:11},safari:{min:534},mozilla:{min:31.0},chrome:{min:37}};var current=String($.browser.version);var supported;if($.browser.msie||(!(window.ActiveXObject)&&"ActiveXObject"in window)){supported=supportedBrowsers[\'msie\'];}',
                            contentStart: {
                                offset: 4147,
                                line: 44,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 4226,
                                        line: 44,
                                        column: 79,
                                    },
                                    end: {
                                        offset: 4232,
                                        line: 44,
                                        column: 85,
                                    },
                                },
                            ],
                        },
                        {
                            content:
                                "SUGAR.isIE=isSupportedIE();var isSafari=(navigator.userAgent.toLowerCase().indexOf('safari')!=-1);RegExp.escape=function(text){if(!arguments.callee.sRE){var specials=['/','.','*','+','?','|','(',')','[',']','{','}','\\\\'];arguments.callee.sRE=new RegExp('(\\\\'+specials.join('|\\\\')+')','g');}",
                            contentStart: {
                                offset: 5153,
                                line: 53,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 5186,
                                        line: 53,
                                        column: 33,
                                    },
                                    end: {
                                        offset: 5192,
                                        line: 53,
                                        column: 39,
                                    },
                                },
                                {
                                    start: {
                                        offset: 5237,
                                        line: 53,
                                        column: 84,
                                    },
                                    end: {
                                        offset: 5243,
                                        line: 53,
                                        column: 90,
                                    },
                                },
                            ],
                        },
                        {
                            content: "else if($.browser.safari){supported=supportedBrowsers['safari'];}}",
                            contentStart: {
                                offset: 4696,
                                line: 47,
                                column: 0,
                            },
                            ranges: [
                                {
                                    start: {
                                        offset: 4714,
                                        line: 47,
                                        column: 18,
                                    },
                                    end: {
                                        offset: 4720,
                                        line: 47,
                                        column: 24,
                                    },
                                },
                            ],
                        },
                    ],
                },
            ],

            'Match type,Repository,Repository external URL,File path,File URL,Path matches [path [start end]],Chunk matches [line [start end]]\ncontent,github.com/servo/servo,http://localhost:3443/github.com/servo/servo,tests/wpt/web-platform-tests/tools/wpt/run.py,http://localhost:3443/github.com/servo/servo/-/blob/tests/wpt/web-platform-tests/tools/wpt/run.py,,"[623, [[6, 12]]]; [624, [[12, 18]]]; [625, [[26, 32]]]; [751, [[5, 11] [14, 20]]]; [635, [[52, 58]]]"\ncontent,github.com/salesagility/SuiteCRM,http://localhost:3443/github.com/salesagility/SuiteCRM,jssource/src_files/include/javascript/sugar_3.js,http://localhost:3443/github.com/salesagility/SuiteCRM/-/blob/jssource/src_files/include/javascript/sugar_3.js,,"[191, [[4, 10] [27, 33]]]; [233, [[6, 12] [59, 65]]]; [209, [[23, 29] [36, 42]]]; [210, [[37, 43]]]; [3496, [[24, 30]]]; [3515, [[38, 44]]]; [3516, [[86, 92]]]"\ncontent,github.com/salesagility/SuiteCRM,http://localhost:3443/github.com/salesagility/SuiteCRM,include/javascript/sugar_3.js,http://localhost:3443/github.com/salesagility/SuiteCRM/-/blob/include/javascript/sugar_3.js,,"[44, [[79, 85]]]; [53, [[33, 39] [84, 90]]]; [47, [[18, 24]]]"',
        ],
        [
            'repo',
            [
                {
                    type: 'repo',
                    repository: 'github.com/lorenzodifuccia/safaribooks',
                    repositoryMatches: [
                        {
                            start: {
                                offset: 27,
                                line: 0,
                                column: 27,
                            },
                            end: {
                                offset: 33,
                                line: 0,
                                column: 33,
                            },
                        },
                    ],
                    repoStars: 3978,
                    repoLastFetched: '2022-12-15T03:33:35.440014Z',
                    description:
                        "Download and generate EPUB of your favorite books from O'Reilly Learning (aka Safari Books Online) library.",
                    metadata: {
                        oss: undefined,
                        deprecated: undefined,
                        'some",non-standard-key': 'value',
                    },
                },
                {
                    type: 'repo',
                    repository: 'github.com/rfletcher/safari-json-formatter',
                    repositoryMatches: [
                        {
                            start: {
                                offset: 21,
                                line: 0,
                                column: 21,
                            },
                            end: {
                                offset: 27,
                                line: 0,
                                column: 27,
                            },
                        },
                    ],
                    repoStars: 846,
                    repoLastFetched: '2022-12-15T07:26:46.526438Z',
                    description: 'A Safari extension which makes valid JSON documents human-readable.',
                },
                {
                    type: 'repo',
                    repository: 'github.com/AdguardTeam/AdGuardForSafari',
                    repositoryMatches: [
                        {
                            start: {
                                offset: 33,
                                line: 0,
                                column: 33,
                            },
                            end: {
                                offset: 39,
                                line: 0,
                                column: 39,
                            },
                        },
                    ],
                    repoStars: 806,
                    repoLastFetched: '2022-12-10T22:17:25.11809Z',
                    description: 'AdGuard for Safari app extension',
                },
                {
                    type: 'repo',
                    repository: 'github.com/kishikawakatsumi/SourceKitForSafari',
                    repositoryMatches: [
                        {
                            start: {
                                offset: 40,
                                line: 0,
                                column: 40,
                            },
                            end: {
                                offset: 46,
                                line: 0,
                                column: 46,
                            },
                        },
                    ],
                    repoStars: 673,
                    repoLastFetched: '2022-12-11T00:25:24.608934Z',
                    description:
                        'SourceKit for Safari is a Safari extension for GitHub, that enables Xcode features like go to definition, find references, or documentation on hover.',
                },
                {
                    type: 'repo',
                    repository: 'github.com/shaojiankui/iOS-UDID-Safari',
                    repositoryMatches: [
                        {
                            start: {
                                offset: 32,
                                line: 0,
                                column: 32,
                            },
                            end: {
                                offset: 38,
                                line: 0,
                                column: 38,
                            },
                        },
                    ],
                    repoStars: 635,
                    repoLastFetched: '2022-12-15T01:18:57.148369Z',
                    description:
                        'iOS-UDID-Safari,（不能上架Appstore！！！）通过Safari获取iOS设备真实UDID，use safari and mobileconfig get  ios device real udid',
                },
            ],

            'Match type,Repository,Repository external URL\nrepo,github.com/lorenzodifuccia/safaribooks,http://localhost:3443/github.com/lorenzodifuccia/safaribooks\nrepo,github.com/rfletcher/safari-json-formatter,http://localhost:3443/github.com/rfletcher/safari-json-formatter\nrepo,github.com/AdguardTeam/AdGuardForSafari,http://localhost:3443/github.com/AdguardTeam/AdGuardForSafari\nrepo,github.com/kishikawakatsumi/SourceKitForSafari,http://localhost:3443/github.com/kishikawakatsumi/SourceKitForSafari\nrepo,github.com/shaojiankui/iOS-UDID-Safari,http://localhost:3443/github.com/shaojiankui/iOS-UDID-Safari',
        ],
        [
            'path',
            [
                {
                    type: 'path',
                    path: 'bin/safari',
                    pathMatches: [
                        {
                            start: {
                                offset: 4,
                                line: 0,
                                column: 4,
                            },
                            end: {
                                offset: 10,
                                line: 0,
                                column: 10,
                            },
                        },
                    ],
                    repository: 'github.com/zzzeyez/dots',
                    repoStars: 124,
                    repoLastFetched: '2022-12-15T03:50:02.660709Z',
                    branches: [''],
                    commit: 'b69d3d3e6fb0473d6d25d28bd3446772904ff579',
                },
                {
                    type: 'path',
                    path: '.config/bin/safari',
                    pathMatches: [
                        {
                            start: {
                                offset: 12,
                                line: 0,
                                column: 12,
                            },
                            end: {
                                offset: 18,
                                line: 0,
                                column: 18,
                            },
                        },
                    ],
                    repository: 'github.com/rockyzhang24/dotfiles',
                    repoStars: 88,
                    repoLastFetched: '2022-12-10T17:52:55.463709Z',
                    branches: [''],
                    commit: '58517770e9fcdd08ba6a2971aaf33a2fcec1b49c',
                },
                {
                    type: 'path',
                    path: 'Beats/Safari',
                    pathMatches: [
                        {
                            start: {
                                offset: 6,
                                line: 0,
                                column: 6,
                            },
                            end: {
                                offset: 12,
                                line: 0,
                                column: 12,
                            },
                        },
                    ],
                    repository: 'github.com/DavidsFiddle/Sonic-Pi-Code-Bits',
                    repoStars: 120,
                    repoLastFetched: '2022-12-15T00:06:38.47244Z',
                    branches: [''],
                    commit: '6a1adbe8a5a556042dd20b7c064635b0b5475d0e',
                },
                {
                    type: 'path',
                    path: 'bin/safari',
                    pathMatches: [
                        {
                            start: {
                                offset: 4,
                                line: 0,
                                column: 4,
                            },
                            end: {
                                offset: 10,
                                line: 0,
                                column: 10,
                            },
                        },
                    ],
                    repository: 'github.com/unixorn/tumult.plugin.zsh',
                    repoStars: 144,
                    repoLastFetched: '2022-12-15T12:33:18.180322Z',
                    branches: [''],
                    commit: 'cd1ac844ba533c463d105330cc2ffaf4421ab42e',
                },
                {
                    type: 'path',
                    path: 'bin/osx/safari',
                    pathMatches: [
                        {
                            start: {
                                offset: 8,
                                line: 0,
                                column: 8,
                            },
                            end: {
                                offset: 14,
                                line: 0,
                                column: 14,
                            },
                        },
                    ],
                    repository: 'github.com/jaymcgavren/dotfiles',
                    repoStars: 34,
                    repoLastFetched: '2022-12-15T07:07:48.594905Z',
                    branches: [''],
                    commit: '67beb23333f82d3d35a1698a13f7ee574a7b2c43',
                },
            ],

            'Match type,Repository,Repository external URL,File path,File URL,Path matches [path [start end]],Chunk matches [line [start end]]\npath,github.com/zzzeyez/dots,http://localhost:3443/github.com/zzzeyez/dots,bin/safari,http://localhost:3443/github.com/zzzeyez/dots/-/blob/bin/safari,"[bin/safari, [[4, 10]]]",\npath,github.com/rockyzhang24/dotfiles,http://localhost:3443/github.com/rockyzhang24/dotfiles,.config/bin/safari,http://localhost:3443/github.com/rockyzhang24/dotfiles/-/blob/.config/bin/safari,"[.config/bin/safari, [[12, 18]]]",\npath,github.com/DavidsFiddle/Sonic-Pi-Code-Bits,http://localhost:3443/github.com/DavidsFiddle/Sonic-Pi-Code-Bits,Beats/Safari,http://localhost:3443/github.com/DavidsFiddle/Sonic-Pi-Code-Bits/-/blob/Beats/Safari,"[Beats/Safari, [[6, 12]]]",\npath,github.com/unixorn/tumult.plugin.zsh,http://localhost:3443/github.com/unixorn/tumult.plugin.zsh,bin/safari,http://localhost:3443/github.com/unixorn/tumult.plugin.zsh/-/blob/bin/safari,"[bin/safari, [[4, 10]]]",\npath,github.com/jaymcgavren/dotfiles,http://localhost:3443/github.com/jaymcgavren/dotfiles,bin/osx/safari,http://localhost:3443/github.com/jaymcgavren/dotfiles/-/blob/bin/osx/safari,"[bin/osx/safari, [[8, 14]]]",',
        ],
        [
            'symbol',
            [
                {
                    type: 'symbol',
                    path: 'enterprise/internal/batches/types/code_host.go',
                    repository: 'github.com/sourcegraph/sourcegraph',
                    repoStars: 7163,
                    repoLastFetched: '2022-12-15T14:22:13.114583Z',
                    branches: [''],
                    commit: '6c09a2246d9c9ddf6efa2d0cdf095ea8b86b0c9a',
                    symbols: [
                        {
                            url: '/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/batches/types/code_host.go?L6:6-6:14',
                            name: 'CodeHost',
                            containerName: 'types',
                            kind: 'STRUCT',
                            line: 6,
                        },
                    ],
                },
                {
                    type: 'symbol',
                    path: 'internal/database/external_services.go',
                    repository: 'github.com/sourcegraph/sourcegraph',
                    repoStars: 7163,
                    repoLastFetched: '2022-12-15T14:22:13.114583Z',
                    branches: [''],
                    commit: '6c09a2246d9c9ddf6efa2d0cdf095ea8b86b0c9a',
                    symbols: [
                        {
                            url: '/github.com/sourcegraph/sourcegraph/-/blob/internal/database/external_services.go?L225:2-225:10',
                            name: 'CodeHost',
                            containerName: 'database.ExternalServiceKind',
                            kind: 'FIELD',
                            line: 225,
                        },
                    ],
                },
                {
                    type: 'symbol',
                    path: 'pkg/cli/upgradeassistant/internal/repository/models/codehost.go',
                    repositoryID: 49049446,
                    repository: 'github.com/koderover/zadig',
                    repoStars: 1835,
                    repoLastFetched: '2022-12-14T15:31:10.839428Z',
                    branches: [''],
                    commit: '039158ab468a062c9f36dcb1d1b9d259083fba0a',
                    symbols: [
                        {
                            url: '/github.com/koderover/zadig/-/blob/pkg/cli/upgradeassistant/internal/repository/models/codehost.go?L19:6-19:14',
                            name: 'CodeHost',
                            containerName: 'models',
                            kind: 'STRUCT',
                            line: 19,
                        },
                    ],
                },
                {
                    type: 'symbol',
                    path: 'internal/extsvc/codehost.go',
                    repository: 'github.com/sourcegraph/sourcegraph',
                    repoStars: 7163,
                    repoLastFetched: '2022-12-15T14:22:13.114583Z',
                    branches: [''],
                    commit: '6c09a2246d9c9ddf6efa2d0cdf095ea8b86b0c9a',
                    symbols: [
                        {
                            url: '/github.com/sourcegraph/sourcegraph/-/blob/internal/extsvc/codehost.go?L10:6-10:14',
                            name: 'CodeHost',
                            containerName: 'extsvc',
                            kind: 'STRUCT',
                            line: 10,
                        },
                        {
                            url: '/github.com/sourcegraph/sourcegraph/-/blob/internal/extsvc/codehost.go?L64:9-64:20',
                            name: 'NewCodeHost',
                            containerName: 'extsvc',
                            kind: 'FUNCTION',
                            line: 64,
                        },
                    ],
                },
            ] as SymbolMatch[],

            'Match type,Repository,Repository external URL,File path,File URL,Symbols [kind name url]\nsymbol,github.com/sourcegraph/sourcegraph,http://localhost:3443/github.com/sourcegraph/sourcegraph,enterprise/internal/batches/types/code_host.go,http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/batches/types/code_host.go,"[STRUCT, CodeHost, http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/enterprise/internal/batches/types/code_host.go?L6:6-6:14]"\nsymbol,github.com/sourcegraph/sourcegraph,http://localhost:3443/github.com/sourcegraph/sourcegraph,internal/database/external_services.go,http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/internal/database/external_services.go,"[FIELD, CodeHost, http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/internal/database/external_services.go?L225:2-225:10]"\nsymbol,github.com/koderover/zadig,http://localhost:3443/github.com/koderover/zadig,pkg/cli/upgradeassistant/internal/repository/models/codehost.go,http://localhost:3443/github.com/koderover/zadig/-/blob/pkg/cli/upgradeassistant/internal/repository/models/codehost.go,"[STRUCT, CodeHost, http://localhost:3443/github.com/koderover/zadig/-/blob/pkg/cli/upgradeassistant/internal/repository/models/codehost.go?L19:6-19:14]"\nsymbol,github.com/sourcegraph/sourcegraph,http://localhost:3443/github.com/sourcegraph/sourcegraph,internal/extsvc/codehost.go,http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/internal/extsvc/codehost.go,"[STRUCT, CodeHost, http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/internal/extsvc/codehost.go?L10:6-10:14]; [FUNCTION, NewCodeHost, http://localhost:3443/github.com/sourcegraph/sourcegraph/-/blob/internal/extsvc/codehost.go?L64:9-64:20]"',
        ],
        [
            'diff',
            [
                {
                    type: 'commit',
                    url: '/github.com/EbookFoundation/free-programming-books/-/commit/76f7742b20f4875d3ca21774eac3d04322112821',
                    repository: 'github.com/EbookFoundation/free-programming-books',
                    oid: '76f7742b20f4875d3ca21774eac3d04322112821',
                    message:
                        'Create Norwegian How to (#8632)\n\n* Create HOWTO-no.md\r\n\r\n* Update HOWTO-no.md\r\n\r\n* fix? syntax\r\n\r\n* add -no link\r\n\r\nCo-authored-by: Eric Hellman <eric@hellman.net>',
                    authorName: 'Oscar Skille',
                    authorDate: '2022-11-15T23:25:02Z',
                    committerName: 'GitHub',
                    committerDate: '2022-11-15T23:25:02Z',
                    repoStars: 258422,
                    repoLastFetched: '2022-12-15T13:48:14.891675Z',
                    content:
                        '```diff\n/dev/null docs/HOWTO-no.md\n@@ -0,0 +14,3 @@ \n+* [Opprette en pull-forespørsel](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request)\n+* [GitHub Hei Verden](https://docs.github.com/en/get-started/quickstart/hello-world)\n+* [YouTube - Slik deler du en GitHub-rep og sender inn en pull-forespørsel](https://www.youtube.com/watch?v=0fKg7e37bQE)\n\n```',
                    ranges: [[4, 73, 5]],
                },
                {
                    type: 'commit',
                    url: '/github.com/freeCodeCamp/freeCodeCamp/-/commit/ff95595fa50d9085166dc07035d84364125099e0',
                    repository: 'github.com/freeCodeCamp/freeCodeCamp',
                    oid: 'ff95595fa50d9085166dc07035d84364125099e0',
                    message: 'feat(Cypress): give basic structure to cypress typescript files (#48734)',
                    authorName: 'Sem Bauke',
                    authorDate: '2022-12-12T17:07:03Z',
                    committerName: 'GitHub',
                    committerDate: '2022-12-12T17:07:03Z',
                    repoStars: 358237,
                    repoLastFetched: '2022-12-15T14:10:12.253959Z',
                    content:
                        "```diff\ncypress/e2e/default/learn/challenges/code-storage.js /dev/null\n@@ -6,3 +0,0 @@ \n-const location =\n-  '/learn/responsive-web-design/basic-html-and-html5/say-hello-to-html-elements';\n-\n@@ -14,5 +0,0 @@ \n-  it('renders seed code without localStorage', () => {\n-    const editorContents = `<h1>Hello</h1>`;\n-    cy.get(selectors.editor).as('editor').contains(editorContents);\n-    cy.get('@editor').click().focused().type(`{movetoend}<h1>Hello World</h1>`);\n-    cy.reload();\n@@ -22,3 +0,0 @@ \n-  it('renders code from localStorage after \"Ctrl + S\"', () => {\n-    const editorContents = `<h1>Hello</h1>`;\n-    cy.get(selectors.editor).as('editor').contains(editorContents);\n/dev/null cypress/e2e/default/learn/challenges/code-storage.ts\n@@ -0,0 +3,3 @@ \n+const location =\n+  '/learn/responsive-web-design/basic-html-and-html5/say-hello-to-html-elements';\n+\n@@ -0,0 +11,3 @@ \n+  it('renders seed code without localStorage', () => {\n+    const editorContents = `<h1>Hello</h1>`;\n+    cy.get(selectors.class.reactMonacoEditor)\n@@ -0,0 +15,3 @@ \n+      .contains(editorContents);\n+    cy.get('@editor').click().focused().type(`{movetoend}<h1>Hello World</h1>`);\n+    cy.reload();\n\n```",
                    ranges: [
                        [4, 58, 5],
                        [8, 33, 5],
                        [10, 62, 5],
                        [14, 33, 5],
                    ],
                },
            ],

            'Match type,Repository,Repository external URL,Date,Author,Subject,oid,Commit URL\ncommit,github.com/EbookFoundation/free-programming-books,http://localhost:3443/github.com/EbookFoundation/free-programming-books,2022-11-15T23:25:02Z,Oscar Skille,"Create Norwegian How to (#8632)* Create HOWTO-no.md\r\r* Update HOWTO-no.md\r\r* fix? syntax\r\r* add -no link\r\rCo-authored-by: Eric Hellman <eric@hellman.net>",76f7742b20f4875d3ca21774eac3d04322112821,http://localhost:3443/github.com/EbookFoundation/free-programming-books/-/commit/76f7742b20f4875d3ca21774eac3d04322112821\ncommit,github.com/freeCodeCamp/freeCodeCamp,http://localhost:3443/github.com/freeCodeCamp/freeCodeCamp,2022-12-12T17:07:03Z,Sem Bauke,"feat(Cypress): give basic structure to cypress typescript files (#48734)",ff95595fa50d9085166dc07035d84364125099e0,http://localhost:3443/github.com/freeCodeCamp/freeCodeCamp/-/commit/ff95595fa50d9085166dc07035d84364125099e0',
        ],
        [
            'commit',
            [
                {
                    type: 'commit',
                    url: '/github.com/EbookFoundation/free-programming-books/-/commit/91cdcd907bbf070b1310f9120b06bb03f481edb6',
                    repository: 'github.com/EbookFoundation/free-programming-books',
                    oid: '91cdcd907bbf070b1310f9120b06bb03f481edb6',
                    message:
                        "Added two courses R PT_BR (#7415)\n\n* Added two courses R PT_BR\r\n\r\nHello! I'm trying to help by adding courses that helped me!\r\n\r\n* fix: URL fragments goes in lowercase\r\n\r\n* lint: leave 2 blank lines between listings and next heading\r\n\r\n* Update courses/free-courses-pt_BR.md\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\n\r\n* Update courses/free-courses-pt_BR.md\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\nCo-authored-by: Eric Hellman <eric@hellman.net>",
                    authorName: 'Karen Ketlyn Ferreira Barcelos',
                    authorDate: '2022-10-30T03:07:11Z',
                    committerName: 'GitHub',
                    committerDate: '2022-10-30T03:07:11Z',
                    repoStars: 258422,
                    repoLastFetched: '2022-12-15T13:48:14.891675Z',
                    content:
                        "```COMMIT_EDITMSG\nAdded two courses R PT_BR (#7415)\n\n* Added two courses R PT_BR\r\n\r\nHello! I'm trying to help by adding courses that helped me!\r\n\r\n* fix: URL fragments goes in lowercase\r\n\r\n* lint: leave 2 blank lines between listings and next heading\r\n\r\n* Update courses/free-courses-pt_BR.md\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\n\r\n* Update courses/free-courses-pt_BR.md\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\n\r\nCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\nCo-authored-by: Eric Hellman <eric@hellman.net>\n```",
                    ranges: [[5, 0, 5]],
                },
                {
                    type: 'commit',
                    url: '/github.com/freeCodeCamp/freeCodeCamp/-/commit/3ecc381f95383dfa8314846a5a7d815b359d95d1',
                    repository: 'github.com/freeCodeCamp/freeCodeCamp',
                    oid: '3ecc381f95383dfa8314846a5a7d815b359d95d1',
                    message:
                        'fix(curriculum): corrected grammar in line 12 of the description section of learn HTML forms.md (#46859)\n\n* Update 60fad1cafcde010995e15306.md\r\n\r\nHello, \r\nI have changed the sentence, please have a look.\r\n\r\n* Update curriculum/challenges/english/14-responsive-web-design-22/learn-html-forms-by-building-a-registration-form/60fad1cafcde010995e15306.md\r\n\r\nThank you, I have committed the changes suggested by you.\r\n\r\nCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>\r\n\r\nCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>',
                    authorName: 'Harsh Kashyap',
                    authorDate: '2022-07-12T07:13:40Z',
                    committerName: 'GitHub',
                    committerDate: '2022-07-12T07:13:40Z',
                    repoStars: 358237,
                    repoLastFetched: '2022-12-15T14:10:12.253959Z',
                    content:
                        '```COMMIT_EDITMSG\nfix(curriculum): corrected grammar in line 12 of the description section of learn HTML forms.md (#46859)\n\n* Update 60fad1cafcde010995e15306.md\r\n\r\nHello, \r\nI have changed the sentence, please have a look.\r\n\r\n* Update curriculum/challenges/english/14-responsive-web-design-22/learn-html-forms-by-building-a-registration-form/60fad1cafcde010995e15306.md\r\n\r\nThank you, I have committed the changes suggested by you.\r\n\r\nCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>\r\n\r\nCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>\n```',
                    ranges: [[5, 0, 5]],
                },
                {
                    type: 'commit',
                    url: '/github.com/EbookFoundation/free-programming-books/-/commit/c019ebd5db296ff9fba99a2ebb73ba5d3aa5f854',
                    repository: 'github.com/EbookFoundation/free-programming-books',
                    oid: 'c019ebd5db296ff9fba99a2ebb73ba5d3aa5f854',
                    message:
                        "Change link of repojacking vulnerable link (#7723)\n\nHello from Hacktoberfest :)\r\nThe link to https://raw.githubusercontent.com/teten777/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf is vulnerable to repojacking (it redirects to the orignial project that changed name), you should change the link to the current name of the project. if you won't change the link, an attacker can open the linked repository and attacks users that trust your links.\r\nThe new repository name is:\r\nhttps://raw.githubusercontent.com/teten-nugraha/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf",
                    authorName: 'yakirk',
                    authorDate: '2022-10-07T14:42:09Z',
                    committerName: 'GitHub',
                    committerDate: '2022-10-07T14:42:09Z',
                    repoStars: 258422,
                    repoLastFetched: '2022-12-15T13:48:14.891675Z',
                    content:
                        "```COMMIT_EDITMSG\nChange link of repojacking vulnerable link (#7723)\n\nHello from Hacktoberfest :)\r\nThe link to https://raw.githubusercontent.com/teten777/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf is vulnerable to repojacking (it redirects to the orignial project that changed name), you should change the link to the current name of the project. if you won't change the link, an attacker can open the linked repository and attacks users that trust your links.\r\nThe new repository name is:\r\nhttps://raw.githubusercontent.com/teten-nugraha/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf\n```",
                    ranges: [[3, 0, 5]],
                },
                {
                    type: 'commit',
                    url: '/github.com/EbookFoundation/free-programming-books/-/commit/a55f727e669f2a6ce694ac5821ff06a15a889328',
                    repository: 'github.com/EbookFoundation/free-programming-books',
                    oid: 'a55f727e669f2a6ce694ac5821ff06a15a889328',
                    message:
                        'New Django Course! (#6945)\n\nI have added a new course under the Django section. It was undoubtedly a must. It teaches everything from head to toe! From creating a simple **Hello World** site to a fully functional website to deployment! All is here!',
                    authorName: 'Muhammad Anas',
                    authorDate: '2022-07-21T05:52:57Z',
                    committerName: 'GitHub',
                    committerDate: '2022-07-21T05:52:57Z',
                    repoStars: 258422,
                    repoLastFetched: '2022-12-15T13:48:14.891675Z',
                    content:
                        '```COMMIT_EDITMSG\nNew Django Course! (#6945)\n\nI have added a new course under the Django section. It was undoubtedly a must. It teaches everything from head to toe! From creating a simple **Hello World** site to a fully functional website to deployment! All is here!\n```',
                    ranges: [[3, 144, 5]],
                },
                {
                    type: 'commit',
                    url: '/github.com/EbookFoundation/free-programming-books/-/commit/e364fc5254fedbbdbf179912aa59dfd773519d14',
                    repository: 'github.com/EbookFoundation/free-programming-books',
                    oid: 'e364fc5254fedbbdbf179912aa59dfd773519d14',
                    message:
                        'Added Hello SDL by LazyFoo (#4956)\n\n* Addded Hello SDL by LazyFoo\r\n\r\n* removed the redundant part of the URL',
                    authorName: 'Diego Mateos',
                    authorDate: '2020-10-31T20:49:54Z',
                    committerName: 'GitHub',
                    committerDate: '2020-10-31T20:49:54Z',
                    repoStars: 258422,
                    repoLastFetched: '2022-12-15T13:48:14.891675Z',
                    content:
                        '```COMMIT_EDITMSG\nAdded Hello SDL by LazyFoo (#4956)\n\n* Addded Hello SDL by LazyFoo\r\n\r\n* removed the redundant part of the URL\n```',
                    ranges: [[1, 6, 5]],
                },
            ],

            'Match type,Repository,Repository external URL,Date,Author,Subject,oid,Commit URL\ncommit,github.com/EbookFoundation/free-programming-books,http://localhost:3443/github.com/EbookFoundation/free-programming-books,2022-10-30T03:07:11Z,Karen Ketlyn Ferreira Barcelos,"Added two courses R PT_BR (#7415)* Added two courses R PT_BR\r\rHello! I\'m trying to help by adding courses that helped me!\r\r* fix: URL fragments goes in lowercase\r\r* lint: leave 2 blank lines between listings and next heading\r\r* Update courses/free-courses-pt_BR.md\r\rCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\r* Update courses/free-courses-pt_BR.md\r\rCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\r\rCo-authored-by: David Ordás <3125580+davorpa@users.noreply.github.com>\rCo-authored-by: Eric Hellman <eric@hellman.net>",91cdcd907bbf070b1310f9120b06bb03f481edb6,http://localhost:3443/github.com/EbookFoundation/free-programming-books/-/commit/91cdcd907bbf070b1310f9120b06bb03f481edb6\ncommit,github.com/freeCodeCamp/freeCodeCamp,http://localhost:3443/github.com/freeCodeCamp/freeCodeCamp,2022-07-12T07:13:40Z,Harsh Kashyap,"fix(curriculum): corrected grammar in line 12 of the description section of learn HTML forms.md (#46859)* Update 60fad1cafcde010995e15306.md\r\rHello, \rI have changed the sentence, please have a look.\r\r* Update curriculum/challenges/english/14-responsive-web-design-22/learn-html-forms-by-building-a-registration-form/60fad1cafcde010995e15306.md\r\rThank you, I have committed the changes suggested by you.\r\rCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>\r\rCo-authored-by: Jeremy L Thompson <jeremy@jeremylt.org>",3ecc381f95383dfa8314846a5a7d815b359d95d1,http://localhost:3443/github.com/freeCodeCamp/freeCodeCamp/-/commit/3ecc381f95383dfa8314846a5a7d815b359d95d1\ncommit,github.com/EbookFoundation/free-programming-books,http://localhost:3443/github.com/EbookFoundation/free-programming-books,2022-10-07T14:42:09Z,yakirk,"Change link of repojacking vulnerable link (#7723)Hello from Hacktoberfest :)\rThe link to https://raw.githubusercontent.com/teten777/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf is vulnerable to repojacking (it redirects to the orignial project that changed name), you should change the link to the current name of the project. if you won\'t change the link, an attacker can open the linked repository and attacks users that trust your links.\rThe new repository name is:\rhttps://raw.githubusercontent.com/teten-nugraha/free-ebook-springboot-basic/master/Memulai%20Java%20Enterprise%20dengan%20Spring%20Boot.pdf",c019ebd5db296ff9fba99a2ebb73ba5d3aa5f854,http://localhost:3443/github.com/EbookFoundation/free-programming-books/-/commit/c019ebd5db296ff9fba99a2ebb73ba5d3aa5f854\ncommit,github.com/EbookFoundation/free-programming-books,http://localhost:3443/github.com/EbookFoundation/free-programming-books,2022-07-21T05:52:57Z,Muhammad Anas,"New Django Course! (#6945)I have added a new course under the Django section. It was undoubtedly a must. It teaches everything from head to toe! From creating a simple **Hello World** site to a fully functional website to deployment! All is here!",a55f727e669f2a6ce694ac5821ff06a15a889328,http://localhost:3443/github.com/EbookFoundation/free-programming-books/-/commit/a55f727e669f2a6ce694ac5821ff06a15a889328\ncommit,github.com/EbookFoundation/free-programming-books,http://localhost:3443/github.com/EbookFoundation/free-programming-books,2020-10-31T20:49:54Z,Diego Mateos,"Added Hello SDL by LazyFoo (#4956)* Addded Hello SDL by LazyFoo\r\r* removed the redundant part of the URL",e364fc5254fedbbdbf179912aa59dfd773519d14,http://localhost:3443/github.com/EbookFoundation/free-programming-books/-/commit/e364fc5254fedbbdbf179912aa59dfd773519d14',
        ],
        [
            'owner',
            [
                {
                    type: 'person',
                    handle: 'alice',
                    email: 'alice@example.com',
                    user: {
                        username: 'alice',
                        displayName: 'Alice Example',
                        avatarURL: 'https://example.com',
                    },
                },
                {
                    type: 'person',
                    handle: 'bob',
                    email: '',
                },
                {
                    type: 'team',
                    handle: 'example-team',
                    email: 'example-team@example.com',
                    name: 'example-team',
                    displayName: 'Example Team',
                },
            ],
            'Match type,Handle,Email,User or team name,Display name,Profile URL\nperson,alice,alice@example.com,alice,Alice Example,http://localhost:3443/users/alice\nperson,bob,,,,\nteam,example-team,example-team@example.com,example-team,Example Team,http://localhost:3443/teams/example-team',
        ],
    ]

    test.each(data)('returns correct content for searchType "%s"', (searchType, results, content) => {
        expect(searchResultsToFileContent(results, sourcegraphURL)).toEqual(content)
    })

    test('returns correct content for repo match with enableRepositoryMetadata=true', () => {
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        const [, results] = data.find(([searchType]) => searchType === 'repo')!
        expect(searchResultsToFileContent(results, sourcegraphURL, true)).toEqual(
            'Match type,Repository,Repository external URL,Repository metadata\nrepo,github.com/lorenzodifuccia/safaribooks,http://localhost:3443/github.com/lorenzodifuccia/safaribooks,"oss\ndeprecated\nsome"",non-standard-key:value"\nrepo,github.com/rfletcher/safari-json-formatter,http://localhost:3443/github.com/rfletcher/safari-json-formatter,""\nrepo,github.com/AdguardTeam/AdGuardForSafari,http://localhost:3443/github.com/AdguardTeam/AdGuardForSafari,""\nrepo,github.com/kishikawakatsumi/SourceKitForSafari,http://localhost:3443/github.com/kishikawakatsumi/SourceKitForSafari,""\nrepo,github.com/shaojiankui/iOS-UDID-Safari,http://localhost:3443/github.com/shaojiankui/iOS-UDID-Safari,""'
        )
    })
})

describe('buildFileName', () => {
    const data = [
        ['context:global code host ', 'sourcegraph-search-export-context-global-code-host.csv'],
        [
            'context:global code host count:200 type:path ',
            'sourcegraph-search-export-context-global-code-host-count-200-type-path.csv',
        ],
        [
            'context:global code host count:200 type:commit',
            'sourcegraph-search-export-context-global-code-host-count-200-type-commit.csv',
        ],
        ['context:global nice type:diff', 'sourcegraph-search-export-context-global-nice-type-diff.csv'],
        ['context:global reac.. json', 'sourcegraph-search-export-context-global-reac---json.csv'],
        ['context:global function greeting( ', 'sourcegraph-search-export-context-global-function-greeting-.csv'],
        // Test that very long queries are truncated to less than 255 characters
        [
            'context:global (repo:^github.com/sourcegraph/sourcegraph$ OR repo:^github.com/sourcegraph/sourcegraph-jetbrains$ OR repo:^github.com/sourcegraph/prototools$ ) count:all (lang:Java OR lang:Python OR lang:C++ OR lang:JavaScript) select:file',
            'sourcegraph-search-export-context-global--repo--github-com-sourcegraph-sourcegraph--OR-repo--github-com-sourcegraph-sourcegraph-jetbrains--OR-repo--github-com-sourcegraph-prototools----count-all--lang-Java-OR-lang-Python-OR-lang-C---OR-lang-JavaScript.csv',
        ],
    ]

    test.each(data)('builds correct fileName from query "%s"', (query, fileName) => {
        expect(buildFileName(query)).toEqual(fileName)
    })
})
