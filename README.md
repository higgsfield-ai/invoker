# invoker

**Description:**
`invoker` is a powerful command-line utility written in Golang that facilitates running and managing experiments seamlessly. It offers features like running experiments on multiple hosts, specifying container names, and more.

## Installation:

1. **Download Binary:**
   - Visit the [Releases](https://github.com/higgsfield-ai/invoker/releases) section of the GitHub repository.
   - It only supports Linux for now, since we expect servers to be in Linux.
   - Extract the downloaded archive.

2. **Compile from Source:**
   - Ensure you have [Golang](https://golang.org/doc/install) installed.
   - Clone the repository: `git clone https://github.com/higgsfield-ai/invoker.git`
   - Navigate to the project directory: `cd invoker`
   - Build the binary: `go build -o invoker`

3. **Add to PATH:**
   - For easy access, move the binary to a directory included in your `PATH`. For example, on Unix systems:
     ```bash
     mv invoker /usr/local/bin/
     ```

## Usage:

### Basic Commands:

- **Generate a random name:**
  ```bash
  invoker random-name
  ```

- **Generate a random port:**
  ```bash
  invoker random-port
  ```

### Experiment Commands:

- **Run an experiment:**
  ```bash
  invoker experiment run --experiment_name=<experiment_name> --project_name=<project_name> --hosts=<host1,host2,...> [--container_name=<container_name>] [--nproc_per_node=<num_processes>] [--port=<port_number>] [--run_name=<run_name>]
  ```

- **Kill an experiment:**
  ```bash
  invoker experiment kill --experiment_name=<experiment_name> --project_name=<project_name> --hosts=<host1,host2,...> [--container_name=<container_name>]
  ```

### Additional Commands:

- **Decode Secrets:**
  ```bash
  invoker decode-secrets
  ```

- **Generate Autocompletion Script:**
  ```bash
  invoker completion
  ```

### Examples:

- **Run an experiment:**
  ```bash
  invoker experiment run --experiment_name=my_experiment --project_name=my_project --hosts=host1,host2,host3 --container_name=my_container --nproc_per_node=2 --port=5678 --run_name=first_run
  ```

- **Kill an experiment:**
  ```bash
  invoker experiment kill --experiment_name=my_experiment --project_name=my_project --hosts=host1,host2,host3 --container_name=my_container
  ```

## Help:

For more details on each command and its flags, use the `--help` option. For example:
```bash
invoker --help
invoker experiment run --help
invoker experiment kill --help
```
If you encounter any issues or have suggestions, please check the [GitHub Issues](https://github.com/higgsfield-ai/invoker/issues) page.

**Stay fine-tuned!**
