# Pipeline
Proof of concept of a proxy that allows a client to "connect" to a "server" running on a browser.

So, the idea is that the browser connects to the proxy server, and negotiates a websocket.
Once the websocket connection is established, the proxy server stars listening on a port. A third party can then
connect to this port, and the traffic is tunneled between the browser js app and the remote client.
