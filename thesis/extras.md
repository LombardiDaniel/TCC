## HOMOLOGATION

After the MVP development, tests where developed to guarantee the working and efficiency of the system. Firstly, we need to set the baseline system performance numbers. All tests were ran on a 12-core machine with 16GB of ram, with limiting of resources only in the router elements in the architecture, all other systems (i.e. queues, dbs etc) had no resource limits. The CPU of the router elements was capped at 100 mCPU per replica. All tests were being executed by 50 workers simultaniously to increase the load on the routers, bringing it to a more realistic value. The request timeout was considered 60 seconds, so if a reply was not recieved within this timeout, the reply window was closed and the request deemed unsuccessful.

The values of the tests were measured using a telemetry service that saves the latency of the RPCs to MongoDB asynchrosnesly in a different coroutine, as to reduce the impact of the measurements. The MongoDB instance is running on a separate server to not interfere with the experiments.

As to capture the complete system latency and execution time, the only collected data was on the worker-side as shown.

```go
// task_worker with telemetry
func main() {
    for task := range rabbitMqMsgsChan {
        esls := dbService.BusinessLogic(tasks)

        var msgs []models.RoutingMessage
        for esl := range esls {
            m := models.RoutingMessage{}.FromEsl(esl)
            msgs = append(msgs, m)
        }

        t0 := time.Now()
        reps := commsBackbone.Forward(msgs) // This method waits for replies
        if len(reps) != len(macs) {
            return errors.New("not all devices replied")
        }
        delta := time.Since(t0)
        asyncTelemetryService.AddDelta(delta)
        // ...
    }
}
```

> Telemetry implemented on task_worker pseudo-code

### Functional Tests

To guarantee and homologate the workings of our system, we start from the baseline; firstly it should be able to simply follow the sequence diagram of the architecture. As the purpose of the first test is to test if the system functions as expected, we executed the test for 1k (one thousand) tasks. It was executed with only 1 router replica. After running the experiment, we plotted the histogram of the latency of each request.

![histogram_baseline_1k_1_router_replicas](/thesis/static/experiments/histogram_baseline_1k_1_router_replicas.png)

We can see a large grouping around the 50ms mark, that originated from the bulk of the requests with concurrent code being executed, also a small peak of less than 10ms that may have originated from the first and last few requests, where code was mostly not concurrent.

With the results from the tests being mostly under 1 second, far below the 60 seconds timeout, we can guarantee that all requests reached the ESL emulator and got replied successfully, independant of ACK or NACK status.

Moving on to our developed Custom Communcation Backbone, with exatcly the same layout as before, we have:

![histogram_custom_1k_1_router_replicas](/thesis/static/experiments/histogram_custom_1k_1_router_replicas.png)

As we can see, the router also worked exactly as expected, with a small increase in mean delay, as well as a increase in p99 delay, but still concluded with a maximum of around 250ms, well below the timeout of 60 seconds. As such, we successfully validade our proposed custom architecture for a single replica.

This establish the base functionality of our systems, both the custom backbone and our baseline backbone are working as expected under base circustances, 50 task workers executing in parallel with a single router replica routing the requests to the device gateways.

### Performance Tests

For the performance tests of the systems, we'll increase the tasks count to 10k (ten thousand) to increase the bulk of concurrent operations, making sure the routes requests overlap much more than in the functional test, stressing the router node. We will also increase the number of routers replicas in each experiment to test horizontal scalability.

Our first performance test will run with 2 replicas for 1k requests in the baseline backbone.

![histogram_baseline_1k_2_router_replicas](/thesis/static/experiments/histogram_baseline_1k_2_router_replicas.png)

Here we already trip into the main pitfall, the misrouting issue with the baseline system is clearly seen by the large amount of requests escaping the 60 seconds timeout window. This means that all requests in the righ-most part of the graph have timedout, being unable to correctly reply to the worker. The few left-most values have randomly succeeded from being routed to the same node that made the request. Note that with only 2 replicas the ratio of misrouted messages to correctly routed ones already greatly surpases the acceptable limit of eventual failures, a ratio that will only increase with the number of replicas of the router.

As such, tests with increased task counts and router replicas will NOT be done on the baseline router case.

With our custom backbone, using the same parameters for test.

![histogram_custom_1k_2_router_replicas](/thesis/static/experiments/histogram_custom_1k_2_router_replicas.png)

Here we continue with the correct routing of messages, wich is clearly showed by still no requests having timed out. Most importantly, we see a small increase in the median when compared to a single router replica.

We can then run a heavier experiment, with 10 thousand requests and keeping the same 2 replicas:

![histogram_custom_10k_2_router_replicas](/thesis/static/experiments/histogram_custom_10k_2_router_replicas.png)

Here we see a great increase in the mean execution time, meaning our system is starting to bottleneck. As such, we also increase the amount of router replicas to 10 and later 20, to allow better load bancing of the routing and compare the results.

![histogram_custom_10k_20_router_replicas](/thesis/static/experiments/histogram_custom_10k_20_router_replicas.png)

Now we see much better metrics, with both the mean and p99 brought down significantly, the mean from 2000ms down to around 120ms.

When measuring the custom backbone with 20 router replicas, we reach a system with a minimal increase in latency and show a true increase in scalability and a proven elasticity of the system.

As a simple "for fun" experiment, we'll also run one with 200 router replicas, to compare how the optimum, "maximum" scaled case for the routers would be. For this, we reserved a machine 64vCPU, 256GB RAM on MagaluCloud.

![histogram_custom_10k_200_router_replicas](/thesis/static/experiments/histogram_custom_10k_200_router_replicas.png)

Here we can see a great decrease in latency, down to around 6ms and a p99 of 57ms. This means that increasing the number of router replicas does lead to a greater scalability. This is because the messages will be waiting on the queue in RabbitMQ for less time before being collected by the routers and forwarded to MQTT.

However, the ideal configuration parameters for scalability using a platform such as KEDA for Kubernetes will not be developed in this study.
