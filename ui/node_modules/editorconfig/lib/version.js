function Version(version) {
  var args = arguments;
  this.components = typeof version === "string" ?
    version.split(".").map(function(x){return parseInt(x, 10);}) :
    Object.keys(arguments).map(function(k){return args[k];});

  var len = this.components.length;
  this.major = len ? this.components[0] : 0;
  this.minor = len > 1 ? this.components[1] : 0;
  this.build = len > 2 ? this.components[2] : 0;
  this.revision = len > 3 ? this.components[3] : 0;

  if (typeof version !== "string") {
    return;
  }

  var ext = version.split("-");
  if (ext.length === 2) {
    this.configuration = ext[1];
  }
}

Version.prototype = {
  toString: function() {
    var version = this.components.join(".");
    if (typeof this.configuration !== "undefined") {
      version += "-" + this.configuration;
    }
    return version;
  },
  gte: function(other){
    if (this.major < other.major) {
      return false;
    }
    if (this.minor < other.minor) {
      return false;
    }
    if (this.build < other.build) {
      return false;
    }
    if (this.revision < other.revision) {
      return false;
    }
    return true;
  }
};

module.exports = Version;
