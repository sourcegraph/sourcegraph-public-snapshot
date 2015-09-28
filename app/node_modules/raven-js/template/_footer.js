// Expose Raven to the world
if (typeof define === 'function' && define.amd) {
    // AMD
    window.Raven = Raven;
    define('raven', [], function() {
      return Raven;
    });
} else if (typeof module === 'object') {
    // browserify
    module.exports = Raven;
} else if (typeof exports === 'object') {
    // CommonJS
    exports = Raven;
} else {
    // Everything else
    window.Raven = Raven;
}

})(typeof window !== 'undefined' ? window : this);
