'use strict';

module.exports = function(context) {
  function markTypeAsUsed(node) {
    context.markVariableAsUsed(node.id.name);
  }
  return {
    DeclareClass: markTypeAsUsed,
    DeclareFunction: markTypeAsUsed,
    DeclareModule: markTypeAsUsed,
    DeclareVariable: markTypeAsUsed,
    GenericTypeAnnotation: function(node) {
      var typeId;
      if (node.id.type === 'Identifier') {
        typeId = node.id;
      } else if (node.id.type === 'QualifiedTypeIdentifier') {
        typeId = node.id;
        do { typeId = typeId.qualification; } while (typeId.qualification);
      }

      for (var scope = context.getScope(); scope; scope = scope.upper) {
        var variable = scope.set.get(typeId.name);
        if (variable && variable.defs.length) {
          context.markVariableAsUsed(typeId.name);
          break;
        }
      }
    }
  };
};

module.exports.schema = [];
