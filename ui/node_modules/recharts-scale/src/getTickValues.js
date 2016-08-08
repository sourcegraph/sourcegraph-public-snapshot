import { compose, map, range } from './util/utils';
import getNiceTickValues from './getNiceTickValues';

function getTickValues(domain, tickCount) {
  if (domain[0] === domain[1]) {
    return getNiceTickValues(domain, tickCount);
  }

  tickCount = Math.max(tickCount, 2);

  const step = (domain[1] - domain[0]) / (tickCount - 1);

  const fn = compose(
    map(n => (domain[0] + n * step)),
    range
  );

  return fn(0, tickCount);
}

export default getTickValues;
