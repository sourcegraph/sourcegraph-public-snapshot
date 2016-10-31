// file I/O helpers
'use strict';
const fileIO = require('./fileIO');

const generateNamespace = (namespaceName, interfaces) => `// graphql typescript definitions

declare namespace ${namespaceName} {
${interfaces}
}
`;

const writeNamespaceToFile = (outputFile, namespace) => fileIO.writeToFile(outputFile, namespace);

module.exports = {
  writeNamespaceToFile,
  generateNamespace
}
