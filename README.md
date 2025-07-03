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

Check out the [docker-compose.yml](/docker-compose.yml)! Note that all of our services have multiple replicas and the requests are still able to be routed correctly.

Happy path:

```sh
docker compose up --build
```

Access `http://localhost:15672/`, login with user: `guest` as password: `guest`. In `task_queue` queue, publish the example task:

```json
{
  "action": "test",
  "transaction_id": "1234567890",
  "product_id": "1234567890",
  "ts": "2025-06-25T18:30:32.56958179Z"
}
```

Example reply:

```sh
docker run -ti --network tcc_default eclipse-mosquitto:1.6.15 ash

mosquitto_pub -h mqtt -t /gw/GW_MAC/response -m {\"deviceMac\":\"000000000001\",\"ack\":true} && mosquitto_pub -h mqtt -t /gw/GW_MAC/response -m {\"deviceMac\":\"000000000002\",\"ack\":true}
```
