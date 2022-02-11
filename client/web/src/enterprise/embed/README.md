# Developing the EmbeddedWebApp

- To enable the development of the `embed` bundle, you'll have to set the `EMBED_DEVELOPMENT=true` env variable
- To test your local changes on 3rd party sites, you will have to proxy your dev environment using [ngrok](https://ngrok.com/)
- Run ngrok using this command: `./ngrok http 3080 --host-header=rewrite`
- Copy the Forwarding http:// address and use it to replace the `externalURL` property in your `site-config.json`
- Start (or restart) your local dev environment
- Navigate to a 3rd party site (e.g., codepen.io), create a new page, and embed an iframe using the ngrok URL
  - Example iframe tag: `<iframe src="http://your-dev-env.ngrok.io/embed/notebooks/123" frameborder="0" sandbox="allow-scripts allow-same-origin"></iframe>`
