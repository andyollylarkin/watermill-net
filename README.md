# Watermill Net Pub/Sub
<img align="right" width="200" src="https://threedots.tech/watermill-io/watermill-logo.png">

This is net **(tcp, udp, unix socket)** Pub/Sub for the [Watermill](https://watermill.io/) project.

All Pub/Sub implementations can be found at [https://watermill.io/pubsubs/](https://watermill.io/pubsubs/).

Watermill is a Go library for working efficiently with message streams. It is intended
for building event driven applications, enabling event sourcing, RPC over messages,
sagas and basically whatever else comes to your mind. You can use conventional pub/sub
implementations like Kafka or RabbitMQ, but also HTTP or MySQL binlog if that fits your use case.

Documentation: https://watermill.io/

Getting started guide: https://watermill.io/docs/getting-started/

Issues: https://github.com/ThreeDotsLabs/watermill/issues

## Contributing

All contributions are very much welcome. If you'd like to help with Watermill development,
please see [open issues](https://github.com/ThreeDotsLabs/watermill/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+)
and submit your pull request via GitHub.

## Implement new transport

Currently implemented tcp4 transport. For implementing additional transport see _Connection_ interface and _pkg/connection/tcp4Connection_ as example.

## Support

If you didn't find the answer to your question in [the documentation](https://watermill.io/), feel free to ask us directly!

Please join us on the `#watermill` channel on the [Gophers slack](https://gophers.slack.com/): You can get an invite [here](https://gophersinvite.herokuapp.com/).

## License

[MIT License](./LICENSE)

## TODO
- [x] Transport wrapper for reconnection
- [x] CI pipeline
- [ ] More tests for subscriber and publisher