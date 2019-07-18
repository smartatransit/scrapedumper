package config

//GlobalConfig describes the configuration that can be optionally loaded from
//command-line flags or environment variables, rather than from the yml
//file that specifies which dumpers to use.
type GlobalConfig struct {
	OutputLocation    *string `long:"output-location" env:"OUTPUT_LOCATION" description:"local path to output"`
	DynamoTableName   *string `long:"dynamo-table-name" env:"DYNAMO_TABLE_NAME" description:"dynamo table name"`
	S3BucketName      *string `long:"s3-bucket-name" env:"S3_BUCKET_NAME" description:"s3 bucket to dump stuff into"`
	MartaAPIKey       string  `long:"marta-api-key" env:"MARTA_API_KEY" description:"marta api key" required:"true"`
	PollTimeInSeconds int     `long:"poll-time-in-seconds" env:"POLL_TIME_IN_SECONDS" description:"time to poll marta api every second" required:"true"`
}
