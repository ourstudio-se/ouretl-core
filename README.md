# ouretl-core - a plugin based ETL orchestrator

*ouretl-core* preferably reads a configuration file to include plugins into a working set, meaning a chain of DataHandlerPlugins and with at least one WorkerPlugin as source (to push data on to the chain of handlers). Any DataHandlerPlugin can be used as a sink for the ETL, as well as a data transformer.

Configuration file example:

    inherit_settings_from_env = true

    [[plugin]]
    name = "ouretl-plugin-stdin-reader"
    path = "/tmp/ouretl-plugins/stdin-reader.so.1.0.0"
    version = "1.0.0"
    priority = 1

    [[plugin]]
    name = "ouretl-plugin-data-transform"
    path = "/tmp/ouretl-plugins/data-transform.so.1.0.0"
    version = "1.0.0"
    priority = 10
    settings_file = "/tmp/plugin-settings/data-tranform.toml"

    [[plugin]]
    name = "ouretl-plugin-stdout-writer"
    path = "/tmp/ouretl-plugins/stdout-writer.so.1.0.0"
    version = "1.0.0"
    priority = 20

To run *ouretl-core* using this configuration, simply pass it as a parameter using `ouretl-core -config=/any/path/ouretl-config.conf` or use the default file path `/etc/ouretl/default.conf`.

## Development

To run *ouretl-core* in dev mode, you can pass in environment variables before the `go run` command;

    MY_SETTING="myValue" ./run-local.sh -config=/any/path/ouretl-config.conf