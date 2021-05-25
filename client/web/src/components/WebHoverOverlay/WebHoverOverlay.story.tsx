import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { AggregableBadge, HoverAlert } from 'sourcegraph'

import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { registerHighlightContributions } from '@sourcegraph/shared/src/highlight/contributions'
import {
    commonProps,
    FIXTURE_ACTIONS,
    FIXTURE_SEMANTIC_BADGE,
    FIXTURE_CONTENT,
} from '@sourcegraph/shared/src/hover/HoverOverlay.story'

import { WebStory } from '../WebStory'

import { WebHoverOverlay } from './WebHoverOverlay'

registerHighlightContributions()

const { add } = storiesOf('web/WebHoverOverlay', module).addDecorator(story => <WebStory>{() => story()}</WebStory>)

const FIXTURE_CONTENT_LONG_CODE = {
    ...FIXTURE_CONTENT,
    value:
        '```typescript\nexport interface LongTestInterface<A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W, X, Y, Z, A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W, X, Y, Z>\n```\n\n' +
        '---\n\nNisi id deserunt culpa dolore aute pariatur ut amet veniam. Proident id Lorem reprehenderit veniam sunt velit.\n',
}

const FIXTURE_CONTENT_LONG_TEXT_ONLY = {
    ...FIXTURE_CONTENT,
    value:
        'WebAssembly (sometimes abbreviated Wasm) is an open standard that defines a portable binary-code format for executable programs, and a corresponding textual assembly language, as well as interfaces for facilitating interactions between such programs and their host environment. The main goal of WebAssembly is to enable high-performance applications on web pages, but the format is designed to be executed and integrated in other environments as well, including standalone ones.  WebAssembly (i.e. WebAssembly Core Specification (then version 1.0, 1.1 is in draft) and WebAssembly JavaScript Interface) became a World Wide Web Consortium recommendation on 5 December 2019, alongside HTML, CSS, and JavaScript. WebAssembly can support (at least in theory) any language (e.g. compiled or interpreted) on any operating system (with help of appropriate tools), and in practice all of the most popular languages already have at least some level of support.  The Emscripten SDK can compile any LLVM-supported languages (such as C, C++ or Rust, among others) source code into a binary file which runs in the same sandbox as JavaScript code. Emscripten provides bindings for several commonly used environment interfaces like WebGL. There is no direct Document Object Model (DOM) access; however, it is possible to create proxy functions for this, for example through stdweb, web_sys, and js_sys when using the Rust language.  WebAssembly implementations usually use either ahead-of-time (AOT) or just-in-time (JIT) compilation, but may also use an interpreter. While the currently best-known implementations may be in web browsers, there are also non-browser implementations for general-purpose use, including WebAssembly Micro Runtime, a Bytecode Alliance project (with subproject iwasm code VM supporting "interpreter, ahead of time compilation (AoT) and Just-in-Time compilation (JIT)"), Wasmtime, and Wasmer.',
}

const FIXTURE_PARITAL_BADGE: AggregableBadge = {
    ...FIXTURE_SEMANTIC_BADGE,
    text: 'partial semantic',
}

const FIXTURE_SMALL_TEXT_MARKDOWN_ALERT: HoverAlert = {
    summary: {
        kind: MarkupKind.Markdown,
        value:
            '<small>This is an info alert wrapped into a \\<small\\> element. Enim esse quis commodo ex. Pariatur tempor laborum officiairure est do est laborum nostrud cillum. Cupidatat id consectetur et eiusmod Loremproident cupidatat ullamco dolor nostrud. Cupidatat sit do dolor aliqua labore adlaboris cillum deserunt dolor. Sunt labore veniam Lorem reprehenderit quis occaecatsint do mollit aliquip. Consectetur mollit mollit magna eiusmod duis ex. Sint nisilabore labore nulla laboris.</small>',
    },
    type: 'test-alert-type',
}

const FIXTURE_WARNING_MARKDOWN_ALERT: HoverAlert = {
    summary: {
        kind: MarkupKind.Markdown,
        value:
            'This is a warning alert. [It uses Markdown.](https://sourcegraph.com) `To render things easily`. *Cool!*',
    },
    type: 'test-alert-type',
    iconKind: 'warning',
}

add('Loading', () => <WebHoverOverlay {...commonProps()} hoverOrError="loading" actionsOrError={FIXTURE_ACTIONS} />)

add('Error', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={
            new Error(
                'Something terrible happened: Eiusmod voluptate deserunt in sint cillum pariatur laborum eiusmod.'
            )
        }
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('No hover information', () => (
    <WebHoverOverlay {...commonProps()} hoverOrError={null} actionsOrError={FIXTURE_ACTIONS} />
))

add('Common content without actions', () => (
    <WebHoverOverlay {...commonProps()} hoverOrError={{ contents: [FIXTURE_CONTENT] }} />
))

add('Common content with actions', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Aggregated Badges', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Long code', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT_LONG_CODE],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Long text only', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT_LONG_TEXT_ONLY],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('Multiple MarkupContents', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT, FIXTURE_CONTENT, FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
    />
))

add('With small-text alert', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [FIXTURE_SMALL_TEXT_MARKDOWN_ALERT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With one-line alert', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.PlainText,
                        value: 'This is a test alert.',
                    },
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With alert with warning icon', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.PlainText,
                        value: 'This is a warning alert.',
                    },
                    iconKind: 'warning',
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With dismissible alert with icon', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [
                {
                    summary: {
                        kind: MarkupKind.Markdown,
                        value: 'This is a test alert.',
                    },
                    type: 'test-alert-type',
                    iconKind: 'info',
                },
            ],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('With long markdown text dismissible alert with icon.', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT],
            alerts: [FIXTURE_WARNING_MARKDOWN_ALERT],
            aggregatedBadges: [FIXTURE_PARITAL_BADGE, FIXTURE_SEMANTIC_BADGE],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))

add('Multiple MarkupContents with badges and alerts', () => (
    <WebHoverOverlay
        {...commonProps()}
        hoverOrError={{
            contents: [FIXTURE_CONTENT, FIXTURE_CONTENT, FIXTURE_CONTENT],
            aggregatedBadges: [FIXTURE_SEMANTIC_BADGE],
            alerts: [FIXTURE_SMALL_TEXT_MARKDOWN_ALERT, FIXTURE_WARNING_MARKDOWN_ALERT],
        }}
        actionsOrError={FIXTURE_ACTIONS}
        onAlertDismissed={action('onAlertDismissed')}
    />
))
