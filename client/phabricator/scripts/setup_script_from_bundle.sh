# concatenates base.js to the bundle.js in the home directory, and stores them in the correct plugin directory
cat ~/base.js ~/sgdev.bundle.js > /var/www/phabricator/webroot/rsrc/js/sourcegraph/phabricator-sourcegraph.js && sh ~/restart_phabricator.sh
