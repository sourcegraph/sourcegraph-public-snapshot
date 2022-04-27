import { storiesOf } from '@storybook/react'

import { Link } from '@sourcegraph/wildcard'

import { WebStory } from '../WebStory'

import { DismissibleAlert } from './DismissibleAlert'

const { add } = storiesOf('web/DismissibleAlert', module).addDecorator(story => <WebStory>{() => story()}</WebStory>)

add('One-line alert', () => (
    <DismissibleAlert variant="info" partialStorageKey="dismissible-alert-one-line">
        <span>
            1 bulk operation has recently failed running. Click the <Link to="?">bulk operations tab</Link> to view.
        </span>
    </DismissibleAlert>
))

add('Multiline alert', () => (
    <DismissibleAlert variant="info" partialStorageKey="dismissible-alert-multiline">
        WebAssembly (sometimes abbreviated Wasm) is an open standard that defines a portable binary-code format for
        executable programs, and a corresponding textual assembly language, as well as interfaces for facilitating
        interactions between such programs and their host environment. The main goal of WebAssembly is to enable
        high-performance applications on web pages, but the format is designed to be executed and integrated in other
        environments as well, including standalone ones.
    </DismissibleAlert>
))
