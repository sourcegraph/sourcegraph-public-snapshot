import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { NEVER } from 'rxjs'

import { MarkupKind } from '@sourcegraph/extension-api-classes'

import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { subtypeOf } from '../util/types'

import { HoverOverlay, HoverOverlayProps } from './HoverOverlay'

jest.mock('../actions/ActionItem', () => ({
    ActionItem: 'ActionItem',
}))

describe('HoverOverlay', () => {
    const NOOP_EXTENSIONS_CONTROLLER = { executeCommand: () => Promise.resolve() }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => undefined, settings: NEVER }
    const history = H.createMemoryHistory({ keyLength: 0 })
    const commonProps = subtypeOf<HoverOverlayProps>()({
        location: history.location,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        extensionsController: NOOP_EXTENSIONS_CONTROLLER,
        platformContext: NOOP_PLATFORM_CONTEXT,
        showCloseButton: false,
        hoveredToken: { repoName: 'r', commitID: 'c', revision: 'v', filePath: 'f', line: 1, character: 2 },
        overlayPosition: { left: 0, top: 0 },
        isLightTheme: false,
    })

    test('actions and hover undefined', () => {
        expect(renderer.create(<HoverOverlay {...commonProps} />).toJSON()).toMatchSnapshot()
    })

    test('actions loading', () => {
        expect(renderer.create(<HoverOverlay {...commonProps} actionsOrError="loading" />).toJSON()).toMatchSnapshot()
    })

    test('hover loading', () => {
        expect(renderer.create(<HoverOverlay {...commonProps} hoverOrError="loading" />).toJSON()).toMatchSnapshot()
    })

    test('actions and hover loading', () => {
        expect(
            renderer.create(<HoverOverlay {...commonProps} actionsOrError="loading" hoverOrError="loading" />).toJSON()
        ).toMatchSnapshot()
    })

    test('actions empty', () => {
        const component = renderer.create(<HoverOverlay {...commonProps} actionsOrError={[]} />)
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('hover empty', () => {
        expect(renderer.create(<HoverOverlay {...commonProps} hoverOrError={null} />).toJSON()).toMatchSnapshot()
    })

    test('actions and hover empty', () => {
        expect(
            renderer.create(<HoverOverlay {...commonProps} actionsOrError={[]} hoverOrError={null} />).toJSON()
        ).toMatchSnapshot()
    })

    test('actions present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={[{ action: { id: 'a', command: 'c', title: 'Some title' }, active: true }]}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('hover present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('multiple hovers present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        hoverOrError={{
                            contents: [
                                { kind: MarkupKind.Markdown, value: 'v' },
                                { kind: MarkupKind.Markdown, value: 'v2' },
                            ],
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('actions and hover present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={[{ action: { id: 'a', command: 'c' }, active: true }]}
                        hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('actions, hover and alert present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={[{ action: { id: 'a', command: 'c' }, active: true }]}
                        hoverOrError={{
                            contents: [{ kind: MarkupKind.Markdown, value: 'v' }],
                            alerts: [
                                {
                                    summary: {
                                        kind: MarkupKind.Markdown,
                                        value: 'Testing `markdown` rendering.',
                                    },
                                    type: 'test-alert-dismissalType',
                                },
                            ],
                        }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('actions present, hover loading', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={[{ action: { id: 'a', command: 'c' }, active: true }]}
                        hoverOrError="loading"
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('hover present, actions loading', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError="loading"
                        hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('actions error', () => {
        expect(
            renderer.create(<HoverOverlay {...commonProps} actionsOrError={{ message: 'm', name: 'c' }} />).toJSON()
        ).toMatchSnapshot()
    })

    test('hover error', () => {
        expect(
            renderer.create(<HoverOverlay {...commonProps} hoverOrError={{ message: 'm', name: 'c' }} />).toJSON()
        ).toMatchSnapshot()
    })

    test('actions and hover error', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={{ message: 'm1', name: 'c1' }}
                        hoverOrError={{ message: 'm2', name: 'c2' }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('actions error, hover present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={{ message: 'm', name: 'c' }}
                        hoverOrError={{ contents: [{ kind: MarkupKind.Markdown, value: 'v' }] }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('hover error, actions present', () => {
        expect(
            renderer
                .create(
                    <HoverOverlay
                        {...commonProps}
                        actionsOrError={[{ action: { id: 'a', command: 'c' }, active: true }]}
                        hoverOrError={{ message: 'm', name: 'c' }}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
