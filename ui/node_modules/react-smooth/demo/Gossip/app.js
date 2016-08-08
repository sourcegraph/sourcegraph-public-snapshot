import Animate from 'react-smooth';
import React, { Component } from 'react';
import ReactDom from 'react-dom';

const getSTEPS = onAnimationEnd => [{
  duration: 1000,
  style: {
    opacity: 0,
  },
}, {
  duration: 1000,
  style: {
    opacity: 1,
    transformOrigin: '110px 110px',
    transform: 'rotate(0deg) translate(0px, 0px)',
  },
  easing: 'ease-in',
}, {
  duration: 1000,
  style: {
    transform: 'rotate(500deg) translate(0px, 0px)',
  },
  easing: 'ease-in-out',
}, {
  duration: 2000,
  style: {
    transformOrigin: '610px 610px',
    transform: 'rotate(1440deg) translate(500px, 500px)',
  },
}, {
  duration: 50,
  style: {
    transformOrigin: 'center center',
    transform: 'translate(500px, 500px) scale(1)',
  },
  onAnimationEnd,
}, {
  duration: 1000,
  style: {
    transformOrigin: 'center center',
    transform: 'translate(500px, 500px) scale(1.6)',
  },
}];

const createPoint = (x, y) => {
  const currX = x;
  const currY = y;

  return {
    getPath: cmd => [cmd, currX, currY].join(' '),
    getCircle: (props) => <circle cx={currX} cy={currY} { ...props } />,
    x: currX,
    y: currY,
  };
};

const getArcPath = (radius, rotation, isLarge, isSweep, dx, dy) => {
  return ['A', radius, radius, rotation, isLarge, isSweep, dx, dy].join(' ');
};

class Gossip extends Component {
  static displayName = 'Gossip';

  constructor(props, ctx) {
    super(props, ctx);

    this.state = { canBegin: false };
    this.handleTextAniamtionBegin = this.handleTextAniamtionBegin.bind(this);
    this.STEPS = getSTEPS(this.handleTextAniamtionBegin);
  }

  handleTextAniamtionBegin() {
    this.setState({
      canBegin: true,
    });
  }

  renderPath() {
    const cx = 110;
    const cy = 110;
    const r = 100;
    const sr = r / 2;

    const beginPoint = createPoint(cx, cy - r);
    const endPoint = createPoint(cx, cy + r);
    const move = beginPoint.getPath('M');
    const A = getArcPath(sr, 0, 0, 0, cx, cy);
    const A2 = getArcPath(sr, 0, 0, 1, endPoint.x, endPoint.y);
    const A3 = getArcPath(r, 0, 0, 1, beginPoint.x, beginPoint.y);

    return <path d={[move, A, A2, A3].join('\n')} />;
  }

  renderSmallCircles() {
    const cx = 110;
    const cy = 110;
    const r = 100;
    const sr = r / 2;
    const tr = 5;

    const centers = [createPoint(cx, cy - sr), createPoint(cx, cy + sr)];
    const circles = centers.map((p, i) =>
      p.getCircle({
        r: tr,
        fill: i ? 'white' : 'black',
        key: i,
      })
    );

    return <g className="small-circles">{circles}</g>;
  }

  renderText() {
    return (
      <Animate canBegin={this.state.canBegin}
        duration={1000}
        from={{ opacity: 0, transform: 'scale(1)' }}
        to={{ opacity: 1, transform: 'scale(1.5)' }}
      >
        <g style={{ transformOrigin: 'center center' }}>
        <text x="500" y="300">May you no bug this year</text>
        </g>
      </Animate>
    );
  }

  render() {
    return (
      <svg width="1000" height="1000">
        <Animate steps={this.STEPS} >
          <g className="gossip">
            <circle cx="110" cy="110" r="100" style={{ stroke: 'black', fill: 'white' }} />
            {this.renderPath()}
            {this.renderSmallCircles()}
          </g>
        </Animate>
        {this.renderText()}
      </svg>
    );
  }
}

ReactDom.render(<Gossip />, document.getElementById('app'));
