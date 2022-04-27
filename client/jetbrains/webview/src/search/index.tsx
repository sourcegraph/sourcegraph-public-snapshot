import { render } from 'react-dom'

import { App } from './App'

const node = document.querySelector('#main') as HTMLDivElement
render(<App />, node)
