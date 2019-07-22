# GMonMQTT

Console based dashboard for monitoring the health of an MQTT Broker.

This should be easily adaptable to a RabbitMQ broker as well.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisities

*  Go (1.6+) (https://golang.org/doc/install)
   The Go Language Compiler and tools.

*  Glide (https://github.com/Masterminds/glide)
   A GoLang dependency management tool

```
# Installing Glide on OSX:
brew install glide
```

### Clone the project
```
    git clone git@git.synapse-wireless.com:terry.bayne/GMonMQTT.git
```

### Building

Download or Update all of the libraries listed in the project's glide.yaml file, putting them in the vendor directory (also recursively walks through package dependencies.

```
    glide up
```

Then invoke the Go compiler to build the program

```
    go build
```

## Deployment

To deploy the executable, simply copy it to the target system (preferably somewhere in that system's defined **path**.)


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

* **Terry Bayne** - *Initial work* - 

## License

TBD

