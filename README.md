# grpc-push-notif
There are primarily 4 types of gRPC apis, req/resp, req/streaming-resp, streaming-req/streaming-resp and all gRPC apis are initiated by the client and gRPC, as of now, doesnt natively support server-initiated rpc calls. However, there are many use cases where a server may have to initiate a request. gRPC allows this mechanism in a bit of a twisted way. I have come across many questions on different forums requesting the same. The effort here is to really showcase that capability of grpc, it is done by having a long-lived streaming api.

The implementation here emulates a very simple, rudimentary smart thermostate which connects to its server, subscribes to a bunch notifications and listens to them. The screenshot below provides the overview, fairly simple.

![Screenshot](screenshot.gif)
