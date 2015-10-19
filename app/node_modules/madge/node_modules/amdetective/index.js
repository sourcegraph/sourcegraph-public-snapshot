var esprima = require('esprima'),
    parse = require('./lib/parse.js');

function find(fileContents, options) {
    options = options || {};

    //Set up source input
    var moduleDeps = [],
        moduleList = [],
        astRoot = esprima.parse(fileContents);

    parse.recurse(astRoot, function (callName, config, name, deps) {
        if (!deps) {
            deps = [];
        }

        if (!name) {
            //If there is no module name, the dependencies are for
            //this file/default module name.
            moduleDeps = moduleDeps.concat(deps);
        } else {
            moduleList.push({
                name: name,
                deps: deps
            });
        }

        //If define was found, no need to dive deeper, unless
        //the config explicitly wants to dig deeper.
        return !!options.findNestedDependencies;
    }, options);

    return {
      moduleDeps: moduleDeps,
      moduleList: moduleList
    };
}

function findSimple(fileContents, options) {
  var result = find(fileContents, options);
  return result.moduleDeps.concat(result.moduleList);
}

module.exports = findSimple;
module.exports.find = find;
