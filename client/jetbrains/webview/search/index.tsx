import React from 'react'

import { render } from 'react-dom'

const node = document.querySelector('#main') as HTMLDivElement

render(<p>Hello from React!</p>, node)
