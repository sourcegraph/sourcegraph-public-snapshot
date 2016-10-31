#!/usr/bin/env node
'use strict';
const program = require('commander');

// file I/O helpers
const fileIO = require('./util/fileIO');

// Interface Utils
const interfaceUtils = require('./util/interface');

// Namespace Utils
const namespaceUtils = require('./util/namespace')

program
  .version('0.3.1')
  .usage('[options] <schema.json>')
  .option('-o --output-file [outputFile]', 'name for ouput file, defaults to graphqlInterfaces.d.ts', 'graphqlInterfaces.d.ts')
  .option('-n --namespace [namespace]', 'name for the namespace, defaults to "GQL"', 'GQL')
  .option('-i --ignored-types <ignoredTypes>', 'names of types to ignore (comma delimited)', v => v.split(','), [])
  .action((fileName, options) => {
    let schema = fileIO.readFile(fileName);

    let interfaces = interfaceUtils.schemaToInterfaces(schema, options);

    let namespace = namespaceUtils.generateNamespace(options.namespace, interfaces);

    namespaceUtils.writeNamespaceToFile(options.outputFile, namespace);
  })
  .parse(process.argv);

if (!process.argv.slice(2).length) {
  program.outputHelp();
}
