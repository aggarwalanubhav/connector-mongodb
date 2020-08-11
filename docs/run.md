# Run

> Back-up is uploaded/restored by streaming to/from the Storj network.

The following flags can be used with the `store` command:

* `accesskey` - Connects to the Storj network using a serialized access key instead of an API key, satellite url and encryption passphrase.
* `share` - Generates a restricted shareable serialized access with the restrictions specified in the Storj configuration file.

The following flags  can be used with the `restore` command:

* `accesskey` - Connects to the Storj network using a serialized access key instead of an API key, satellite url and encryption passphrase.
* `match` - Matches to regular expression with the databases whose back-up(s) are uplaoded to Storj network and restores the latest back-up of all the matching databases.
* `latest` - Restores the latest back-up of the specified MongoDB database.

Once you have built the project you can run the following:

## Get help

```
$ ./connector-mongodb --help
```

## Check version

```
$ ./connector-mongodb --version
```

## Upload back-up data to Storj

```
$ ./connector-mongodb store --local <path_to_mongodb_config_file> --storj <path_to_storj_config_file>
```

## Upload back-up data to Storj bucket using Access Key

```
$ ./connector-mongodb store --accesskey
```

## Upload back-up data to Storj and generate a Shareable Access Key based on restrictions in `storj_config.json`

```
$ ./connector-mongodb store --share
```

## Restore the lastest back-up of the specified database

```
$ ./connector-mongodb restore --latest <database_name>
```

## Match all the database on the Storj network with the provided regular expression and restore the lastest back-up of the all the matching databases

```
$ ./connector-mongodb restore --match <regex>
```
