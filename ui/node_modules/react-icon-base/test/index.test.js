import { default as React } from 'react'
import { default as TestUtils } from 'react-addons-test-utils'
import { default as expect } from 'expect'
import { default as IconBase } from '..'

const renderer = TestUtils.createRenderer()

describe('IconBase', () => {
  let outer

  beforeEach(() => {
    renderer.render(<IconBase />)
    outer = renderer.getRenderOutput()
  })

  it('renders svg', () => {
    expect(outer.type).toEqual('svg')
  })

  it('has default props', () => {
    expect(outer.props.fill).toEqual('currentColor')
    expect(outer.props.preserveAspectRatio).toEqual('xMidYMid meet')
    expect(outer.props.height).toEqual('1em')
    expect(outer.props.width).toEqual('1em')
    expect(outer.props.style.verticalAlign).toEqual('middle')
  })

  it('has does not have a default color', () => {
    expect(outer.props.style.color).toNotExist()
  })
})
