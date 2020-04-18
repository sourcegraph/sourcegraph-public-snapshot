import React from 'react'
import renderer from 'react-test-renderer'
import { ViewPage } from './ViewPage'
import * as H from 'history'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { Controller } from '../../../../shared/src/extensions/controller'
import { MarkupKind } from '@sourcegraph/extension-api-classes'

jest.mock('@sourcegraph/react-loading-spinner', () => ({ LoadingSpinner: 'LoadingSpinner' }))
jest.mock('./QueryInputInViewContent', () => ({ QueryInputInViewContent: 'QueryInputInViewContent' }))

const commonProps: Omit<React.ComponentProps<typeof ViewPage>, 'viewID' | 'extraPath' | '_useView'> = {
    settingsCascade: { final: {}, subjects: [] },
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    setCaseSensitivity: () => undefined,
    setPatternType: () => undefined,
    history: H.createMemoryHistory(),
    location: H.createLocation('/'),
    extensionsController: { services: { contribution: { getContributions: () => ({}) } } } as Controller,
}

describe('ViewPage', () => {
    test('view is loading', () => {
        expect(
            renderer.create(<ViewPage {...commonProps} viewID="v" extraPath="" _useView={() => undefined} />).toJSON()
        ).toMatchSnapshot()
    })

    test('view not found', () => {
        expect(
            renderer.create(<ViewPage {...commonProps} viewID="v" extraPath="" _useView={() => null} />).toJSON()
        ).toMatchSnapshot()
    })

    test('renders view', () => {
        expect(
            renderer
                .create(
                    <ViewPage
                        {...commonProps}
                        viewID="v"
                        extraPath=""
                        _useView={() => ({
                            title: 't',
                            content: [
                                {
                                    kind: MarkupKind.Markdown,
                                    value: '**a**',
                                },
                                {
                                    kind: MarkupKind.PlainText,
                                    value: '*b*',
                                },
                                {
                                    value: '*c*',
                                },
                                {
                                    component: 'QueryInput',
                                    props: { implicitQueryPrefix: 'x' },
                                },
                            ],
                        })}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
