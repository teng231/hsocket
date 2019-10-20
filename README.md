## hsocket is a simple app realtime.

* **Basic:** You can using it like a serverless or can use like a micro service.
* **Simple:** Simple flow looklike a service notification.
* **JS support:** Integrated `js` client with support browser or nodejs enviroment.
* **Fast:** You can run like a server very simple and fast.
* **Other:** Can use like a `broker` realtime service.

## Installtion
Using golang
``` sh
go get -v github.com/my0sot1s/hsocket
```

JS client request [wsclient](https://github.com/my0sot1s/hsocket/blob/master/wsClient.js)

``` js
	http://<path>/wsClient.js
```
### Simple flow

```
                          +-----+
                          |client
+------------+   /ws      |     |
|            +----------->+-----+
|     ws     |
|            |   /ws      +-----+
+------+-----+-------<----+client
       ^                  |     |
       |                  +-----+
       |/ws-firer
       |
  +----+----+
  |         |
  |other service
  |         |
  +---------+

```

## Machenics

* Ws is a simple server between client and other realtime server. Ws just received message define `Command`, So server  just listen request `subscribe` or `unsubscribe` from client.
* `other service` is something connect to ws. We don't know.
Ws can listen with  endpoint `\ws-firer` with payload is a `Message` to broadcast to all client listen topic.
* You can see my struct at [core.go](https://github.com/my0sot1s/hsocket/blob/master/core.go)