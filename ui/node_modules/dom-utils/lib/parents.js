/**
 * Returns an array of a DOM element's parent elements.
 * @param {Element} element The DOM element whose parents to get.
 * @return {Array} An array of all parent elemets, or an empty array if no
 *     parent elements are found.
 */
module.exports = function parents(element) {
  var list = [];
  while (element && element.parentNode && element.parentNode.nodeType == 1) {
    list.push(element = element.parentNode);
  }
  return list;
};
