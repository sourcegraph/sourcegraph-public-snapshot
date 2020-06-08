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
        stat {
            added
            changed
            deleted
        }
        internalID
    }
`
