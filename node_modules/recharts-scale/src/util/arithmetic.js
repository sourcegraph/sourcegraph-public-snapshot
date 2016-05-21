/**
 * @fileOverview 一些公用的运算方法
 * @author xile611
 * @date 2015-09-17
 */
import { curry } from './utils';

/**
 * 判断数据是否为浮点类型
 *
 * @param {Number} num 输入值
 * @return {Boolean} 是否是浮点类型
 */
function isFloat(num) {
  return /^([+-]?)\d*\.\d+$/.test(num);
}

/**
 * 获取数值的位数
 * 其中绝对值属于区间[0.1, 1)， 得到的值为0
 * 绝对值属于区间[0.01, 0.1)，得到的位数为 -1
 * 绝对值属于区间[0.001, 0.01)，得到的位数为 -2
 *
 * @param  {Number} value 数值
 * @return {Integer} 位数
 */
function getDigitCount(value) {
  const abs = Math.abs(value);
  let result;

  if (value === 0) {
    result = 1;
  } else if (abs < 1) {
    result = Math.floor(Math.log(abs) / Math.log(10)) + 1;
  } else {
    const str = '' + value;
    const ary = str.split('.');

    result = ary[0].length;
  }

  return result;
}
/**
 * 计算数值的小数点后的位数
 * @param  {Number} a 数值，可能为整数，也可能为浮点数
 * @return {Integer}   位数
 */
function getDecimalDigitCount(a) {
  const str = a ? '' + a : '';
  const ary = str.split('.');

  return ary.length > 1 ? ary[1].length : 0;
}
/**
 * 加法运算，解决了js运算的精度问题
 * @param  {Number} a 被加数
 * @param  {Number} b 加数
 * @return {Number}   和
 */
function sum(a, b) {
  let count = Math.max(getDecimalDigitCount(a), getDecimalDigitCount(b));

  count = Math.pow(10, count);

  return (multiply(a, count) + multiply(b, count)) / count;
}
/**
 * 减法运算，解决了js运算的精度问题
 * @param  {Number} a 被减数
 * @param  {Number} b 减数
 * @return {Number}   差
 */
function minus(a, b) {
  return sum(a, -b);
}
/**
 * 乘法运算，解决了js运算的精度问题
 * @param  {Number} a 被乘数
 * @param  {Number} b 乘数
 * @return {Number}   积
 */
function multiply(a, b) {
  const intA = parseInt(('' + a).replace('.', ''), 10);
  const intB = parseInt(('' + b).replace('.', ''), 10);
  const count = getDecimalDigitCount(a) + getDecimalDigitCount(b);

  return (intA * intB) / Math.pow(10, count);
}
/**
 * 除法运算，解决了js运算的精度问题
 * @param  {Number} a 被除数
 * @param  {Number} b 除数
 * @return {Number}   结果
 */
function divide(a, b) {
  const ca = getDecimalDigitCount(a);
  const cb = getDecimalDigitCount(b);
  const intA = parseInt(('' + a).replace('.', ''), 10);
  const intB = parseInt(('' + b).replace('.', ''), 10);

  return (intA / intB) * Math.pow(10, cb - ca);
}

/**
 * 按照固定的步长获取[start, end)这个区间的数据
 * 并且需要处理js计算精度的问题
 *
 * @param  {Number} start 起点
 * @param  {Number} end   终点，不包含该值
 * @param  {Number} step  步长
 * @return {Array}        若干数值
 */
function rangeStep(start, end, step) {
  let num = start;
  const result = [];

  while (num < end) {
    result.push(num);

    num = sum(num, step);
  }

  return result;
}
/**
 * 对数值进行线性插值
 *
 * @param  {Number} a  定义域的极点
 * @param  {Number} b  定义域的极点
 * @param  {Number} t  [0, 1]内的某个值
 * @return {Number}    定义域内的某个值
 */
const interpolateNumber = curry((a, b, t) => {
  const newA = +a;
  const newB = +b;

  return newA + t * (newB - newA);
});
/**
 * 线性插值的逆运算
 *
 * @param  {Number} a 定义域的极点
 * @param  {Number} b 定义域的极点
 * @param  {Number} x 可以认为是插值后的一个输出值
 * @return {Number}   当x在 a ~ b这个范围内时，返回值属于[0, 1]
 */
const uninterpolateNumber = curry((a, b, x) => {
  let diff = b - (+a);

  diff = diff ? diff : Infinity;

  return (x - a) / diff;
});
/**
 * 线性插值的逆运算，并且有截断的操作
 *
 * @param  {Number} a 定义域的极点
 * @param  {Number} b 定义域的极点
 * @param  {Number} x 可以认为是插值后的一个输出值
 * @return {Number}   当x在 a ~ b这个区间内时，返回值属于[0, 1]，
 * 当x不在 a ~ b这个区间时，会截断到 a ~ b 这个区间
 */
const uninterpolateTruncation = curry((a, b, x) => {
  let diff = b - (+a);

  diff = diff ? diff : Infinity;

  return Math.max(0, Math.min(1, (x - a) / diff));
});

export default {
  rangeStep,
  isFloat,
  getDigitCount,
  getDecimalDigitCount,

  sum,
  minus,
  multiply,
  divide,

  interpolateNumber,
  uninterpolateNumber,
  uninterpolateTruncation,
};
