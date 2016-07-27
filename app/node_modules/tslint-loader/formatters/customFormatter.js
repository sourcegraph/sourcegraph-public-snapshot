var Lint = require("tslint/lib/lint");

var __extends = this.__extends || function (d, b) {
  for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
  function __() { this.constructor = d; }
  __.prototype = b.prototype;
  d.prototype = new __();
};

var Formatter = (function (_super) {
  __extends(Formatter, _super);
  function Formatter() {
    _super.apply(this, arguments);
  }
  Formatter.prototype.format = function (failures) {
    var outputLines = failures.map(function (failure) {
      var failureString = failure.getFailure();
      var lineAndCharacter = failure.getStartPosition().getLineAndCharacter();
      var positionTuple = "[" + (lineAndCharacter.line + 1) + ", " + (lineAndCharacter.character + 1) + "]";
      return positionTuple + ": " + failureString;
    });
    return outputLines.join("\n") + "\n";
  };
  return Formatter;
})(Lint.Formatters.AbstractFormatter);
exports.Formatter = Formatter;
