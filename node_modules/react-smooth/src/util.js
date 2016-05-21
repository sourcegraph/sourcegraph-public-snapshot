import { intersection } from 'lodash';

const PREFIX_LIST = ['Webkit', 'Moz', 'O', 'ms'];
const IN_LINE_PREFIX_LIST = ['-webkit-', '-moz-', '-o-', '-ms-'];
const IN_COMPATIBLE_PROPERTY = ['transform', 'transformOrigin', 'transition'];

export const getIntersectionKeys = (preObj, nextObj) =>
  intersection(Object.keys(preObj), Object.keys(nextObj));

export const identity = param => param;

/*
 * @description: convert camel case to dash case
 * string => string
 */
export const getDashCase = name => name.replace(/([A-Z])/g, v => `-${v.toLowerCase()}`);

/*
 * @description: add compatible style prefix
 * (string, string) => object
 */
export const generatePrefixStyle = (name, value) => {
  if (IN_COMPATIBLE_PROPERTY.indexOf(name) === -1) {
    return { [name]: value };
  }

  const isTransition = name === 'transition';
  const camelName = name.replace(/(\w)/, v => v.toUpperCase());
  let styleVal = value;

  return PREFIX_LIST.reduce((result, property, i) => {
    if (isTransition) {
      styleVal = value.replace(/(transform|transform-origin)/gim, '-webkit-$1');
    }

    return {
      ...result,
      [property + camelName]: styleVal,
    };
  }, {});
};

export const log = console.log.bind(console);

/*
 * @description: log the value of a varible
 * string => any => any
 */
export const debug = name => item => {
  log(name, item);

  return item;
};

/*
 * @description: log name, args, return value of a function
 * function => function
 */
export const debugf = (tag, f) => (...args) => {
  const res = f(...args);
  const name = tag || f.name || 'anonymous function';
  const argNames = `(${args.map(JSON.stringify).join(', ')})`;

  log(`${name}: ${argNames} => ${JSON.stringify(res)}`);

  return res;
};

/*
 * @description: map object on every element in this object.
 * (function, object) => object
 */
export const mapObject = (fn, obj) =>
  Object.keys(obj).reduce((res, key) => ({
    ...res,
    [key]: fn(key, obj[key]),
  }), {});

/*
 * @description: add compatible prefix to style
 * object => object
 */
export const translateStyle = style =>
  Object.keys(style).reduce((res, key) => ({
    ...res,
    ...generatePrefixStyle(key, res[key]),
  }), style);

export const compose = (...args) => {
  if (!args.length) {
    return identity;
  }

  const fns = args.reverse();
  // first function can receive multiply arguments
  const firstFn = fns[0];
  const tailsFn = fns.slice(1);

  return (...composeArgs) =>
    tailsFn.reduce((res, fn) =>
      fn(res),
      firstFn(...composeArgs)
  );
};

export const getTransitionVal = (props, duration, easing) =>
  props.map(prop =>
    `${getDashCase(prop)} ${duration}ms ${easing}`)
    .join(',');

const __DEV__ = process.env.NODE_ENV !== 'production';

export const warn = (condition, format, a, b, c, d, e, f) => {
  if (__DEV__ && typeof console !== 'undefined' && console.warn) {
    if (format === undefined) {
      console.warn('LogUtils requires an error message argument');
    }

    if (!condition) {
      if (format === undefined) {
        console.warn(
          'Minified exception occurred; use the non-minified dev environment ' +
          'for the full error message and additional helpful warnings.'
        );
      } else {
        const args = [a, b, c, d, e, f];
        let argIndex = 0;

        console.warn(format.replace(/%s/g, () => args[argIndex++]));
      }
    }
  }
};
