import React, { Component, PropTypes } from 'react';
import ReactDom from 'react-dom';
import raf, { caf } from 'raf';
import { translateStyle } from 'react-smooth';

const g = 9.8;

function Circle(props) {
  const { r, currTheta, ropeLength } = props;
  const cx = (ropeLength - r) * Math.sin(currTheta) + ropeLength - r;
  const cy = (ropeLength - r) * Math.cos(currTheta) - r;
  const translate = `translate(${cx}px, ${cy}px)`;

  const style = {
    width: 2 * r,
    height: 2 * r,
    borderRadius: r,
    transform: translate,
    WebkitTransform: translate,
    background: `radial-gradient(circle at ${r * 2 / 3}px ${r * 2 / 3}px,#5cabff,#000)`,
    position: 'absolute',
    top: 0,
    left: 0,
  };

  return (
    <div className="circle-ball"
      style={translateStyle(style)}
      { ...props }
    />
  );
}

Circle.prototype.propTypes = {
  r: PropTypes.number,
  currTheta: PropTypes.number,
  ropeLength: PropTypes.number,
};

function Line(props) {
  const { ropeLength, currTheta } = props;
  const x1 = ropeLength;
  const x2 = x1;
  const y2 = ropeLength;

  return (
    <svg
      width={ropeLength * 2}
      height={ropeLength * 2}
      viewPort={`0 0 ${ropeLength * 2} ${ropeLength * 2}`}
      version="1.1"
      xmlns="http://www.w3.org/2000/svg"
    >
      <line
        x1={x1}
        y1="0"
        x2={x2}
        y2={y2}
        stroke="black"
        strokeWidth="2"
        style={translateStyle({
          transform: `rotate(${-currTheta / Math.PI * 180}deg)`,
          transformOrigin: 'top',
        })}
      />
    </svg>
  );
}

Line.prototype.propTypes = {
  ropeLength: PropTypes.number,
  currTheta: PropTypes.number,
};

class Pendulum extends Component {
  static propTypes = {
    ropeLength: PropTypes.number,
    theta: PropTypes.number,
    radius: PropTypes.number,
  };

  state = {
    currTheta: this.props.theta,
  };

  componentDidMount() {
    this.cafId = raf(this.update.bind(this));
  }

  componentWillUnmount() {
    if (this.cafId) {
      caf(this.cafId);
    }
  }

  update(now) {
    if (!this.initialTime) {
      this.initialTime = now;

      this.cafId = raf(this.update.bind(this));
    }

    const { ropeLength, theta } = this.props;
    const { currTheta } = this.state;
    const A = theta;
    const omiga = Math.sqrt(g / ropeLength * 200);

    this.setState({
      currTheta: theta * Math.cos(omiga * (now - this.initialTime) / 1000),
    });

    this.cafId = raf(this.update.bind(this));
  }

  render() {
    const { currTheta } = this.state;
    const { ropeLength, radius } = this.props;

    return (
      <div className="pendulum" style={{ position: 'relative' }}>
        <Line ropeLength={ropeLength} currTheta={currTheta} />
        <Circle r={radius} currTheta={currTheta} ropeLength={ropeLength} />
      </div>
    );
  }
}

class App extends Component {
  state = {
    ropeLength: 300,
    theta: 18,
  };

  handleThetaChange(_theta) {
    this.setState({
      theta: _theta,
    });
  }

  handleRopeChange(ropeLength) {
    this.setState({ ropeLength });
  }

  render() {
    const { theta, ropeLength } = this.state;

    const thetaValueLink = {
      value: theta,
      requestChange: this.handleThetaChange.bind(this),
    };
    const ropeValueLink = {
      value: ropeLength,
      requestChange: this.handleRopeChange.bind(this),
    };

    return (
      <div className="pendulum-app">
        theta: <input type="number" valueLink={thetaValueLink} placeholder="0 ~ 90" />
        <br />
        rope length: <input type="number" valueLink={ropeValueLink} placeholder="rope length" />
        <Pendulum
          theta={ theta && (parseInt(theta, 10) / 180 * Math.PI) || 0.3 }
          ropeLength={ ropeLength && parseInt(ropeLength, 10) || 300 }
          radius={30}
        />
      </div>
    );
  }
}

ReactDom.render(<App />, document.getElementById('app'));
