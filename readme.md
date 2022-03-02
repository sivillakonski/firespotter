# Firespotter
A helper tool to run the Plow benchmarking tool from multiple hosts.

NOTE: Currently, it is the Plow benchmarking tool supported (https://github.com/six-ddc/plow).

## Enablement

1. Make sure you have the `plow` utility installed on the node. Please, see the installation instructions from the [Plow repo](https://github.com/six-ddc/plow).
2. Compile the `Firespotter Agent` agent by running the `go build -a -o firespotter` from the root of this repo. This one will read instructions from the `Firespotter Commander` app and run multiple instances of `plow`.
3. Compile the `Firespotter Commander` app -> `cd mock && go build -a -o firespotter-commander`.
4. Run the `Firespotter Commander` app first (`./firespotter-commander`) it will serve the `conf.json` file by default with the target hosts (HTTP, PORT 8080).
5. Run the `Firespotter Agent` app on each node with the flag `--commander-address=ADDRESS_OF_COMMANDER_APP` (`http://127.0.0.1:8080` by default).

Remember, this is the benchmarking tool only, do not use it to abuse any legal services!