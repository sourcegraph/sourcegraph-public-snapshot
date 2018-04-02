import * as assert from 'assert'
import * as enzyme from 'enzyme'
import Adapter from 'enzyme-adapter-react-16'
import { createMemoryHistory } from 'history'
import * as React from 'react'
import { Tree2 } from './Tree2'

enzyme.configure({ adapter: new Adapter() })

// TODO!(sqs)
describe('Tree2', () => {
    it.skip('renders correctly', () => {
        const history = createMemoryHistory()
        const hello = enzyme.shallow(
            <Tree2
                repoPath="r"
                history={history}
                paths={['a/1', 'a/2', 'a/b/1']}
                activePath="a/1"
                activePathIsDir={false}
            />
        )
        assert.strictEqual(hello.text(), 'TODO')
    })
})
