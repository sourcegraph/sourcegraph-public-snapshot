+++
title = "Public server"
+++

By default, only specific users (chosen by the administrator) may
access a Sourcegraph server. To allow anyone with a Sourcegraph account
to access your server (for open-source projects, for example), you should
run the following commands from a terminal:

1. Get your Sourcegraph server's config:
	
		src --endpoint http://example.com meta config

2. Copy the IDKey string from the JSON response of the above command.

3. Update your server's access configuration on sourcegraph.com:
	
		src --endpoint http://sourcegraph.com registered-clients update --allow-logins=all <IDKey>

4. Your server is now set up for anyone to access. To revert back to restricted logins, run:
	
		src --endpoint http://sourcegraph.com registered-clients update --allow-logins=restricted <IDKey>
