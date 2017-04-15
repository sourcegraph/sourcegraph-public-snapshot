# Copies the sgdev-sourcegraph.js script to the plugin dir and restarts
cp sgdev-sourcegraph.js /var/www/phabricator/webroot/rsrc/js/sourcegraph/sgdev-sourcegraph.js && sh ~/restart_phabricator.sh
