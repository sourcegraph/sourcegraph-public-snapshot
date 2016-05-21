import raf from 'raf';

export default function setRafTimeout(callback, timeout = 0) {
  let currTime = -1;

  const shouldUpdate = now => {
    if (currTime < 0) {
      currTime = now;
    }

    if (now - currTime > timeout) {
      callback(now);
      currTime = -1;
    } else {
      raf(shouldUpdate);
    }
  };

  raf(shouldUpdate);
}
