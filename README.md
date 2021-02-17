# LBLight
Light/Simple load balancer

## Summary

LBLight is a simple Layer 7 load balancer, without a lot of bells and whistles. The primary aim is to be able to route traffic based on either the beginning of the HTTP path or based on HTTP headers. It can be configured so (for example) any queries starting with "/foo" will get directed to backends X or Y... whereas any queries starting with "/bar" will be directed to backend Z. Similarly queries with a HTTP header of "X-Node" of value "A" might get redirected to backend XX and those with a value of "B" could go the backend YY. 

A primary aim of this is to be able to run as an Azure App Service. The default external Azure load balancers cannot route based on path nor header. Other Azure load balancing options (Application Gateway) do exactly what we want, but appear to get a bit pricey once we hit a certain limit. This is an attempt to give basic load balancing/routing while being hosted in App Services. This can of course be hosted elsewhere, but App Services is the current target/purpose.

A few concepts are important to know before attempting to configure. There are 3 key structures/terms within LBLight.

**BackendRouter** : This is configured to handle a set of paths (/foo, /bar, /whatever) or a set of HTTP headers (key/value pairs). Each BackendRouter has a number of Backends associated with it.

**Backend**: Is a structure tied to a specific destination host (whether a VM, cluster, another LB etc). This will be used by the BackendRouter to deliver the right traffic to the right target host. eg. The BackendRouter might be configured to know any traffic starting with "/foo" needs to go to IPs 10.0.0.1 and 10.0.0.2. For this The BackendRouter will have two Backends configured, one for each IP address. Each connection between a Backend and the destination host will have a BackendConnection instance stored within the Backend.

**BackendConnection**: Retains the actual connection between LBLight and the destination/target host. In reality this is a very small struct that primarily has a Go httputil.ReverseProxy instance which shuffles traffic between the LBLight client and the destination host. Each one of these will be running in it's own go routine.

## Building

no fancy makefiles etc... a simple "go build ." will do the trick.

## Config

Configuration of LBLight is through the lblight.json file. The format I hope is self explanatory, but if not, the key parts are:

- There is a list of BackendRouterConfigs. 
- Each BackendRouterConfig has:
  - A list of AcceptedPaths (eg. /foo, /bar etc).
  - A list of AcceptedHeaders (key/value pairs for HTTP headers)
  - A list of BackendConfigs. Each of which contains the host, port and maximum number of connections allowed for each destination host.

Given LBLight needs to serve encrypted (HTTPS, WSS) traffic it will require certificates. Currently development is purely using self signed signatures (created with OpenSSL). Am not providing certificates, but are easy enough to create (google it :) )

## Running

Running locally, simple run the command with lblight.json in the same directory.

Running in App Service....  still working on that :)



## TODO

(should add these as Github issues)

- Round robin backend selection
- Least # active connections backend selection
- Health check for backend
- Azure App Service running (HTTP and HTTPS)
- Web sockets via Azure App Service
- Prometheus endpoint for metrics
- Keep connections to destination host open (allow pooling)
- Prove can handle 1000 parallel web sockets
- Improve logging



