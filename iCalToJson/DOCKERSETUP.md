# Docker Setup

To build the docker image and move it over:

Where you have cloned this repository:
1. `docker build -t go-ical-parser .`
2. `docker save go-ical-parser:latest -o go-ical-parser.tar`
3. `scp go-ical-parser.tar your-user@your-hosted-ip:/where/you/want/it/to/go`

On Receiver:
1. `cd` into wherever you sent the file to.
2. `docker load -i go-ical-parser.tar`.