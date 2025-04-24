# Optimizing IoT Cloud Applications for Scalability: Leveraging RPC with Distributed Queues for Seamless Operations

TCC

---

![architecture](/static/arch.png)

Router enables worker to have purelly sync code while still allowing for horizontal scalability in all sides.

- Routing Queue acts as both RPC framework and service and load balancer.
- Shared Memory acts as service discovery, storage for return addrs of workers (via ephemeral in-memroy queues in RabbitMQ).

![rpc](/static/rpc.png)

This creates a decoupled system that does not need a long-lived connection between the router and the worker, allowing for better scalability.

Different levels of Async and Sync communication are common pitfals for scalability, where a single service may be needed to avoid resource locks and such. With this architecture, the TaskWorker (where business logic reside) is purelly synchronous. This allows for more reliable development and testing.

In a regular architecture, the Router would not be able to scale horizontally, because a reply from MQTT (round-robin) may fall on a different node, one that did not make the request, making it unable to ack the request (made by the worker).
