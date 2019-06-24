![smarta](https://user-images.githubusercontent.com/8289478/57379460-f873e280-7174-11e9-9c32-b737bc49650c.png)
<img src="https://user-images.githubusercontent.com/8289478/56633099-d6357d00-662a-11e9-9592-0c58dab8ca55.png" width="300" height="72" />

`scrapedumper` is a data-dump tool that is currently coupled to a `MARTA` client.

It's primary purpose is to consume `MARTA` realtime data and upload it to various providers.

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

## Running

After running `go build` to obtain the binary, you can run the binary as long as you provide the required environment variables:
```
type options struct {
	OutputLocation    string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	S3BucketName      string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	MartaAPIKey       string `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	PollTimeInSeconds int    `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`
}
```

`./scrapedumper --output-location=. --marta-api-key={{key}} --poll-time-in-seconds=15`
