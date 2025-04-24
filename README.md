# Optimizing IoT Cloud Applications for Scalability: Leveraging RPC with Distributed Queues for Seamless Operations

TCC

---

![architecture](/static/arch.png)

Router enables worker to have purelly sync code while still allowing for horizontal scalability in all sides.

- Routing Queue acts as both RPC framework and service and load balancer.
- Shared Memory acts as service discovery, storage for return addrs of workers (via ephemeral in-memroy queues in RabbitMQ).

![rpc](/static/rpc.png)

This creates a decoupled system that does not need a long-lived connection between the router and the worker, allowing for better scalability.
