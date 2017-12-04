#!/bin/bash

set -e

echo "You must restart your Phabricator instance, please add appropriate commands to this file."
echo "See https://secure.phabricator.com/book/phabricator/article/restarting/ for more information."

# examples:
#
# sudo service apache2 restart
#
# supervisorctl -c /app/supervisord.conf restart php-fpm
# supervisorctl -c /app/supervisord.conf restart nginx
# supervisorctl -c /app/supervisord.conf restart nginx-ssl