import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { TreeEntriesSection } from './TreeEntriesSection'

describe('TreeEntriesSection', () => {
    it('should render a grid of tree entries at the root', () => {
        expect(
            renderWithBrandedContext(
                <TreeEntriesSection
                    parentPath=""
                    entries={[
                        {
                            name: 'src',
                            path: 'src',
                            isDirectory: true,
                            url: '/github.com/sourcegraph/codeintellify/-/tree/src',
                        },
                        {
                            name: 'testdata',
                            path: 'testdata',
                            isDirectory: true,
                            url: '/github.com/sourcegraph/codeintellify/-/tree/testdata',
                        },
                        {
                            name: '.editorconfig',
                            path: '.editorconfig',
                            isDirectory: false,
                            url: '/github.com/sourcegraph/codeintellify/-/blob/.editorconfig',
                        },
                        {
                            name: '.eslintrc.json',
                            path: '.eslintrc.json',
                            isDirectory: false,
                            url: '/github.com/sourcegraph/codeintellify/-/blob/.eslintrc.json',
                        },
                    ]}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
    it('should render a grid of tree entries in a subdirectory', () => {
        expect(
            renderWithBrandedContext(
                <TreeEntriesSection
                    parentPath="src"
                    entries={[
                        {
                            name: 'testutils',
                            path: 'src/testutils',
                            isDirectory: true,
                            url: '/github.com/sourcegraph/codeintellify/-/tree/src/testutils',
                        },
                        {
                            name: 'typings',
                            path: 'src/typings',
                            isDirectory: true,
                            url: '/github.com/sourcegraph/codeintellify/-/tree/src/typings',
                        },
                        {
                            name: 'errors.ts',
                            path: 'src/errors.ts',
                            isDirectory: false,
                            url: '/github.com/sourcegraph/codeintellify/-/blob/src/errors.ts',
                        },
                        {
                            name: 'helpers.ts',
                            path: 'src/helpers.ts',
                            isDirectory: false,
                            url: '/github.com/sourcegraph/codeintellify/-/blob/src/helpers.ts',
                        },
                    ]}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
    it('should render only direct children', () => {
        expect(
            renderWithBrandedContext(
                <TreeEntriesSection
                    parentPath="x"
                    entries={[
                        {
                            name: 'ref',
                            path: 'x/ref',
                            isDirectory: true,
                            url: '/github.com/vanadium/core/-/tree/x/ref',
                        },
                        {
                            name: 'cmd',
                            path: 'x/ref/cmd',
                            isDirectory: true,
                            url: '/github.com/vanadium/core/-/tree/x/ref/cmd',
                        },
                        {
                            name: 'examples',
                            path: 'x/ref/examples',
                            isDirectory: true,
                            url: '/github.com/vanadium/core/-/tree/x/ref/examples',
                        },
                        {
                            name: 'README.md',
                            path: 'x/ref/README.md',
                            isDirectory: false,
                            url: '/github.com/vanadium/core/-/blob/x/ref/README.md',
                        },
                        {
                            name: 'envvar.go',
                            path: 'x/ref/envvar.go',
                            isDirectory: false,
                            url: '/github.com/vanadium/core/-/blob/x/ref/envvar.go',
                        },
                    ]}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
