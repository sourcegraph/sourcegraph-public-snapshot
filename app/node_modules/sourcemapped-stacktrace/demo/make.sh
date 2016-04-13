# (don't forget to run npm install first, to update dist/smst.js)
babel bork.es6 -s true > bork.babel.js
coffee -c -m bork_coffee.coffee
cp ../dist/sourcemapped-stacktrace.js .
