# How to use a static IP for updates and pings

In some rare occurences, it may be required to use a single static IP for a Sourcegraph application to be able to perform update checks and ping back to our licensing API. 

## Forcing the URL through the environment

Setting the `UPDATE_CHECK_BASE_URL` to the value provided by the Sourcegraph support on the `frontend` application manifest will ensure that 
all update checks and pings are going to that specific URL.

Please note that only the `https` scheme is allowed. If the provided value is invalid or with any other scheme, the app will default to the default base URL.
