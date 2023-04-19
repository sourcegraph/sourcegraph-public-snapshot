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
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default meta

export const Simple: ComponentStoryObj<typeof App> = {
    render: () => <App vscodeAPI={dummyVSCodeAPI} />,
}

const dummyVSCodeAPI: VSCodeWrapper = {
    onMessage: cb => {
        // Send initial message so that the component is fully rendered.
        cb({ type: 'config', config: { debug: true, hasAccessToken: true, serverEndpoint: 'https://example.com' } })
        return () => {}
    },
    postMessage: () => {},
}
