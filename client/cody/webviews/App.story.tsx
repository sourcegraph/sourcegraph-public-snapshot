import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { App } from './App'
import { VSCodeStoryDecorator } from './storybook/VSCodeStoryDecorator'
import { VSCodeWrapper } from './utils/VSCodeApi'

const meta: ComponentMeta<typeof App> = {
    title: 'cody/App',
    component: App,

    decorators: [VSCodeStoryDecorator],

    parameters: {
        component: App,
        chromatic: {
            disableSnapshot: false,
        },
    },
}

// When this Story is rendered on Chromatic, the below DOM API will error. It
// doesn't seem necessary for this UI at all and is likely coming from the VS
// Code component library so we just stub it out for now.
// eslint-disable-next-line @typescript-eslint/unbound-method
const originalSetValidity = ElementInternals.prototype.setValidity
ElementInternals.prototype.setValidity = function () {
    return originalSetValidity.call(this, {})
}

export default meta

export const Simple: ComponentStoryObj<typeof App> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <App vscodeAPI={dummyVSCodeAPI} />
        </div>
    ),
}

const dummyVSCodeAPI: VSCodeWrapper = {
    onMessage: cb => {
        // Send initial message so that the component is fully rendered.
        cb({ type: 'config', config: { debug: true, hasAccessToken: true, serverEndpoint: 'https://example.com' } })
        return () => {}
    },
    postMessage: () => {},
}
