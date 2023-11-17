# Geolocation data

This package contains a small geolocation database that gets embedded in the `internal/requestclient/geolocation` package.
We currently use the "IP2Location LITE" `IP-COUNTRY` (`db1`) database, which only includes country-level data.
IP2Location LITE allows for redistribution with the following acknowledgement on all "sites, advertising materials and documentation mentioning features or use of this database":

> This site or product includes IP2Location LITE data available from http://www.ip2location.com.

For more details, refer to [`IP2LOCATION-LITE-DB1.IPV6.BIN/README_LITE.TXT`](./IP2LOCATION-LITE-DB1.IPV6.BIN/README_LITE.TXT), which is part of database download.

## Updating the database

This database was last updated 11/16/2023 - the [IP2Location FAQ](https://www.ip2location.com/faqs) indicates the database should be updated yearly.

The database must be downloaded manually from the IP2Location website, and requires an account in order to agree to IP2Location terms (such as the attribution required above).
An existing account and detailed instructions are available in [this shared 1Password entry](https://my.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/fonquuhrzrdr4irultcuj6gqby).

The contents of `IP2LOCATION-LITE-DB1.IPV6.BIN` should include everything in the download, including the license and README text.
