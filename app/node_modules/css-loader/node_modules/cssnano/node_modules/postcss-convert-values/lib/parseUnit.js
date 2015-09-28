'use strict';

// From longer to shorter!
var units = [
    'vmin',
    'vmax',
    'rem',
    'em',
    'ex',
    'vw',
    'vh',
    'vm',
    'ch',
    'in',
    'cm',
    'mm',
    'pt',
    'pc',
    'px',
    'ms',
    's',
    '%'
];

module.exports = function (value) {
    var max, index;

    for (max = 4; max !== 0; max -= 1) {
        index = units.indexOf(value.slice(-max));
        if (~index) {
            return units[index];
        }
    }

    return null;
};
