import React, { Component, Children, PropTypes } from 'react';
import Animate from './Animate';

class AnimateGroupChild extends Component {
  static propTypes = {
    appear: PropTypes.object,
    leave: PropTypes.object,
    enter: PropTypes.object,
    children: PropTypes.element,
  };

  state = {
    isActive: false,
  };

  handleStyleActive(style, done) {
    if (style) {
      const onAnimationEnd = style.onAnimationEnd ?
        () => {
          style.onAnimationEnd();
          done();
        } :
        done;

      this.setState({
        ...style,
        onAnimationEnd,
        isActive: true,
      });
    } else {
      done();
    }
  }

  componentWillAppear(done) {
    this.handleStyleActive(this.props.appear, done);
  }

  componentWillEnter(done) {
    this.handleStyleActive(this.props.enter, done);
  }

  componentWillLeave(done) {
    this.handleStyleActive(this.props.leave, done);
  }

  render() {
    return (
      <Animate { ...this.state }>
        {Children.only(this.props.children)}
      </Animate>
    );
  }
}

export default AnimateGroupChild;
