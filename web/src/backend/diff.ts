import { gql } from '../../../shared/src/graphql/graphql'

export const FileDiffHunkRangeFields = gql`
    fragment FileDiffHunkRangeFields on FileDiffHunkRange {
        startLine
        lines
    }
`

export const DiffStatFields = gql`
    fragment DiffStatFields on DiffStat {
        added
        changed
        deleted
    }
`

export const FileDiffHunkFields = gql`
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
        highlight(disableTimeout: false, isLightTheme: $isLightTheme) {
            aborted
            lines {
                kind
                html
            }
        }
    }
`

export const FileDiffFields = gql`
    fragment FileDiffFields on FileDiff {
        __typename
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
        }
        hunks {
            ...FileDiffHunkFields
        }
        stat {
            added
            changed
            deleted
        }
        internalID
    }

    ${FileDiffHunkFields}
`
