import '../platform/polyfills'

import * as Comlink from 'comlink'
import React from 'react'
import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { ExtensionCoreAPI, HelpSidebarAPI } from '../../contract'
import { createEndpointsForWebToNode } from '../comlink/webviewEndpoint'

import { HelpSidebarView } from './HelpSidebarView'

const vsCodeApi = window.acquireVsCodeApi()

const { proxy, expose } = createEndpointsForWebToNode(vsCodeApi)

export const extensionCoreAPI: Comlink.Remote<ExtensionCoreAPI> = Comlink.wrap(proxy)

const helpSidebarAPI: HelpSidebarAPI = {}

Comlink.expose(helpSidebarAPI, expose)

setLinkComponent(AnchorLink)

const Main: React.FC = () => <HelpSidebarView />

render(<Main />, document.querySelector('#root'))
