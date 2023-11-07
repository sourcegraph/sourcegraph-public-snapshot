import { describe, expect, test } from '@jest/globals'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ExternalServiceKind } from '../../graphql-operations'

import { CommitMessageWithLinks } from './CommitMessageWithLinks'

describe('CommitMessageWithLinks', () => {
    test('works with a commit link', () => {
        const content = renderWithBrandedContext(
            <CommitMessageWithLinks
                message="dev/sg: migrate to urfave/cli (#1234)"
                to="/foo/bar"
                className=""
                externalURLs={[
                    {
                        serviceKind: ExternalServiceKind.GITHUB,
                        url: 'https://github.com/sourcegraph/sourcegraph/commit/aad9f1050b914041feb8d965deaa49a301cb3a28',
                    },
                ]}
            />
        )
        expect(content.asFragment()).toMatchSnapshot()
    })

    test('works with a repo link', () => {
        const content = renderWithBrandedContext(
            <CommitMessageWithLinks
                message="dev/sg: migrate to urfave/cli (#1234)"
                to="/foo/bar"
                className=""
                externalURLs={[
                    {
                        serviceKind: ExternalServiceKind.GITHUB,
                        url: 'https://github.com/sourcegraph/sourcegraph',
                    },
                ]}
            />
        )
        expect(content.asFragment()).toMatchSnapshot()
    })
})
