import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import type { Progress } from '@sourcegraph/shared/src/search/stream'

import { StreamingProgressCount } from './StreamingProgressCount'

describe('StreamingProgressCount', () => {
    it('should render correctly for 0 items in progress', () => {
        const progress: Progress = {
            durationMs: 0,
            matchCount: 0,
            skipped: [],
        }

        expect(render(<StreamingProgressCount state="loading" progress={progress} />).asFragment()).toMatchSnapshot()
    })

    it('should render correctly for 0 repositories', () => {
        const progress: Progress = {
            durationMs: 0,
            matchCount: 0,
            repositoriesCount: 0,
            skipped: [],
        }

        expect(render(<StreamingProgressCount state="loading" progress={progress} />).asFragment()).toMatchSnapshot()
    })

    it('should render correctly for 1 item complete', () => {
        const progress: Progress = {
            durationMs: 1250,
            matchCount: 1,
            repositoriesCount: 1,
            skipped: [],
        }

        expect(render(<StreamingProgressCount state="complete" progress={progress} />).asFragment()).toMatchSnapshot()
    })

    it('should render correctly for 123 items complete', () => {
        const progress: Progress = {
            durationMs: 1250,
            matchCount: 123,
            repositoriesCount: 500,
            skipped: [],
        }

        expect(render(<StreamingProgressCount state="complete" progress={progress} />).asFragment()).toMatchSnapshot()
    })

    it('should render correctly for big numbers complete', () => {
        const progress: Progress = {
            durationMs: 52500,
            matchCount: 1234567,
            repositoriesCount: 8901,
            skipped: [],
        }

        expect(render(<StreamingProgressCount state="complete" progress={progress} />).asFragment()).toMatchSnapshot()
    })

    it('should render correctly for limithit', () => {
        const progress: Progress = {
            durationMs: 1250,
            matchCount: 123,
            repositoriesCount: 500,
            skipped: [
                {
                    reason: 'document-match-limit',
                    title: 'match limit',
                    message: 'match limit',
                    severity: 'warn',
                },
            ],
        }

        expect(render(<StreamingProgressCount state="complete" progress={progress} />).asFragment()).toMatchSnapshot()
    })
})
