import assert from 'assert'
import Enzyme, { shallow } from 'enzyme'
import Adapter from 'enzyme-adapter-react-16'
import H from 'history'
import React from 'react'
import { Redirect } from 'react-router'
// import sinon from 'sinon'
import { ClientAuthorizationFlow } from './ClientAuthorizationFlow'

Enzyme.configure({ adapter: new Adapter() })

describe('<ClientAuthorizationFlow />', () => {
    it('renders three <Foo /> components', () => {
        const wrapper = shallow(<ClientAuthorizationFlow location={H.createLocation('/')} authenticatedUser={null} />)
        assert.strictEqual(wrapper.find(Redirect).length, 1)
    })

    /*   it('renders an `.icon-star`', () => {
    const wrapper = shallow(<ClientAuthorizationFlow />);
    expect(wrapper.find('.icon-star')).to.have.lengthOf(1);
  });

  it('renders children when passed in', () => {
    const wrapper = shallow((
      <ClientAuthorizationFlow>
        <div className="unique" />
      </ClientAuthorizationFlow>
    ));
    expect(wrapper.contains(<div className="unique" />)).to.equal(true);
  });

  it('simulates click events', () => {
    const onButtonClick = sinon.spy();
    const wrapper = shallow(<Foo onButtonClick={onButtonClick} />);
    wrapper.find('button').simulate('click');
    expect(onButtonClick).to.have.property('callCount', 1);
  });
 */
})
