const path = require('path')

const config = {
  files: [
    {
      path: path.join(__dirname, './assets/scripts/*.br'),
      maxSize: '10kb',
    },
    {
      path: path.join(__dirname, './assets/styles/*.br'),
      maxSize: '10kb',
    },
  ],
}

module.exports = config
