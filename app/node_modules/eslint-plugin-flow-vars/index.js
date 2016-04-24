'use strict';

module.exports = {
  rules: {
    'define-flow-type': require('./define-flow-type'),
    'use-flow-type': require('./use-flow-type')
  },
  rulesConfig: {
    'define-flow-type': 0,
    'use-flow-type': 0
  }
};
