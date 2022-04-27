import { render } from '@testing-library/react'
import * as H from 'history'
import { noop } from 'lodash'
import { concat, EMPTY, NEVER, of } from 'rxjs'
import * as sinon from 'sinon'

import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { Location } from '@sourcegraph/extension-api-types'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { Controller } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { HierarchicalLocationsView, HierarchicalLocationsViewProps } from './HierarchicalLocationsView'

describe('<HierarchicalLocationsView />', () => {
    const getProps = () => {
        const registerContributions = sinon.spy<FlatExtensionHostAPI['registerContributions']>(() =>
            pretendProxySubscribable(EMPTY).subscribe(noop as any)
        )

        const extensionsController: Pick<Controller, 'extHostAPI'> = {
            extHostAPI: Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    updateContext: () => Promise.resolve(),
                    registerContributions,
                })
            ),
        }
        const settingsCascade: SettingsCascadeOrError = {
            subjects: null,
            final: null,
        }
        const location: H.Location = {
            hash: '#L36:18&tab=references',
            pathname: '/github.com/sourcegraph/sourcegraph/-/blob/browser/src/libs/phabricator/index.tsx',
            search: '',
            state: {},
        }

        const props: HierarchicalLocationsViewProps = {
            extensionsController,
            settingsCascade,
            location,
            locations: NEVER,
            defaultGroup: 'git://github.com/foo/bar',
            isLightTheme: true,
            fetchHighlightedFileLineRanges: sinon.spy(),
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }
        return { props, registerContributions }
    }

    test('shows a spinner before any locations emissions', () => {
        const { props } = getProps()
        expect(render(<HierarchicalLocationsView {...props} />).asFragment()).toMatchSnapshot()
    })

    test('shows a spinner if locations emits empty and is not complete', () => {
        const { props } = getProps()
        expect(
            render(
                <HierarchicalLocationsView {...props} locations={concat(of({ isLoading: true, result: [] }), NEVER)} />
            ).asFragment()
        ).toMatchSnapshot()
    })

    const SAMPLE_LOCATION: Location = {
        uri: 'git://github.com/foo/bar',
        range: {
            start: {
                line: 1,
                character: 0,
            },
            end: {
                line: 1,
                character: 10,
            },
        },
    }

    test('displays a single location when complete', () => {
        const locations = of<MaybeLoadingResult<Location[]>>({ isLoading: false, result: [SAMPLE_LOCATION] })
        const props = {
            ...getProps().props,
            locations,
        }
        expect(renderWithBrandedContext(<HierarchicalLocationsView {...props} />).asFragment()).toMatchSnapshot()
    })

    test('displays partial locations before complete', () => {
        const props = {
            ...getProps().props,
            locations: concat(of({ isLoading: false, result: [SAMPLE_LOCATION] }), NEVER),
        }
        expect(renderWithBrandedContext(<HierarchicalLocationsView {...props} />).asFragment()).toMatchSnapshot()
    })

    test('displays multiple locations grouped by file', () => {
        const locations: Location[] = [
            {
                uri: 'git://github.com/foo/bar#file1.txt',
                range: {
                    start: {
                        line: 1,
                        character: 0,
                    },
                    end: {
                        line: 1,
                        character: 10,
                    },
                },
            },
            {
                uri: 'git://github.com/foo/bar#file2.txt',
                range: {
                    start: {
                        line: 2,
                        character: 0,
                    },
                    end: {
                        line: 2,
                        character: 10,
                    },
                },
            },
            {
                uri: 'git://github.com/foo/bar#file1.txt',
                range: {
                    start: {
                        line: 3,
                        character: 0,
                    },
                    end: {
                        line: 3,
                        character: 10,
                    },
                },
            },
            {
                uri: 'git://github.com/foo/bar#file2.txt',
                range: {
                    start: {
                        line: 4,
                        character: 0,
                    },
                    end: {
                        line: 4,
                        character: 10,
                    },
                },
            },
            {
                uri: 'git://github.com/foo/bar#file2.txt',
                range: {
                    start: {
                        line: 5,
                        character: 0,
                    },
                    end: {
                        line: 5,
                        character: 10,
                    },
                },
            },
        ]
        const props: HierarchicalLocationsViewProps = {
            ...getProps().props,
            settingsCascade: {
                subjects: null,
                final: {
                    'panel.locations.groupByFile': true,
                },
            },
            locations: of({ isLoading: false, result: locations }),
        }
        expect(renderWithBrandedContext(<HierarchicalLocationsView {...props} />).asFragment()).toMatchSnapshot()
    })
})
