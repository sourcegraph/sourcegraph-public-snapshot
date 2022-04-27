import { render } from 'react-dom'

import { App } from './App'

import 'focus-visible'

const node = document.querySelector('#main') as HTMLDivElement
render(<App />, node)
