"use strict";

exports.__esModule = true;

var _symbol = require("babel-runtime/core-js/symbol");

var _symbol2 = _interopRequireDefault(_symbol);

var _create = require("babel-runtime/core-js/object/create");

var _create2 = _interopRequireDefault(_create);

var _classCallCheck2 = require("babel-runtime/helpers/classCallCheck");

var _classCallCheck3 = _interopRequireDefault(_classCallCheck2);

exports.default = function () {
  return {
    visitor: {
      VariableDeclaration: function VariableDeclaration(path, file) {
        var node = path.node;
        var parent = path.parent;
        var scope = path.scope;

        if (!isBlockScoped(node)) return;
        convertBlockScopedToVar(path, null, parent, scope, true);

        if (node._tdzThis) {
          var nodes = [node];

          for (var i = 0; i < node.declarations.length; i++) {
            var decl = node.declarations[i];
            if (decl.init) {
              var assign = t.assignmentExpression("=", decl.id, decl.init);
              assign._ignoreBlockScopingTDZ = true;
              nodes.push(t.expressionStatement(assign));
            }
            decl.init = file.addHelper("temporalUndefined");
          }

          node._blockHoist = 2;

          if (path.isCompletionRecord()) {
            // ensure we don't break completion record semantics by returning
            // the initialiser of the last declarator
            nodes.push(t.expressionStatement(scope.buildUndefinedNode()));
          }

          path.replaceWithMultiple(nodes);
        }
      },
      Loop: function Loop(path, file) {
        var node = path.node;
        var parent = path.parent;
        var scope = path.scope;

        t.ensureBlock(node);
        var blockScoping = new BlockScoping(path, path.get("body"), parent, scope, file);
        var replace = blockScoping.run();
        if (replace) path.replaceWith(replace);
      },
      "BlockStatement|SwitchStatement|Program": function BlockStatementSwitchStatementProgram(path, file) {
        if (!t.isLoop(path.parent)) {
          var blockScoping = new BlockScoping(null, path, path.parent, path.scope, file);
          blockScoping.run();
        }
      }
    }
  };
};

var _babelTraverse = require("babel-traverse");

var _babelTraverse2 = _interopRequireDefault(_babelTraverse);

var _tdz = require("./tdz");

var _babelTypes = require("babel-types");

var t = _interopRequireWildcard(_babelTypes);

var _values = require("lodash/values");

var _values2 = _interopRequireDefault(_values);

var _extend = require("lodash/extend");

var _extend2 = _interopRequireDefault(_extend);

var _babelTemplate = require("babel-template");

var _babelTemplate2 = _interopRequireDefault(_babelTemplate);

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } else { var newObj = {}; if (obj != null) { for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) newObj[key] = obj[key]; } } newObj.default = obj; return newObj; } }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

/* eslint max-len: 0 */

var buildRetCheck = (0, _babelTemplate2.default)("\n  if (typeof RETURN === \"object\") return RETURN.v;\n");

function isBlockScoped(node) {
  if (!t.isVariableDeclaration(node)) return false;
  if (node[t.BLOCK_SCOPED_SYMBOL]) return true;
  if (node.kind !== "let" && node.kind !== "const") return false;
  return true;
}

function convertBlockScopedToVar(path, node, parent, scope) {
  var moveBindingsToParent = arguments.length <= 4 || arguments[4] === undefined ? false : arguments[4];

  if (!node) {
    node = path.node;
  }
  // https://github.com/babel/babel/issues/255
  if (!t.isFor(parent)) {
    for (var i = 0; i < node.declarations.length; i++) {
      var declar = node.declarations[i];
      declar.init = declar.init || scope.buildUndefinedNode();
    }
  }

  node[t.BLOCK_SCOPED_SYMBOL] = true;
  node.kind = "var";

  // Move bindings from current block scope to function scope.
  if (moveBindingsToParent) {
    var parentScope = scope.getFunctionParent();
    var ids = path.getBindingIdentifiers();
    for (var name in ids) {
      var binding = scope.getOwnBinding(name);
      if (binding) binding.kind = "var";
      scope.moveBindingTo(name, parentScope);
    }
  }
}

function isVar(node) {
  return t.isVariableDeclaration(node, { kind: "var" }) && !isBlockScoped(node);
}

function replace(path, node, scope, remaps) {
  var remap = remaps[node.name];
  if (!remap) return;

  var ownBinding = scope.getBindingIdentifier(node.name);
  if (ownBinding === remap.binding) {
    scope.rename(node.name, remap.uid);
  } else {
    // scope already has it's own binding that doesn't
    // match the one we have a stored replacement for
    if (path) path.skip();
  }
}

var replaceVisitor = {
  ReferencedIdentifier: function ReferencedIdentifier(path, remaps) {
    replace(path, path.node, path.scope, remaps);
  },
  AssignmentExpression: function AssignmentExpression(path, remaps) {
    var ids = path.getBindingIdentifiers();
    for (var name in ids) {
      replace(null, ids[name], path.scope, remaps);
    }
  }
};

function traverseReplace(node, parent, scope, remaps) {
  if (t.isIdentifier(node)) {
    replace(node, parent, scope, remaps);
  }

  if (t.isAssignmentExpression(node)) {
    var ids = t.getBindingIdentifiers(node);
    for (var name in ids) {
      replace(ids[name], parent, scope, remaps);
    }
  }

  scope.traverse(node, replaceVisitor, remaps);
}

var letReferenceBlockVisitor = _babelTraverse2.default.visitors.merge([{
  Function: function Function(path, state) {
    path.traverse(letReferenceFunctionVisitor, state);
    return path.skip();
  }
}, _tdz.visitor]);

var letReferenceFunctionVisitor = _babelTraverse2.default.visitors.merge([{
  ReferencedIdentifier: function ReferencedIdentifier(path, state) {
    var ref = state.letReferences[path.node.name];

    // not a part of our scope
    if (!ref) return;

    // this scope has a variable with the same name so it couldn't belong
    // to our let scope
    var localBinding = path.scope.getBindingIdentifier(path.node.name);
    if (localBinding && localBinding !== ref) return;

    state.closurify = true;
  }
}, _tdz.visitor]);

var hoistVarDeclarationsVisitor = {
  enter: function enter(path, self) {
    var node = path.node;
    var parent = path.parent;


    if (path.isForStatement()) {
      if (isVar(node.init, node)) {
        var nodes = self.pushDeclar(node.init);
        if (nodes.length === 1) {
          node.init = nodes[0];
        } else {
          node.init = t.sequenceExpression(nodes);
        }
      }
    } else if (path.isFor()) {
      if (isVar(node.left, node)) {
        self.pushDeclar(node.left);
        node.left = node.left.declarations[0].id;
      }
    } else if (isVar(node, parent)) {
      path.replaceWithMultiple(self.pushDeclar(node).map(function (expr) {
        return t.expressionStatement(expr);
      }));
    } else if (path.isFunction()) {
      return path.skip();
    }
  }
};

var loopLabelVisitor = {
  LabeledStatement: function LabeledStatement(_ref, state) {
    var node = _ref.node;

    state.innerLabels.push(node.label.name);
  }
};

var continuationVisitor = {
  enter: function enter(path, state) {
    if (path.isAssignmentExpression() || path.isUpdateExpression()) {
      var bindings = path.getBindingIdentifiers();
      for (var name in bindings) {
        if (state.outsideReferences[name] !== path.scope.getBindingIdentifier(name)) continue;
        state.reassignments[name] = true;
      }
    }
  }
};

function loopNodeTo(node) {
  if (t.isBreakStatement(node)) {
    return "break";
  } else if (t.isContinueStatement(node)) {
    return "continue";
  }
}

var loopVisitor = {
  Loop: function Loop(path, state) {
    var oldIgnoreLabeless = state.ignoreLabeless;
    state.ignoreLabeless = true;
    path.traverse(loopVisitor, state);
    state.ignoreLabeless = oldIgnoreLabeless;
    path.skip();
  },
  Function: function Function(path) {
    path.skip();
  },
  SwitchCase: function SwitchCase(path, state) {
    var oldInSwitchCase = state.inSwitchCase;
    state.inSwitchCase = true;
    path.traverse(loopVisitor, state);
    state.inSwitchCase = oldInSwitchCase;
    path.skip();
  },
  "BreakStatement|ContinueStatement|ReturnStatement": function BreakStatementContinueStatementReturnStatement(path, state) {
    var node = path.node;
    var parent = path.parent;
    var scope = path.scope;

    if (node[this.LOOP_IGNORE]) return;

    var replace = void 0;
    var loopText = loopNodeTo(node);

    if (loopText) {
      if (node.label) {
        // we shouldn't be transforming this because it exists somewhere inside
        if (state.innerLabels.indexOf(node.label.name) >= 0) {
          return;
        }

        loopText = loopText + "|" + node.label.name;
      } else {
        // we shouldn't be transforming these statements because
        // they don't refer to the actual loop we're scopifying
        if (state.ignoreLabeless) return;

        //
        if (state.inSwitchCase) return;

        // break statements mean something different in this context
        if (t.isBreakStatement(node) && t.isSwitchCase(parent)) return;
      }

      state.hasBreakContinue = true;
      state.map[loopText] = node;
      replace = t.stringLiteral(loopText);
    }

    if (path.isReturnStatement()) {
      state.hasReturn = true;
      replace = t.objectExpression([t.objectProperty(t.identifier("v"), node.argument || scope.buildUndefinedNode())]);
    }

    if (replace) {
      replace = t.returnStatement(replace);
      replace[this.LOOP_IGNORE] = true;
      path.skip();
      path.replaceWith(t.inherits(replace, node));
    }
  }
};

var BlockScoping = function () {
  function BlockScoping(loopPath, blockPath, parent, scope, file) {
    (0, _classCallCheck3.default)(this, BlockScoping);

    this.parent = parent;
    this.scope = scope;
    this.file = file;

    this.blockPath = blockPath;
    this.block = blockPath.node;

    this.outsideLetReferences = (0, _create2.default)(null);
    this.hasLetReferences = false;
    this.letReferences = (0, _create2.default)(null);
    this.body = [];

    if (loopPath) {
      this.loopParent = loopPath.parent;
      this.loopLabel = t.isLabeledStatement(this.loopParent) && this.loopParent.label;
      this.loopPath = loopPath;
      this.loop = loopPath.node;
    }
  }

  /**
   * Start the ball rolling.
   */

  BlockScoping.prototype.run = function run() {
    var block = this.block;
    if (block._letDone) return;
    block._letDone = true;

    var needsClosure = this.getLetReferences();

    // this is a block within a `Function/Program` so we can safely leave it be
    if (t.isFunction(this.parent) || t.isProgram(this.block)) {
      this.updateScopeInfo();
      return;
    }

    // we can skip everything
    if (!this.hasLetReferences) return;

    if (needsClosure) {
      this.wrapClosure();
    } else {
      this.remap();
    }

    this.updateScopeInfo();

    if (this.loopLabel && !t.isLabeledStatement(this.loopParent)) {
      return t.labeledStatement(this.loopLabel, this.loop);
    }
  };

  BlockScoping.prototype.updateScopeInfo = function updateScopeInfo() {
    var scope = this.scope;
    var parentScope = scope.getFunctionParent();
    var letRefs = this.letReferences;

    for (var key in letRefs) {
      var ref = letRefs[key];
      var binding = scope.getBinding(ref.name);
      if (!binding) continue;
      if (binding.kind === "let" || binding.kind === "const") {
        binding.kind = "var";
        scope.moveBindingTo(ref.name, parentScope);
      }
    }
  };

  BlockScoping.prototype.remap = function remap() {
    var hasRemaps = false;
    var letRefs = this.letReferences;
    var scope = this.scope;

    // alright, so since we aren't wrapping this block in a closure
    // we have to check if any of our let variables collide with
    // those in upper scopes and then if they do, generate a uid
    // for them and replace all references with it
    var remaps = (0, _create2.default)(null);

    for (var key in letRefs) {
      // just an Identifier node we collected in `getLetReferences`
      // this is the defining identifier of a declaration
      var ref = letRefs[key];

      // todo: could skip this if the colliding binding is in another function
      if (scope.parentHasBinding(key) || scope.hasGlobal(key)) {
        var uid = scope.generateUidIdentifier(ref.name).name;
        ref.name = uid;

        hasRemaps = true;
        remaps[key] = remaps[uid] = {
          binding: ref,
          uid: uid
        };
      }
    }

    if (!hasRemaps) return;

    //

    var loop = this.loop;
    if (loop) {
      traverseReplace(loop.right, loop, scope, remaps);
      traverseReplace(loop.test, loop, scope, remaps);
      traverseReplace(loop.update, loop, scope, remaps);
    }

    this.blockPath.traverse(replaceVisitor, remaps);
  };

  BlockScoping.prototype.wrapClosure = function wrapClosure() {
    var block = this.block;

    var outsideRefs = this.outsideLetReferences;

    // remap loop heads with colliding variables
    if (this.loop) {
      for (var name in outsideRefs) {
        var id = outsideRefs[name];

        if (this.scope.hasGlobal(id.name) || this.scope.parentHasBinding(id.name)) {
          delete outsideRefs[id.name];
          delete this.letReferences[id.name];

          this.scope.rename(id.name);

          this.letReferences[id.name] = id;
          outsideRefs[id.name] = id;
        }
      }
    }

    // if we're inside of a for loop then we search to see if there are any
    // `break`s, `continue`s, `return`s etc
    this.has = this.checkLoop();

    // hoist let references to retain scope
    this.hoistVarDeclarations();

    // turn outsideLetReferences into an array
    var params = (0, _values2.default)(outsideRefs);
    var args = (0, _values2.default)(outsideRefs);

    var isSwitch = this.blockPath.isSwitchStatement();

    // build the closure that we're going to wrap the block with, possible wrapping switch(){}
    var fn = t.functionExpression(null, params, t.blockStatement(isSwitch ? [block] : block.body));
    fn.shadow = true;

    // continuation
    this.addContinuations(fn);

    var ref = fn;

    if (this.loop) {
      ref = this.scope.generateUidIdentifier("loop");
      this.loopPath.insertBefore(t.variableDeclaration("var", [t.variableDeclarator(ref, fn)]));
    }

    // build a call and a unique id that we can assign the return value to
    var call = t.callExpression(ref, args);
    var ret = this.scope.generateUidIdentifier("ret");

    // handle generators
    var hasYield = _babelTraverse2.default.hasType(fn.body, this.scope, "YieldExpression", t.FUNCTION_TYPES);
    if (hasYield) {
      fn.generator = true;
      call = t.yieldExpression(call, true);
    }

    // handlers async functions
    var hasAsync = _babelTraverse2.default.hasType(fn.body, this.scope, "AwaitExpression", t.FUNCTION_TYPES);
    if (hasAsync) {
      fn.async = true;
      call = t.awaitExpression(call);
    }

    this.buildClosure(ret, call);

    // replace the current block body with the one we're going to build
    if (isSwitch) this.blockPath.replaceWithMultiple(this.body);else block.body = this.body;
  };

  /**
   * Push the closure to the body.
   */

  BlockScoping.prototype.buildClosure = function buildClosure(ret, call) {
    var has = this.has;
    if (has.hasReturn || has.hasBreakContinue) {
      this.buildHas(ret, call);
    } else {
      this.body.push(t.expressionStatement(call));
    }
  };

  /**
   * If any of the outer let variables are reassigned then we need to rename them in
   * the closure so we can get direct access to the outer variable to continue the
   * iteration with bindings based on each iteration.
   *
   * Reference: https://github.com/babel/babel/issues/1078
   */

  BlockScoping.prototype.addContinuations = function addContinuations(fn) {
    var state = {
      reassignments: {},
      outsideReferences: this.outsideLetReferences
    };

    this.scope.traverse(fn, continuationVisitor, state);

    for (var i = 0; i < fn.params.length; i++) {
      var param = fn.params[i];
      if (!state.reassignments[param.name]) continue;

      var newParam = this.scope.generateUidIdentifier(param.name);
      fn.params[i] = newParam;

      this.scope.rename(param.name, newParam.name, fn);

      // assign outer reference as it's been modified internally and needs to be retained
      fn.body.body.push(t.expressionStatement(t.assignmentExpression("=", param, newParam)));
    }
  };

  BlockScoping.prototype.getLetReferences = function getLetReferences() {
    var block = this.block;

    var declarators = [];

    if (this.loop) {
      var init = this.loop.left || this.loop.init;
      if (isBlockScoped(init)) {
        declarators.push(init);
        (0, _extend2.default)(this.outsideLetReferences, t.getBindingIdentifiers(init));
      }
    }

    //
    if (block.body) {
      for (var i = 0; i < block.body.length; i++) {
        var declar = block.body[i];
        if (t.isClassDeclaration(declar) || t.isFunctionDeclaration(declar) || isBlockScoped(declar)) {
          var declarPath = this.blockPath.get("body")[i];
          if (isBlockScoped(declar)) {
            convertBlockScopedToVar(declarPath, null, block, this.scope);
          }
          declarators = declarators.concat(declar.declarations || declar);
        }
      }
    }

    if (block.cases) {
      for (var _i = 0; _i < block.cases.length; _i++) {
        var consequents = block.cases[_i].consequent;

        for (var j = 0; j < consequents.length; j++) {
          var _declar = consequents[j];
          if (t.isClassDeclaration(_declar) || t.isFunctionDeclaration(_declar) || isBlockScoped(_declar)) {
            var _declarPath = this.blockPath.get("cases")[_i];
            if (isBlockScoped(_declar)) {
              convertBlockScopedToVar(_declarPath, _declar, block, this.scope);
            }
            declarators = declarators.concat(_declar.declarations || _declar);
          }
        }
      }
    }

    //
    for (var _i2 = 0; _i2 < declarators.length; _i2++) {
      var _declar2 = declarators[_i2];
      var keys = t.getBindingIdentifiers(_declar2);
      (0, _extend2.default)(this.letReferences, keys);
      this.hasLetReferences = true;
    }

    // no let references so we can just quit
    if (!this.hasLetReferences) return;

    var state = {
      letReferences: this.letReferences,
      closurify: false,
      file: this.file
    };

    // traverse through this block, stopping on functions and checking if they
    // contain any local let references
    this.blockPath.traverse(letReferenceBlockVisitor, state);

    return state.closurify;
  };

  /**
   * If we're inside of a loop then traverse it and check if it has one of
   * the following node types `ReturnStatement`, `BreakStatement`,
   * `ContinueStatement` and replace it with a return value that we can track
   * later on.
   */

  BlockScoping.prototype.checkLoop = function checkLoop() {
    var state = {
      hasBreakContinue: false,
      ignoreLabeless: false,
      inSwitchCase: false,
      innerLabels: [],
      hasReturn: false,
      isLoop: !!this.loop,
      map: {},
      LOOP_IGNORE: (0, _symbol2.default)()
    };

    this.blockPath.traverse(loopLabelVisitor, state);
    this.blockPath.traverse(loopVisitor, state);

    return state;
  };

  /**
   * Hoist all let declarations in this block to before it so they retain scope
   * once we wrap everything in a closure.
   */

  BlockScoping.prototype.hoistVarDeclarations = function hoistVarDeclarations() {
    this.blockPath.traverse(hoistVarDeclarationsVisitor, this);
  };

  /**
   * Turn a `VariableDeclaration` into an array of `AssignmentExpressions` with
   * their declarations hoisted to before the closure wrapper.
   */

  BlockScoping.prototype.pushDeclar = function pushDeclar(node) {
    var declars = [];
    var names = t.getBindingIdentifiers(node);
    for (var name in names) {
      declars.push(t.variableDeclarator(names[name]));
    }

    this.body.push(t.variableDeclaration(node.kind, declars));

    var replace = [];

    for (var i = 0; i < node.declarations.length; i++) {
      var declar = node.declarations[i];
      if (!declar.init) continue;

      var expr = t.assignmentExpression("=", declar.id, declar.init);
      replace.push(t.inherits(expr, declar));
    }

    return replace;
  };

  BlockScoping.prototype.buildHas = function buildHas(ret, call) {
    var body = this.body;

    body.push(t.variableDeclaration("var", [t.variableDeclarator(ret, call)]));

    var retCheck = void 0;
    var has = this.has;
    var cases = [];

    if (has.hasReturn) {
      // typeof ret === "object"
      retCheck = buildRetCheck({
        RETURN: ret
      });
    }

    if (has.hasBreakContinue) {
      for (var key in has.map) {
        cases.push(t.switchCase(t.stringLiteral(key), [has.map[key]]));
      }

      if (has.hasReturn) {
        cases.push(t.switchCase(null, [retCheck]));
      }

      if (cases.length === 1) {
        var single = cases[0];
        body.push(t.ifStatement(t.binaryExpression("===", ret, single.test), single.consequent[0]));
      } else {
        // https://github.com/babel/babel/issues/998
        for (var i = 0; i < cases.length; i++) {
          var caseConsequent = cases[i].consequent[0];
          if (t.isBreakStatement(caseConsequent) && !caseConsequent.label) {
            caseConsequent.label = this.loopLabel = this.loopLabel || this.scope.generateUidIdentifier("loop");
          }
        }

        body.push(t.switchStatement(ret, cases));
      }
    } else {
      if (has.hasReturn) {
        body.push(retCheck);
      }
    }
  };

  return BlockScoping;
}();

module.exports = exports["default"];