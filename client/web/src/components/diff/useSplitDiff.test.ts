import { renderHook, act } from '@testing-library/react-hooks'
import { DiffHunkLineType } from '../../graphql-operations'
import { useSplitDiff } from './useSplitDiff'

describe('useSplitDiff', () => {
    const mockedHunk = [
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">\t}})\n</span></div>',
            oldLine: 200,
            newLine: 200,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L200',
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">defer </span><span style="color:#b58900;">endObservation</span><span style="color:#657b83;">(</span><span style="color:#6c71c4;">1</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">observation</span><span style="color:#657b83;">.</span><span style="color:#268bd2;">Args</span><span style="color:#657b83;">{})\n</span></div>',
            oldLine: 201,
            newLine: 201,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L201',
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            oldLine: 202,
            newLine: 202,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L202',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#268bd2;">out</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">err </span><span style="color:#657b83;">:= </span><span style="color:#268bd2;">c</span><span style="color:#657b83;">.</span><span style="color:#b58900;">execResolveRevGitCommand</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">ctx</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">repositoryID</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#859900;">append</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">[]</span><span style="color:#859900;">string</span><span style="color:#657b83;">{</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">ls-tree</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">--name-only</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">--</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">}, </span><span style="color:#b58900;">cleanDirectoriesForLsTree</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">)</span><span style="color:#859900;">...</span><span style="color:#657b83;">)</span><span style="color:#859900;">...</span><span style="color:#657b83;">)\n</span></div>',
            oldLine: 203,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L203',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">if </span><span style="color:#268bd2;">err </span><span style="color:#859900;">!= </span><span style="color:#b58900;">nil </span><span style="color:#657b83;">{\n</span></div>',
            oldLine: 204,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L204',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#657b83;">\t\t</span><span style="color:#859900;">return </span><span style="color:#b58900;">nil</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">err\n</span></div>',
            oldLine: 205,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L205',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#657b83;">\t}\n</span></div>',
            oldLine: 206,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L206',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            oldLine: 207,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L207',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#b58900;">parseDirectoryChildren</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">strings</span><span style="color:#657b83;">.</span><span style="color:#b58900;">Split</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">out</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">)), </span><span style="color:#b58900;">nil\n</span></div>',
            oldLine: 208,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L208',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#657b83;">}\n</span></div>',
            oldLine: 209,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L209',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            oldLine: 210,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L210',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#767676;">// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a\n</span></div>',
            oldLine: 211,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L211',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#767676;">// few peculiarities handled here:\n</span></div>',
            oldLine: 212,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L212',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#767676;">//\n</span></div>',
            oldLine: 213,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L213',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            oldLine: 268,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L268',
        },
        {
            kind: DiffHunkLineType.DELETED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">childrenMap\n</span></div>',
            oldLine: 269,
            newLine: 0,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L269',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">pathexistence</span><span style="color:#657b83;">.</span><span style="color:#b58900;">GitGetChildren</span><span style="color:#657b83;">(\n</span></div>',
            oldLine: 0,
            newLine: 203,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R203',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">func</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">args </span><span style="color:#859900;">...string</span><span style="color:#657b83;">) (</span><span style="color:#859900;">string</span><span style="color:#657b83;">, </span><span style="color:#859900;">error</span><span style="color:#657b83;">) {\n</span></div>',
            oldLine: 0,
            newLine: 204,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R204',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#657b83;">\t\t\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">c</span><span style="color:#657b83;">.</span><span style="color:#b58900;">execResolveRevGitCommand</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">ctx</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">repositoryID</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">args</span><span style="color:#859900;">...</span><span style="color:#657b83;">)\n</span></div>',
            oldLine: 0,
            newLine: 205,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R205',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html: '<div><span style="color:#657b83;">\t\t},\n</span></div>',
            oldLine: 0,
            newLine: 206,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R206',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">,\n</span></div>',
            oldLine: 0,
            newLine: 207,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R207',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html:
                '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">,\n</span></div>',
            oldLine: 0,
            newLine: 208,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R208',
        },
        {
            kind: DiffHunkLineType.ADDED,
            html: '<div><span style="color:#657b83;">\t)\n</span></div>',
            oldLine: 0,
            newLine: 209,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R209',
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">}\n</span></div>',
            oldLine: 270,
            newLine: 210,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L270',
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html: '<div><span style="color:#657b83;">\n</span></div>',
            oldLine: 271,
            newLine: 211,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L271',
        },
        {
            kind: DiffHunkLineType.UNCHANGED,
            html:
                '<div><span style="color:#767676;">// FileExists determines whether a file exists in a particular commit of a repository.\n</span></div>',
            oldLine: 272,
            newLine: 212,
            anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L272',
        },
    ]

    test('will zip the data structure', () => {
        const expected = [
            {
                kind: 'UNCHANGED',
                html: '<div><span style="color:#657b83;">\t}})\n</span></div>',
                oldLine: 200,
                newLine: 200,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L200',
            },
            {
                kind: 'UNCHANGED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">defer </span><span style="color:#b58900;">endObservation</span><span style="color:#657b83;">(</span><span style="color:#6c71c4;">1</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">observation</span><span style="color:#657b83;">.</span><span style="color:#268bd2;">Args</span><span style="color:#657b83;">{})\n</span></div>',
                oldLine: 201,
                newLine: 201,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L201',
            },
            {
                kind: 'UNCHANGED',
                html: '<div><span style="color:#657b83;">\n</span></div>',
                oldLine: 202,
                newLine: 202,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L202',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#268bd2;">out</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">err </span><span style="color:#657b83;">:= </span><span style="color:#268bd2;">c</span><span style="color:#657b83;">.</span><span style="color:#b58900;">execResolveRevGitCommand</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">ctx</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">repositoryID</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#859900;">append</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">[]</span><span style="color:#859900;">string</span><span style="color:#657b83;">{</span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">ls-tree</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">--name-only</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#2aa198;">--</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">}, </span><span style="color:#b58900;">cleanDirectoriesForLsTree</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">)</span><span style="color:#859900;">...</span><span style="color:#657b83;">)</span><span style="color:#859900;">...</span><span style="color:#657b83;">)\n</span></div>',
                oldLine: 203,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L203',
            },
            {
                kind: 'ADDED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">pathexistence</span><span style="color:#657b83;">.</span><span style="color:#b58900;">GitGetChildren</span><span style="color:#657b83;">(\n</span></div>',
                oldLine: 0,
                newLine: 203,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R203',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">if </span><span style="color:#268bd2;">err </span><span style="color:#859900;">!= </span><span style="color:#b58900;">nil </span><span style="color:#657b83;">{\n</span></div>',
                oldLine: 204,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L204',
            },
            {
                kind: 'ADDED',
                html:
                    '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">func</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">args </span><span style="color:#859900;">...string</span><span style="color:#657b83;">) (</span><span style="color:#859900;">string</span><span style="color:#657b83;">, </span><span style="color:#859900;">error</span><span style="color:#657b83;">) {\n</span></div>',
                oldLine: 0,
                newLine: 204,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R204',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#657b83;">\t\t</span><span style="color:#859900;">return </span><span style="color:#b58900;">nil</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">err\n</span></div>',
                oldLine: 205,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L205',
            },
            {
                kind: 'ADDED',
                html:
                    '<div><span style="color:#657b83;">\t\t\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">c</span><span style="color:#657b83;">.</span><span style="color:#b58900;">execResolveRevGitCommand</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">ctx</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">repositoryID</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">args</span><span style="color:#859900;">...</span><span style="color:#657b83;">)\n</span></div>',
                oldLine: 0,
                newLine: 205,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R205',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#657b83;">\t}\n</span></div>',
                oldLine: 206,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L206',
            },
            {
                kind: 'ADDED',
                html: '<div><span style="color:#657b83;">\t\t},\n</span></div>',
                oldLine: 0,
                newLine: 206,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R206',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#657b83;">\n</span></div>',
                oldLine: 207,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L207',
            },
            {
                kind: 'ADDED',
                html:
                    '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">commit</span><span style="color:#657b83;">,\n</span></div>',
                oldLine: 0,
                newLine: 207,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R207',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#b58900;">parseDirectoryChildren</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">, </span><span style="color:#268bd2;">strings</span><span style="color:#657b83;">.</span><span style="color:#b58900;">Split</span><span style="color:#657b83;">(</span><span style="color:#268bd2;">out</span><span style="color:#657b83;">, </span><span style="color:#839496;">&#34;</span><span style="color:#dc322f;">\\n</span><span style="color:#839496;">&#34;</span><span style="color:#657b83;">)), </span><span style="color:#b58900;">nil\n</span></div>',
                oldLine: 208,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L208',
            },
            {
                kind: 'ADDED',
                html:
                    '<div><span style="color:#657b83;">\t\t</span><span style="color:#268bd2;">dirnames</span><span style="color:#657b83;">,\n</span></div>',
                oldLine: 0,
                newLine: 208,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R208',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#657b83;">}\n</span></div>',
                oldLine: 209,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L209',
            },
            {
                kind: 'ADDED',
                html: '<div><span style="color:#657b83;">\t)\n</span></div>',
                oldLine: 0,
                newLine: 209,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3R209',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#657b83;">\n</span></div>',
                oldLine: 210,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L210',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#767676;">// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a\n</span></div>',
                oldLine: 211,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L211',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#767676;">// few peculiarities handled here:\n</span></div>',
                oldLine: 212,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L212',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#767676;">//\n</span></div>',
                oldLine: 213,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L213',
            },
            {
                kind: 'DELETED',
                html: '<div><span style="color:#657b83;">\n</span></div>',
                oldLine: 268,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L268',
            },
            {
                kind: 'DELETED',
                html:
                    '<div><span style="color:#657b83;">\t</span><span style="color:#859900;">return </span><span style="color:#268bd2;">childrenMap\n</span></div>',
                oldLine: 269,
                newLine: 0,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L269',
            },
            {
                kind: 'UNCHANGED',
                html: '<div><span style="color:#657b83;">}\n</span></div>',
                oldLine: 270,
                newLine: 210,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L270',
            },
            {
                kind: 'UNCHANGED',
                html: '<div><span style="color:#657b83;">\n</span></div>',
                oldLine: 271,
                newLine: 211,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L271',
            },
            {
                kind: 'UNCHANGED',
                html:
                    '<div><span style="color:#767676;">// FileExists determines whether a file exists in a particular commit of a repository.\n</span></div>',
                oldLine: 272,
                newLine: 212,
                anchor: 'diff-f97c6fe6cd53340835033f861a2f45c3L272',
            },
        ]
        const { result } = renderHook(() => useSplitDiff(mockedHunk))

        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        act(() => {
            // eslint-disable-next-line no-unused-expressions
            result.current.diff
        })

        expect(result.current.diff).toStrictEqual(expected)
    })
})
