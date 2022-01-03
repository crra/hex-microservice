# About

Showcase and very simple example project to explore the application of a (monolithic) Ports and Adapter Pattern / Hexagonal Architecture. It uses golang (1.18 with generics) for the implementation.

Tries to give an answer if in an example code one can find: "error handling omitted for simplicity".

The service represents a very simple URL shortener service. Offers basic CRD (create, read, delete) operations via REST.

Based on the series from https://github.com/tensor-programming/hex-microservice.git and recommendations from https://github.com/katzien/go-structure-examples.

# Disclaimer

The implementation in this repository is somehow over-engineered and not to be considered as reference (yet).

For example technology/domain boundaries convert data too often. The intention behind the multiple conversions are to mimic golang's use of interfaces (Interface Segregation Principle, https://dave.cheney.net/2016/08/20/solid-go-design). So a services consumes are 'storage view' with own entity DTOs rather than using a global domain or storage entity (ORM).

# Todo and Ideas

- implement and test other backends than `memory`
- implement and test other routers than `chi`
- implement the code generator that creates the conversion code that performs the conversion without runtime inspection (reflection)
- compare this custom golang lib version (this) with an existing framework like spring boot (e.g. input validation)
- handle key collisions
- dockerize (also for macOS)
- docker-compose with different storage backends
- custom short ids
- time to live (ttl)
- top10 (update on read)
- internal event sourcing to simulate Command and Query Responsibility Segregation (CQRS)?
