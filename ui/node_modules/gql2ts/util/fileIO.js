'use strict';
const fs = require('fs');

module.exports = {
  readFile: fileName => JSON.parse(fs.readFileSync(fileName).toString()),
  writeToFile: (fileName, data) => fs.writeFile(fileName, data, err => err && console.error('Failed to write', err) )
}
