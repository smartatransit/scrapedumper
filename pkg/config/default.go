package config

//DefaultWorkConfig is the default collection of dumpers
var DefaultWorkConfig = WorkConfig{
	TrainDumper: DumpConfig{
		Kind: RoundRobinKind,
		Components: []DumpConfig{
			DumpConfig{Kind: S3DumperKind},
			DumpConfig{Kind: FileDumperKind},
			DumpConfig{Kind: DynamoDBDumperKind},
		},
	},
	BusDumper: DumpConfig{
		Kind: RoundRobinKind,
		Components: []DumpConfig{
			DumpConfig{Kind: S3DumperKind},
			DumpConfig{Kind: FileDumperKind},
		},
	},
}
