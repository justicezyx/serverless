# Serverless

README.md files are placed in each of the sub-directories for navigation.
Following are summaries of the demo according to the listed evaluation criteria.

## Code Quality: Is your code clean, well-organized, and documented?

Under this directory, there are `client`, `runtime`, and `dispatcher`
sub-directories.

Client has the code to implement calling the serverless function endpoint
provided by the dispatcher.

Runtime has the code to implement actual functions, and code related to building
actual functions into container image.

Dispatcher has the code to start a http endpoint for client to invoke, and
dispatch invocations to launched container instances defined in runtime.


## Scalability: What strategies can you think for the dispatcher to decide if a container can be kept alive or brought down?

Utilization ratio is calculated to determine when to launch and/or shutdown a
container instance.

Utilization ratio is calculated as the ratio between actual time the container
instance is serving requests to the total time since the container is ready.

## Authentication and Authorization: What strategies and you think of if you have a subset of users that can only invoke RuntimeAlpha but not RuntimeBeta ?

A permission manager is created to statically define what users can call which
function.

## Robustness: How would you make this system robust to single-component failures?

It's difficult to implement fault-tolerance in this demo setup.

In general, reduces the statefulness of component, and persist states into
fault-tolerant distributed storage system is necessary. The component can then
tolerate faults by re-construct its states and join the system by recovering its
state from distributed storage system.

# Economics: How would you track usage and billing for each user account? Who pays for container startup times?

API usage tracker is added. It counts the instances running time.

Typical serverless offering from public clouds are charging users based on the
running time of the underlying VM instances, so the tracker only needs to record
the launch time. Then by `time.Now().Sub(launchTime)` gives the chargable usage.

In this mode, there is no need to distinguish startup time, as it's included as
part of instance running time.

A more serverless type of billing mechanism is to count how many APIs are
invoked. This essentially becomes an SaaS offering, and is not same as
serverless functions.

In that case, the container startup time should be included as part of
the `bursty charging`, which is similar to Uber's rush hour pricing. In this
mode, when users incurs bursty requests, and later scaled back, the short-lived
instances' startup time should be charged to the users according to the running
time.

Lastly, usage and billing are straightforward. The crux is in business model,
i.e., how this service makes profit. Usage and billing need to be designed to
match the business model. Overall, the typical serverless billing model is
certainly more predictable, but offers less margin; the SaaS billing model
requires complicated business modeling to set the right pricing, but offers
vastly more margin.
