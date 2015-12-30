# cPanel api in Go

This is a simple cPanel api written in go.

Currently four intefaces are implemented,

- CGI cPanel LiveApi - Designed for use in plugins, this interface will work through the preauthenticated CGI LiveApi environment as documented [here](https://documentation.cpanel.net/display/SDK/Guide+to+the+LiveAPI+System) (UAPI/API2/API1)
- Authenticated JSON cPanel API (UAPI/API2/API1)
- WHM (WHMAPI1)
- WHM Impersonation to call cPanel API (UAPI/API2/API1)

## Example

A simple command line example is provided in the example folder.

## About

This API forms part of the [Let's Encrypt for cPanel](https://letsencrypt-for-cpanel.com/) plugin which allows cPanel/WHM hosters to provide free [Let's Encrypt](https://letsencrypt.org/) certificates for their clients.