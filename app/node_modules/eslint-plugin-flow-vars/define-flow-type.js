'use strict';

module.exports = function(context) {
  var globalScope;

  // do nearly the same thing that eslint does for config globals
  // https://github.com/eslint/eslint/blob/v2.0.0/lib/eslint.js#L118-L194
  function makeDefined(ident) {
    for (var i = 0; i < globalScope.through.length; i++) {
      var ref = globalScope.through[i];
      if (ref.identifier === ident) {
        // use "__defineGeneric" since we don't have a reference to "escope.Variable"
        globalScope.__defineGeneric(
          ident.name,
          globalScope.set,
          globalScope.variables
        );
        var variable = globalScope.set.get(ident.name);
        variable.writeable = false;
        // "through" contains all references whose definition cannot be found
        // so we need to update references and remove the ones that were added
        globalScope.through.splice(i, 1);
        ref.resolved = variable;
        variable.references.push(ref);
        break;
      }
    }
  }

  return {
    Program: function(node) {
      globalScope = context.getScope();
    },
    GenericTypeAnnotation: function(node) {
      if (node.id.type === 'Identifier') {
        makeDefined(node.id);
      } else if (node.id.type === 'QualifiedTypeIdentifier') {
        var qid = node.id;
        do { qid = qid.qualification; } while (qid.qualification);
        makeDefined(qid);
      }
    },
    ClassImplements: function(node) {
      makeDefined(node.id);
    },
    InterfaceDeclaration: function(node) {
      makeDefined(node.id);
    }
  };
};

module.exports.schema = [];
