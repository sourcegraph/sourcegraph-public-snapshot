import { renderHook, act } from '@testing-library/react-hooks'
import { DiffHunkLineType } from './useSplitDiff'
import { useSplitDiff } from './'

describe('useSplitDiff', () => {
    const mockedHunk = [
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">React</span><span style="color:#657b83;">, { </span><span style="color:#268bd2;">useMemo </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">react</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 1,
            oldLine: 1,
            line: 1,
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">from </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs</span><span style="color:#839496;">&#39;\n</span></div>',
            oldLine: 2,
            line: 2,
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">combineLatest</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">from</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ReplaySubject </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 2,
            line: 2,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">switchMap </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs/operators</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 3,
            oldLine: 3,
            line: 3,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">getContributedActionItems </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../contributions/contributions</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 4,
            oldLine: 4,
            line: 4,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">TelemetryProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../telemetry/telemetryService</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 5,
            oldLine: 5,
            line: 5,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionItem</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ActionItemProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionItem</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 6,
            oldLine: 6,
            line: 6,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionsProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionsContainer</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 7,
            oldLine: 7,
            line: 7,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">classNames </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">classnames</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 8,
            oldLine: 8,
            line: 8,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">useObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../util/useObservable</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 9,
            oldLine: 9,
            line: 9,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">wrapRemoteObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../api/client/api/common</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 10,
            oldLine: 10,
            line: 10,
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">Context</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ContributionScope </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../api/extension/api/context/context</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 11,
            line: 11,
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">useDeepCompareEffectNoCheck </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">use-deep-compare-effect</span><span style="color:#839496;">&#39;\n</span></div>',
            newLine: 12,
            line: 12,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            newLine: 13,
            oldLine: 11,
            line: 13,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#859900;">export </span><span style="color:#268bd2;">interface </span><span style="color:#b58900;">ActionNavItemsClassProps </span><span style="color:#657b83;">{\n</span></div>',
            newLine: 14,
            oldLine: 12,
            line: 14,
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">    </span><span style="color:#767676;">/**\n</span></div>',
            newLine: 15,
            oldLine: 13,
            line: 15,
        },
    ]

    test.only('will transform data correctly', () => {
        const expected = [
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">React</span><span style="color:#657b83;">, { </span><span style="color:#268bd2;">useMemo </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">react</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 1,
                    oldLine: 1,
                    line: 1,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">React</span><span style="color:#657b83;">, { </span><span style="color:#268bd2;">useMemo </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">react</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 1,
                    oldLine: 1,
                    line: 1,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.DELETED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">from </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs</span><span style="color:#839496;">&#39;\n</span></div>',
                    oldLine: 2,
                    line: 2,
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">combineLatest</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">from</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ReplaySubject </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 2,
                    line: 2,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">switchMap </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs/operators</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 3,
                    oldLine: 3,
                    line: 3,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">switchMap </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">rxjs/operators</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 3,
                    oldLine: 3,
                    line: 3,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">getContributedActionItems </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../contributions/contributions</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 4,
                    oldLine: 4,
                    line: 4,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">getContributedActionItems </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../contributions/contributions</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 4,
                    oldLine: 4,
                    line: 4,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">TelemetryProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../telemetry/telemetryService</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 5,
                    oldLine: 5,
                    line: 5,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">TelemetryProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../telemetry/telemetryService</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 5,
                    oldLine: 5,
                    line: 5,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionItem</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ActionItemProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionItem</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 6,
                    oldLine: 6,
                    line: 6,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionItem</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ActionItemProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionItem</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 6,
                    oldLine: 6,
                    line: 6,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionsProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionsContainer</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 7,
                    oldLine: 7,
                    line: 7,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">ActionsProps </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">./ActionsContainer</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 7,
                    oldLine: 7,
                    line: 7,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">classNames </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">classnames</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 8,
                    oldLine: 8,
                    line: 8,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#268bd2;">classNames </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">classnames</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 8,
                    oldLine: 8,
                    line: 8,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">useObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../util/useObservable</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 9,
                    oldLine: 9,
                    line: 9,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">useObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../util/useObservable</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 9,
                    oldLine: 9,
                    line: 9,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">wrapRemoteObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../api/client/api/common</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 10,
                    oldLine: 10,
                    line: 10,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">wrapRemoteObservable </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../api/client/api/common</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 10,
                    oldLine: 10,
                    line: 10,
                },
            ],
            [
                undefined,
                {
                    kind: DiffHunkLineType.ADDED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">Context</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">ContributionScope </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">../api/extension/api/context/context</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 11,
                    line: 11,
                },
            ],
            [
                undefined,
                {
                    kind: DiffHunkLineType.ADDED,
                    html:
                        '<div><span style="color:#cb4b16;">import </span><span style="color:#657b83;">{ </span><span style="color:#268bd2;">useDeepCompareEffectNoCheck </span><span style="color:#657b83;">} </span><span style="color:#cb4b16;">from </span><span style="color:#839496;">&#39;</span><span style="color:#2aa198;">use-deep-compare-effect</span><span style="color:#839496;">&#39;\n</span></div>',
                    newLine: 12,
                    line: 12,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '<div><span style="color:#657b83;">\n</span></div>',
                    newLine: 13,
                    oldLine: 11,
                    line: 13,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '<div><span style="color:#657b83;">\n</span></div>',
                    newLine: 13,
                    oldLine: 11,
                    line: 13,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#859900;">export </span><span style="color:#268bd2;">interface </span><span style="color:#b58900;">ActionNavItemsClassProps </span><span style="color:#657b83;">{\n</span></div>',
                    newLine: 14,
                    oldLine: 12,
                    line: 14,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#859900;">export </span><span style="color:#268bd2;">interface </span><span style="color:#b58900;">ActionNavItemsClassProps </span><span style="color:#657b83;">{\n</span></div>',
                    newLine: 14,
                    oldLine: 12,
                    line: 14,
                },
            ],
            [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#657b83;">    </span><span style="color:#767676;">/**\n</span></div>',
                    newLine: 15,
                    oldLine: 13,
                    line: 15,
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html:
                        '<div><span style="color:#657b83;">    </span><span style="color:#767676;">/**\n</span></div>',
                    newLine: 15,
                    oldLine: 13,
                    line: 15,
                },
            ],
        ]
        const { result } = renderHook(() => useSplitDiff(mockedHunk))

        act(() => {
            result.current.diff
        })

        expect(result.current.diff).toStrictEqual(expected)
    })
})
