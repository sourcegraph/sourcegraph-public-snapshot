import { gql } from '@sourcegraph/http-client'

export const fileDiffHunkRangeFields = gql`
    fragment FileDiffHunkRangeFields on FileDiffHunkRange {
        startLine
        lines
    }
`

export const diffStatFields = gql`
    fragment DiffStatFields on DiffStat {
        __typename
        added
        deleted
    }
`

export const fileDiffHunkFields = gql`
    fragment FileDiffHunkFields on FileDiffHunk {
        oldRange {
            startLine
            lines
        }
        oldNoNewlineAt
        newRange {
            startLine
            lines
        }
        section
        highlight(disableTimeout: false) {
            aborted
            lines {
                kind
                html
            }
        }
    }
`

export const fileDiffFields = gql`
    fragment FileDiffFields on FileDiff {
        oldPath
        oldFile {
            __typename
            binary
            byteSize
        }
        newFile {
            __typename
            binary
            byteSize
        }
        newPath
        mostRelevantFile {
            __typename
            url
            changelistURL
        }
        hunks {
            ...FileDiffHunkFields
        }
        stat {
            added
            deleted
        }
        internalID
    }

    ${fileDiffHunkFields}
`
