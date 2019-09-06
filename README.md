![smarta](https://user-images.githubusercontent.com/8289478/57379460-f873e280-7174-11e9-9c32-b737bc49650c.png)
<img src="https://user-images.githubusercontent.com/8289478/56633099-d6357d00-662a-11e9-9592-0c58dab8ca55.png" width="300" height="72" />

`scrapedumper` is a data-dump tool that is currently coupled to a `MARTA` client.

It's primary purpose is to consume `MARTA` realtime data and upload it to various providers.

## Continuous Integration Status

[![Continuous Integration status](https://travis-ci.org/smartatransit/scrapedumper.svg?branch=master)](https://travis-ci.org/smartatransit/scrapedumper.svg?branch=master)
[![codecov](https://codecov.io/gh/smartatransit/scrapedumper/branch/master/graph/badge.svg)](https://codecov.io/gh/smartatransit/scrapedumper)

## Why is this needed?
This allows people to build a historical dataset of `MARTA` arrival times.  `SMARTA` plans to use this to provide a dataset for `MARTA` train forecasting.

## How does it work?
The `dumper.Dumper` interface provides a `Dump` function that accepts a reader (and presumably dumps data somewhere).

Implementing this interface should allow an extensible way to `Dump` data wherever it is needed.

## Project Goals
- [X] Allow upload to local directories
- [X] Allow upload to `S3`
- [X] Allow upload to `Dynamo`
- [X] Allow multiclient response handling for `Dynamo` handler
- [ ] Use a `Scraper` interface instead of a coupling marta client to it
- [X] `circuitbreaker` in the worker?
- [X] backoff, jitter, retryer on marta client

## Running

After running `go build` to obtain the binary, you can run the binary as long as you provide the required environment variables:
```golang
type options struct {
	OutputLocation    string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName   string `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName      string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	MartaAPIKey       string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	PollTimeInSeconds int    `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`

	ConfigPath *string `long:"config-path" env:"CONFIG_PATH" description:"An optional file that overrides the default configuration of sources and targets."`
}
```

`./scrapedumper --output-location=. --marta-api-key={{key}} --poll-time-in-seconds=15`

### Config Based Approach
```json
{
	"bus_dumper": {
		"kind": "ROUND_ROBIN | FILE | S3 | DYNAMODB",
		"components": [],
		"options" {
			"s3_bucket_name": "",
			"dynamo_table_name": "",
			"local_output_location": "",
		}
	},
	"train_dumper": {
		"kind": "ROUND_ROBIN | FILE | S3 | DYNAMODB",
		"components": [],
		"options" {
			"s3_bucket_name": "",
			"dynamo_table_name": "",
			"local_output_location": "",
		}
	}
}
```

new approach:
```json
{

  "scrapers": [
	{
		"kind": "WEB,
		"dumper":
			{
				"kind": "ROUND_ROBIN | FILE | S3 | DYNAMODB",
				"components": [],
				"options" {
					"s3_bucket_name": "",
					"dynamo_table_name": "",
					"local_output_location": "",
				}
			},
		"options" {
		}
	},
  ]
}
```

`./scrapedumper --config-path=./config --marta-api-key={{key}} --poll-time-in-seconds=15`
