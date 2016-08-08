"use strict";

// Dependencies
var Fs = require("fs");

/**
 * IsThere
 * Checks if a file or directory exists on given path.
 *
 * @name IsThere
 * @function
 * @param {String} path The path to the file or directory.
 * @param {Function} callback The callback function called with a boolean value
 * representing if the file or directory exists. If this parameter is not a
 * function, the function will run the synchronously and return the value.
 * @return {IsThere|Boolean} The `IsThere` function if the `callback` parameter
 * was provided, otherwise a boolean value indicating if the file/directory
 * exists or not.
 */
function IsThere(path, callback) {

    // Async
    if (typeof callback === "function") {
        Fs.stat(path, function (err) {
            callback(!err);
        });
        return IsThere;
    }

    // Sync
    try {
        Fs.statSync(path);
        return true;
    } catch (err) {
        return false;
    };
}

module.exports = IsThere;